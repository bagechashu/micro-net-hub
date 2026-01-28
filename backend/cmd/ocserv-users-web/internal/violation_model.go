package internal

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// ============================================================================
// SQL 常量集中管理
// ============================================================================

// 数据库初始化相关 SQL
const (
	sqlCreateViolationsTable = `
	CREATE TABLE IF NOT EXISTS violations (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		remote_ip TEXT NOT NULL,
		action TEXT NOT NULL CHECK(action IN ('block', 'logonly')),
		reason TEXT,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	sqlCreateIndexTimestamp = `CREATE INDEX IF NOT EXISTS idx_timestamp ON violations(timestamp DESC)`

	sqlCreateIndexUsername = `CREATE INDEX IF NOT EXISTS idx_username ON violations(username)`

	sqlCreateIndexUsernameTimestamp = `CREATE INDEX IF NOT EXISTS idx_username_timestamp ON violations(username, timestamp DESC)`

	sqlCreateIndexAction = `CREATE INDEX IF NOT EXISTS idx_action ON violations(action)`

	sqlCreateIndexActionTimestamp = `CREATE INDEX IF NOT EXISTS idx_action_timestamp ON violations(action, timestamp DESC)`
)

// 插入操作相关 SQL
const (
	sqlInsertViolation = `INSERT INTO violations (id, username, remote_ip, action, reason, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?)`
)

// 查询操作相关 SQL
const (
	sqlSelectCountWithFilter = `SELECT COUNT(0) FROM violations WHERE %s`

	sqlSelectViolationsWithFilter = `
	SELECT id, username, remote_ip, action, reason, timestamp
	FROM violations
	WHERE %s
	ORDER BY timestamp DESC
	LIMIT ? OFFSET ?`

	sqlSelectActionStats = `
	SELECT action, COUNT(0) as count
	FROM violations
	WHERE %s
	GROUP BY action`

	sqlSelectUserStats = `
	SELECT 
		username,
		COUNT(0) as total,
		SUM(CASE WHEN action = 'block' THEN 1 ELSE 0 END) as block_count,
		SUM(CASE WHEN action = 'logonly' THEN 1 ELSE 0 END) as logonly_count
	FROM violations
	WHERE timestamp >= ?
	GROUP BY username
	ORDER BY total DESC
	LIMIT 10`

	sqlSelectLast24hCount = `SELECT COUNT(0) FROM violations WHERE timestamp >= ?`
)

// 删除和维护操作相关 SQL
const (
	sqlDeleteOldViolations = `DELETE FROM violations WHERE timestamp < ?`

	sqlVacuumDB = `VACUUM`
)

// PRAGMA 设置
var sqlPragmas = []string{
	"PRAGMA synchronous=NORMAL",
	"PRAGMA cache_size=20000",
	"PRAGMA temp_store=MEMORY",
	"PRAGMA foreign_keys=ON",
	"PRAGMA query_only=FALSE",
	"PRAGMA busy_timeout=5000",
}

// ============================================================================
// 业务类型定义
// ============================================================================

// DBManager encapsulates database access with RWMutex for thread-safe operations
type DBManager struct {
	db    *sqlx.DB
	mutex sync.RWMutex
}

// GetDB retrieves the database connection with read lock
// The lock is held until the callback completes
func (dm *DBManager) GetDB(fn func(*sqlx.DB) error) error {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	if dm.db == nil {
		return fmt.Errorf("[violation] database not initialized")
	}
	return fn(dm.db)
}

// SetDB sets the database connection with write lock
func (dm *DBManager) SetDB(db *sqlx.DB) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	dm.db = db
}

// Close closes the database with write lock
func (dm *DBManager) Close() error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	if dm.db != nil {
		return dm.db.Close()
	}
	return nil
}

// ViolationLog represents a single VPN access violation record
type ViolationLog struct {
	ID        string    `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	RemoteIP  string    `db:"remote_ip" json:"remote_ip"`
	Action    string    `db:"action" json:"action"` // "block" or "logonly"
	Reason    string    `db:"reason" json:"reason"` // reason for violation
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
}

// ViolationQuery defines filtering parameters for violation queries
type ViolationQuery struct {
	TimeBack       time.Duration
	StartTime      time.Time
	EndTime        time.Time
	UseCustomRange bool
	Username       string
	Action         string
	Limit          int
	Offset         int
}

const (
	defaultLimit        = 20
	maxLimit            = 1000
	maxTimeBackDays     = 365
	defaultTimeBackDays = 7
	queryTimeout        = 30 * time.Second
)

var (
	dbManager = &DBManager{}
	dbPath    = "data/violations.db"
)

// generateID generates a proper UUID v4 for records
func generateID() string {
	return uuid.New().String()
}

// buildWhereClause builds SQL WHERE clause with dynamic filter conditions
// If endTime is zero value, only uses startTime constraint
// If endTime is set, uses startTime AND endTime constraint
// Returns the WHERE clause string and corresponding query arguments
func buildWhereClause(startTime, endTime time.Time, username, action string) (string, []interface{}) {
	var whereConditions []string
	var args []interface{}

	// Handle timestamp conditions
	if endTime.IsZero() {
		// Only startTime constraint
		whereConditions = append(whereConditions, "timestamp >= ?")
		args = append(args, startTime)
	} else {
		// Custom date range
		whereConditions = append(whereConditions, "timestamp >= ? AND timestamp < ?")
		args = append(args, startTime, endTime)
	}

	if username != "" {
		whereConditions = append(whereConditions, "username = ?")
		args = append(args, username)
	}

	if action != "" {
		whereConditions = append(whereConditions, "action = ?")
		args = append(args, action)
	}

	whereClause := strings.Join(whereConditions, " AND ")
	return whereClause, args
}

// InitializeViolationDB initializes the SQLite violation logging database
func InitializeViolationDB(dataPath string) error {
	if dataPath != "" {
		dbPath = dataPath + "/violations.db"
	}

	var err error
	// Use sqlx with sqlite3 driver
	db, err := sqlx.Open("sqlite3", dbPath+"?cache=shared&mode=rwc&_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("[violation] failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		return fmt.Errorf("[violation] failed to ping database: %w", err)
	}

	// Set connection pool parameters for SQLite with optimized settings
	// SQLite doesn't benefit from large connection pools; keep it minimal
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Set pragmas for performance and safety
	for _, pragma := range sqlPragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("[violation] failed to set pragma: %w", err)
		}
	}

	// Create tables with proper schema
	if _, err := db.Exec(sqlCreateViolationsTable); err != nil {
		return fmt.Errorf("[violation] failed to create table: %w", err)
	}

	// Create indexes
	indexes := []string{
		sqlCreateIndexTimestamp,
		sqlCreateIndexUsername,
		sqlCreateIndexUsernameTimestamp,
		sqlCreateIndexAction,
		sqlCreateIndexActionTimestamp,
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("[violation] failed to create index: %w", err)
		}
	}

	// Set the database in manager
	dbManager.SetDB(db)

	// Initialize async violation buffer for non-blocking writes
	violationBuffer = NewViolationBuffer(db, 1000)
	violationBuffer.Start()

	log.Printf("[violation] SQLite3 database initialized at: %s", dbPath)
	return nil
}

// CloseViolationDB closes the violation database and flushes pending writes
func CloseViolationDB() error {
	// Stop the async buffer first to flush all pending writes
	if violationBuffer != nil {
		if err := violationBuffer.Stop(); err != nil {
			log.Printf("[violation] failed to stop buffer: %v", err)
		}
	}

	return dbManager.Close()
}

// RecordViolation records a VPN access violation asynchronously
// This is now non-blocking and uses batching for better performance
func RecordViolation(username, remoteIP, action, reason string) error {
	if violationBuffer == nil {
		return fmt.Errorf("[violation] buffer not initialized")
	}

	// Validate inputs
	username = strings.TrimSpace(username)
	remoteIP = strings.TrimSpace(remoteIP)
	action = strings.ToLower(strings.TrimSpace(action))

	if username == "" || remoteIP == "" {
		return fmt.Errorf("[violation] invalid username or remoteIP")
	}

	if action != "block" && action != "logonly" {
		return fmt.Errorf("[violation] invalid action: %s", action)
	}

	violation := &ViolationLog{
		ID:        generateID(),
		Username:  username,
		RemoteIP:  remoteIP,
		Action:    action,
		Reason:    reason,
		Timestamp: time.Now(),
	}

	// Send to async buffer (non-blocking with retry)
	return violationBuffer.Write(violation)
}

// GetViolations retrieves violations with pagination and optional filters
// Returns (violations, total_count, error)
func GetViolations(query ViolationQuery) ([]ViolationLog, int64, error) {
	var violations []ViolationLog
	var totalCount int64

	err := dbManager.GetDB(func(db *sqlx.DB) error {
		// Validate and sanitize parameters
		if query.Limit <= 0 {
			query.Limit = defaultLimit
		}
		if query.Limit > maxLimit {
			query.Limit = maxLimit
		}
		if query.Offset < 0 {
			query.Offset = 0
		}

		// Sanitize username and action
		query.Username = strings.TrimSpace(query.Username)
		query.Action = strings.ToLower(strings.TrimSpace(query.Action))
		if query.Action != "" && query.Action != "block" && query.Action != "logonly" {
			query.Action = ""
		}

		// Calculate time range
		var queryStartTime, queryEndTime time.Time
		if query.UseCustomRange {
			queryStartTime = query.StartTime
			queryEndTime = query.EndTime
		} else {
			// Use time-back calculation
			timeBack := query.TimeBack
			if timeBack <= 0 {
				timeBack = time.Duration(defaultTimeBackDays) * 24 * time.Hour
			}
			if timeBack > time.Duration(maxTimeBackDays)*24*time.Hour {
				timeBack = time.Duration(maxTimeBackDays) * 24 * time.Hour
			}
			queryStartTime = time.Now().Add(-timeBack)
			queryEndTime = time.Time{}
		}

		whereClause, args := buildWhereClause(queryStartTime, queryEndTime, query.Username, query.Action)

		// Get total count
		countQuery := fmt.Sprintf(sqlSelectCountWithFilter, whereClause)
		countCtx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		err := db.QueryRowContext(countCtx, countQuery, args...).Scan(&totalCount)
		cancel()
		if err != nil {
			return fmt.Errorf("[violation] count query failed: %w", err)
		}

		// Get paginated results
		dataQuery := fmt.Sprintf(sqlSelectViolationsWithFilter, whereClause)
		dataArgs := append(args, query.Limit, query.Offset)

		dataCtx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()

		err = db.SelectContext(dataCtx, &violations, dataQuery, dataArgs...)
		if err != nil {
			return fmt.Errorf("[violation] query failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return violations, totalCount, nil
}

// GetViolationStats retrieves action-based statistics for violations
// If query.UseCustomRange is true, uses StartTime and EndTime; otherwise uses TimeBack
func GetViolationStats(query ViolationQuery) (map[string]interface{}, error) {
	var stats map[string]interface{}

	err := dbManager.GetDB(func(db *sqlx.DB) error {
		stats = make(map[string]interface{})

		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()

		// Calculate time range
		var queryStartTime, queryEndTime time.Time
		if query.UseCustomRange {
			queryStartTime = query.StartTime
			queryEndTime = query.EndTime
		} else {
			// Use time-back calculation
			timeBack := query.TimeBack
			if timeBack <= 0 {
				timeBack = time.Duration(defaultTimeBackDays) * 24 * time.Hour
			}
			if timeBack > time.Duration(maxTimeBackDays)*24*time.Hour {
				timeBack = time.Duration(maxTimeBackDays) * 24 * time.Hour
			}
			queryStartTime = time.Now().Add(-timeBack)
			queryEndTime = time.Time{}
		}

		// Build query conditions
		whereClause, args := buildWhereClause(queryStartTime, queryEndTime, query.Username, "")

		// Get total count
		var totalCount int
		err := db.QueryRowContext(ctx, fmt.Sprintf(
			sqlSelectCountWithFilter, whereClause,
		), args...).Scan(&totalCount)
		if err != nil {
			return fmt.Errorf("[violation] failed to get total count: %w", err)
		}
		stats["total_violations"] = totalCount

		// Get action breakdown
		actionStats := make(map[string]int)
		actionQuery := fmt.Sprintf(sqlSelectActionStats, whereClause)

		rows, err := db.QueryContext(ctx, actionQuery, args...)
		if err != nil {
			return fmt.Errorf("[violation] failed to get action stats: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var action string
			var count int
			if err := rows.Scan(&action, &count); err != nil {
				log.Printf("[violation] failed to scan action stat: %v", err)
				continue
			}
			actionStats[action] = count
		}
		stats["by_action"] = actionStats

		// Get recent violations (last 24 hours) count
		last24h := time.Now().Add(-24 * time.Hour)
		var last24Count int
		if query.UseCustomRange && last24h.After(queryStartTime) {
			last24WhereClause, last24Args := buildWhereClause(last24h, queryEndTime, "", "")
			err = db.QueryRowContext(ctx, fmt.Sprintf(sqlSelectCountWithFilter, last24WhereClause), last24Args...).Scan(&last24Count)
			if err != nil {
				log.Printf("[violation] failed to get 24h count: %v", err)
			}
		} else if !query.UseCustomRange {
			err = db.QueryRowContext(ctx, sqlSelectLast24hCount, last24h).Scan(&last24Count)
			if err != nil {
				log.Printf("[violation] failed to get 24h count: %v", err)
			}
		}
		stats["violations_24h"] = last24Count

		return nil
	})

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Optimized to use a single query with window functions instead of nested queries
func GetViolationUsers(timeBack time.Duration) (map[string]interface{}, error) {
	var stats map[string]interface{}

	err := dbManager.GetDB(func(db *sqlx.DB) error {
		if timeBack <= 0 {
			timeBack = time.Duration(defaultTimeBackDays) * 24 * time.Hour
		}
		if timeBack > time.Duration(maxTimeBackDays)*24*time.Hour {
			timeBack = time.Duration(maxTimeBackDays) * 24 * time.Hour
		}

		startTime := time.Now().Add(-timeBack)
		stats = make(map[string]interface{})

		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()

		// Single optimized query: Get top 10 users with action breakdown in one pass
		// Uses SUM with CASE to aggregate actions
		userStats := make([]map[string]interface{}, 0)
		rows, err := db.QueryContext(ctx, sqlSelectUserStats, startTime)
		if err != nil {
			return fmt.Errorf("[violation] failed to get user stats: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var username string
			var total, blockCount, logonlyCount int
			if err := rows.Scan(&username, &total, &blockCount, &logonlyCount); err != nil {
				log.Printf("[violation] failed to scan user stat: %v", err)
				continue
			}

			actionBreakdown := make(map[string]int)
			if blockCount > 0 {
				actionBreakdown["block"] = blockCount
			}
			if logonlyCount > 0 {
				actionBreakdown["logonly"] = logonlyCount
			}

			userRecord := map[string]interface{}{
				"username":  username,
				"total":     total,
				"by_action": actionBreakdown,
			}
			userStats = append(userStats, userRecord)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("[violation] rows iteration error: %w", err)
		}

		stats["users"] = userStats
		return nil
	})

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ClearOldViolations removes violations older than specified duration
func ClearOldViolations(keepDuration time.Duration) error {
	return dbManager.GetDB(func(db *sqlx.DB) error {
		if keepDuration <= 0 {
			keepDuration = time.Duration(maxTimeBackDays) * 24 * time.Hour
		}

		cutoffTime := time.Now().Add(-keepDuration)

		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()

		result, err := db.ExecContext(ctx, sqlDeleteOldViolations, cutoffTime)
		if err != nil {
			return fmt.Errorf("[violation] failed to delete old violations: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("[violation] failed to get rows affected: %w", err)
		}

		if rowsAffected > 0 {
			log.Printf("[violation] cleared %d old violations before %s", rowsAffected, cutoffTime)
		}

		return nil
	})
}

// VacuumViolationDB optimizes database file
func VacuumViolationDB() error {
	return dbManager.GetDB(func(db *sqlx.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()

		if _, err := db.ExecContext(ctx, sqlVacuumDB); err != nil {
			return fmt.Errorf("[violation] vacuum failed: %w", err)
		}

		log.Println("[violation] database vacuumed")
		return nil
	})
}
