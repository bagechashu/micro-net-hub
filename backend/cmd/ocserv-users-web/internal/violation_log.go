package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// ViolationLog represents a single VPN access violation record
type ViolationLog struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	RemoteIP  string    `json:"remote_ip"`
	Action    string    `json:"action"` // "block" or "logonly"
	Reason    string    `json:"reason"` // reason for violation
	Timestamp time.Time `json:"timestamp"`
}

// PaginationParams defines pagination parameters for queries
type PaginationParams struct {
	Offset int
	Limit  int
}

// ViolationQuery defines filtering parameters for violation queries
type ViolationQuery struct {
	TimeBack time.Duration
	Username string
	Action   string
	Limit    int
	Offset   int
}

const (
	defaultLimit        = 20
	maxLimit            = 1000
	maxTimeBackDays     = 365
	defaultTimeBackDays = 7
	queryTimeout        = 30 * time.Second
)

var (
	violationDB *sql.DB
	dbMutex     sync.RWMutex
	dbPath      = "data/violations.db"
)

// generateID generates a proper UUID v4 for records
func generateID() string {
	return uuid.New().String()
}

// InitializeViolationDB initializes the SQLite violation logging database
func InitializeViolationDB(dataPath string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if dataPath != "" {
		dbPath = dataPath + "/violations.db"
	}

	var err error
	// Use mattn/go-sqlite3 driver with optimized DSN
	violationDB, err = sql.Open("sqlite3", dbPath+"?cache=shared&mode=rwc&_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("[violation] failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = violationDB.PingContext(ctx)
	cancel()
	if err != nil {
		return fmt.Errorf("[violation] failed to ping database: %w", err)
	}

	// Set connection pool parameters for better concurrency
	violationDB.SetMaxOpenConns(50)
	violationDB.SetMaxIdleConns(10)
	violationDB.SetConnMaxLifetime(30 * time.Minute)
	violationDB.SetConnMaxIdleTime(5 * time.Minute)

	// Set pragmas for performance and safety
	pragmaQueries := []string{
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=20000",
		"PRAGMA temp_store=MEMORY",
		"PRAGMA foreign_keys=ON",
		"PRAGMA query_only=FALSE",
		"PRAGMA busy_timeout=5000",
	}

	for _, pragma := range pragmaQueries {
		if _, err := violationDB.Exec(pragma); err != nil {
			return fmt.Errorf("[violation] failed to set pragma: %w", err)
		}
	}

	// Create tables with proper schema
	schema := `
	CREATE TABLE IF NOT EXISTS violations (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		remote_ip TEXT NOT NULL,
		action TEXT NOT NULL CHECK(action IN ('block', 'logonly')),
		reason TEXT,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_timestamp ON violations(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_username ON violations(username);
	CREATE INDEX IF NOT EXISTS idx_username_timestamp ON violations(username, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_action ON violations(action);
	CREATE INDEX IF NOT EXISTS idx_action_timestamp ON violations(action, timestamp DESC);
	`

	if _, err := violationDB.Exec(schema); err != nil {
		return fmt.Errorf("[violation] failed to create schema: %w", err)
	}

	log.Printf("[violation] SQLite3 database initialized at: %s", dbPath)
	return nil
}

// CloseViolationDB closes the violation database
func CloseViolationDB() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if violationDB != nil {
		return violationDB.Close()
	}
	return nil
}

// RecordViolation records a VPN access violation
func RecordViolation(username, remoteIP, action, reason string) error {
	dbMutex.RLock()
	defer dbMutex.RUnlock()

	if violationDB == nil {
		return fmt.Errorf("[violation] database not initialized")
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

	id := generateID()
	now := time.Now()

	query := `
	INSERT INTO violations (id, username, remote_ip, action, reason, timestamp)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	_, err := violationDB.ExecContext(ctx, query, id, username, remoteIP, action, reason, now)
	if err != nil {
		return fmt.Errorf("[violation] failed to insert violation: %w", err)
	}

	return nil
}

// GetViolations retrieves violations with pagination and optional filters
// Returns (violations, total_count, error)
func GetViolations(query ViolationQuery) ([]ViolationLog, int64, error) {
	dbMutex.RLock()
	defer dbMutex.RUnlock()

	if violationDB == nil {
		return nil, 0, fmt.Errorf("[violation] database not initialized")
	}

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

	if query.TimeBack <= 0 {
		query.TimeBack = time.Duration(defaultTimeBackDays) * 24 * time.Hour
	}
	if query.TimeBack > time.Duration(maxTimeBackDays)*24*time.Hour {
		query.TimeBack = time.Duration(maxTimeBackDays) * 24 * time.Hour
	}

	// Sanitize username and action
	query.Username = strings.TrimSpace(query.Username)
	query.Action = strings.ToLower(strings.TrimSpace(query.Action))
	if query.Action != "" && query.Action != "block" && query.Action != "logonly" {
		query.Action = ""
	}

	startTime := time.Now().Add(-query.TimeBack)

	// Build the WHERE clause dynamically
	whereConditions := []string{"timestamp >= ?"}
	args := []interface{}{startTime}

	if query.Username != "" {
		whereConditions = append(whereConditions, "username = ?")
		args = append(args, query.Username)
	}

	if query.Action != "" {
		whereConditions = append(whereConditions, "action = ?")
		args = append(args, query.Action)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM violations WHERE %s", whereClause)
	countCtx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	var totalCount int64
	err := violationDB.QueryRowContext(countCtx, countQuery, args...).Scan(&totalCount)
	cancel()
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, fmt.Errorf("[violation] count query failed: %w", err)
	}

	// Get paginated results
	dataQuery := fmt.Sprintf(`
	SELECT id, username, remote_ip, action, reason, timestamp
	FROM violations
	WHERE %s
	ORDER BY timestamp DESC
	LIMIT ? OFFSET ?
	`, whereClause)

	dataArgs := append(args, query.Limit, query.Offset)

	dataCtx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	rows, err := violationDB.QueryContext(dataCtx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("[violation] query failed: %w", err)
	}
	defer rows.Close()

	var violations []ViolationLog
	for rows.Next() {
		var v ViolationLog
		if err := rows.Scan(&v.ID, &v.Username, &v.RemoteIP, &v.Action, &v.Reason, &v.Timestamp); err != nil {
			log.Printf("[violation] failed to scan record: %v", err)
			continue
		}
		violations = append(violations, v)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("[violation] rows error: %w", err)
	}

	return violations, totalCount, nil
}

// GetViolationsByUsername retrieves all violations for a specific user (deprecated, use GetViolations)
func GetViolationsByUsername(username string) ([]ViolationLog, error) {
	query := ViolationQuery{
		TimeBack: time.Duration(defaultTimeBackDays) * 24 * time.Hour,
		Username: username,
		Limit:    maxLimit,
	}

	violations, _, err := GetViolations(query)
	return violations, err
}

// GetViolationStats retrieves statistics for violations within a time range
func GetViolationStats(timeBack time.Duration) (map[string]interface{}, error) {
	dbMutex.RLock()
	defer dbMutex.RUnlock()

	if violationDB == nil {
		return nil, fmt.Errorf("[violation] database not initialized")
	}

	if timeBack <= 0 {
		timeBack = time.Duration(defaultTimeBackDays) * 24 * time.Hour
	}
	if timeBack > time.Duration(maxTimeBackDays)*24*time.Hour {
		timeBack = time.Duration(maxTimeBackDays) * 24 * time.Hour
	}

	startTime := time.Now().Add(-timeBack)
	stats := make(map[string]interface{})

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	// Get total count
	var totalCount int
	err := violationDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM violations WHERE timestamp >= ?", startTime).Scan(&totalCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("[violation] failed to get total count: %w", err)
	}
	stats["total_violations"] = totalCount

	// Get action breakdown
	actionStats := make(map[string]int)
	rows, err := violationDB.QueryContext(ctx, `
	SELECT action, COUNT(*) as count
	FROM violations
	WHERE timestamp >= ?
	GROUP BY action
	`, startTime)
	if err != nil {
		return nil, fmt.Errorf("[violation] failed to get action stats: %w", err)
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

	// Get top 10 users by violation count
	userStats := make(map[string]int)
	rows, err = violationDB.QueryContext(ctx, `
	SELECT username, COUNT(*) as count
	FROM violations
	WHERE timestamp >= ?
	GROUP BY username
	ORDER BY count DESC
	LIMIT 10
	`, startTime)
	if err != nil {
		return nil, fmt.Errorf("[violation] failed to get user stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		var count int
		if err := rows.Scan(&username, &count); err != nil {
			log.Printf("[violation] failed to scan user stat: %v", err)
			continue
		}
		userStats[username] = count
	}
	stats["by_user"] = userStats

	// Get recent violations (last 24 hours) count
	last24h := time.Now().Add(-24 * time.Hour)
	var last24Count int
	err = violationDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM violations WHERE timestamp >= ?", last24h).Scan(&last24Count)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[violation] failed to get 24h count: %v", err)
	}
	stats["violations_24h"] = last24Count

	return stats, nil
}

// ClearOldViolations removes violations older than specified duration
func ClearOldViolations(keepDuration time.Duration) error {
	dbMutex.RLock()
	defer dbMutex.RUnlock()

	if violationDB == nil {
		return fmt.Errorf("[violation] database not initialized")
	}

	if keepDuration <= 0 {
		keepDuration = time.Duration(maxTimeBackDays) * 24 * time.Hour
	}

	cutoffTime := time.Now().Add(-keepDuration)

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	result, err := violationDB.ExecContext(ctx, "DELETE FROM violations WHERE timestamp < ?", cutoffTime)
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
}

// VacuumViolationDB optimizes database file
func VacuumViolationDB() error {
	dbMutex.RLock()
	defer dbMutex.RUnlock()

	if violationDB == nil {
		return fmt.Errorf("[violation] database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	if _, err := violationDB.ExecContext(ctx, "VACUUM"); err != nil {
		return fmt.Errorf("[violation] vacuum failed: %w", err)
	}

	log.Println("[violation] database vacuumed")
	return nil
}
