package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
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
	sqlCreateLoginHistoryTable = `
	CREATE TABLE IF NOT EXISTS login_history (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		login_time DATETIME NOT NULL,
		logout_time DATETIME,
		public_ip TEXT,
		private_ip TEXT,
		user_agent TEXT,
		session_id INTEGER,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	sqlCreateIndexLoginTime         = `CREATE INDEX IF NOT EXISTS idx_login_time ON login_history(login_time DESC)`
	sqlCreateIndexUsernameLogin     = `CREATE INDEX IF NOT EXISTS idx_username_login ON login_history(username)`
	sqlCreateIndexUsernameLoginTime = `CREATE INDEX IF NOT EXISTS idx_username_login_time ON login_history(username, login_time DESC)`
	sqlCreateIndexLogoutTime        = `CREATE INDEX IF NOT EXISTS idx_logout_time ON login_history(logout_time) WHERE logout_time IS NOT NULL`
)

// 插入操作相关 SQL
const (
	sqlInsertLoginRecordBasic = `INSERT INTO login_history (id, username, login_time, private_ip) VALUES (?, ?, ?, ?)`

	sqlCheckLatestSessionLogoutTime = `SELECT logout_time FROM login_history WHERE username = ? AND private_ip = ? ORDER BY login_time DESC LIMIT 1`

	sqlUpdateLoginDetails = `UPDATE login_history SET public_ip = ?, user_agent = ?, session_id = ? WHERE username = ? AND private_ip = ? AND public_ip IS NULL AND login_time = (SELECT MAX(login_time) FROM login_history WHERE username = ? AND private_ip = ? AND logout_time IS NULL)`

	sqlUpdateLogoutTime = `UPDATE login_history SET logout_time = ? WHERE username = ? AND private_ip = ? AND login_time = (SELECT MAX(login_time) FROM login_history WHERE username = ? AND private_ip = ? AND logout_time IS NULL)`

	sqlUpdateLogoutTimeWithoutPrivateIP = `UPDATE login_history SET logout_time = ? WHERE username = ? AND login_time = (SELECT MAX(login_time) FROM login_history WHERE username = ? AND logout_time IS NULL)`

	sqlUpdateLogoutTimeBySessionID = `UPDATE login_history SET logout_time = ? WHERE session_id = ? AND logout_time IS NULL`
)

// 查询操作相关 SQL
const (
	sqlSelectLoginHistoryCountWithFilter = `SELECT COUNT(0) FROM login_history WHERE %s`

	sqlSelectLoginHistoryWithFilter = `
	SELECT id, username, login_time, logout_time, public_ip, private_ip, user_agent, session_id, created_at
	FROM login_history
	WHERE %s
	ORDER BY login_time DESC
	LIMIT ? OFFSET ?`
)

// 删除和维护操作相关 SQL
const (
	sqlDeleteOldLoginRecords = `DELETE FROM login_history WHERE login_time < ?`
	sqlLoginHistoryVacuumDB  = `VACUUM`
)

// LoginHistory represents a single user login/logout record
type LoginHistory struct {
	ID         string     `db:"id" json:"id"`
	Username   string     `db:"username" json:"username"`
	LoginTime  time.Time  `db:"login_time" json:"login_time"`
	LogoutTime *time.Time `db:"logout_time" json:"logout_time,omitempty"`
	PublicIP   *string    `db:"public_ip" json:"public_ip,omitempty"`
	PrivateIP  string     `db:"private_ip" json:"private_ip"`
	UserAgent  *string    `db:"user_agent" json:"user_agent,omitempty"`
	SessionID  *int       `db:"session_id" json:"session_id,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

// LoginHistoryQuery defines filtering parameters for login history queries
type LoginHistoryQuery struct {
	TimeBack       time.Duration
	StartTime      time.Time
	EndTime        time.Time
	UseCustomRange bool
	Username       string
	Limit          int
	Offset         int
	ShowActiveOnly bool
}

const (
	LoginHistoryDefaultLimit        = 20
	LoginHistoryMaxLimit            = 1000
	LoginHistoryMaxTimeBackDays     = 365
	LoginHistoryDefaultTimeBackDays = 30
	LoginHistoryQueryTimeout        = 30 * time.Second
)

// RecordUserLoginBasic records a basic user login event with only username and private IP
// This is used for non-blocking login recording, with details filled in later
func RecordUserLoginBasic(username, privateIP string) error {
	username = strings.TrimSpace(username)
	privateIP = strings.TrimSpace(privateIP)

	if username == "" {
		return fmt.Errorf("[login_history] invalid username")
	}

	return GetDBManager().GetDB(func(db *sqlx.DB) error {
		// Check the latest session's logout_time for the same username and privateIP
		var logoutTime *time.Time
		err := db.QueryRow(sqlCheckLatestSessionLogoutTime, username, privateIP).Scan(&logoutTime)

		// If no record found, err will be sql.ErrNoRows, which means we can safely insert
		if err != nil {
			if err == sql.ErrNoRows {
				// No existing record found, proceed with insertion
			} else {
				return fmt.Errorf("[login_history] failed to check latest session: %w", err)
			}
		} else {
			// Record found, check if logout_time is NULL (which means session is still active)
			if logoutTime == nil {
				log.Printf("[login_history] latest session is still active for %s (%s), skipping insert", username, privateIP)
				return nil
			}
		}

		record := &LoginHistory{
			ID:        generateLoginHistoryID(),
			Username:  username,
			LoginTime: time.Now(),
			PrivateIP: privateIP,
		}

		_, err = db.Exec(sqlInsertLoginRecordBasic,
			record.ID,
			record.Username,
			record.LoginTime,
			record.PrivateIP,
		)
		if err != nil {
			return fmt.Errorf("[login_history] failed to insert basic login record: %w", err)
		}
		return nil
	})
}

// UpdateUserLoginDetails updates the detailed information for a basic login record
// This should be called after the user has successfully logged in and session details are available
func UpdateUserLoginDetails(username, privateIP, publicIP, userAgent string, sessionID int) error {
	username = strings.TrimSpace(username)
	privateIP = strings.TrimSpace(privateIP)
	publicIP = strings.TrimSpace(publicIP)
	userAgent = strings.TrimSpace(userAgent)

	if username == "" || publicIP == "" {
		return fmt.Errorf("[login_history] invalid username or publicIP")
	}

	return GetDBManager().GetDB(func(db *sqlx.DB) error {
		result, err := db.Exec(sqlUpdateLoginDetails,
			publicIP,
			userAgent,
			sessionID,
			username,
			privateIP,
			username,
			privateIP,
		)
		if err != nil {
			return fmt.Errorf("[login_history] failed to update login details: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("[login_history] failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			log.Printf("[login_history] no basic login record found to update for %s (%s)", username, privateIP)
			// Optionally, we could fall back to inserting a complete record
			// But this might indicate a timing issue or duplicate processing
		} else {
			log.Printf("[login_history] updated login details for %s (%s)", username, privateIP)
		}

		return nil
	})
}

// RecordUserLogout records a user logout event
func RecordUserLogout(username, privateIP string) error {
	username = strings.TrimSpace(username)
	privateIP = strings.TrimSpace(privateIP)

	if username == "" {
		return fmt.Errorf("[login_history] invalid username")
	}

	logoutTime := time.Now()

	return GetDBManager().GetDB(func(db *sqlx.DB) error {
		result, err := db.Exec(sqlUpdateLogoutTime, logoutTime, username, privateIP, username, privateIP)
		if err != nil {
			return fmt.Errorf("[login_history] failed to update logout time: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("[login_history] failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			// Fallback: try without private IP
			result, err = db.Exec(sqlUpdateLogoutTimeWithoutPrivateIP, logoutTime, username, username)
			if err != nil {
				return fmt.Errorf("[login_history] failed to update logout time (fallback): %w", err)
			}

			rowsAffected, err = result.RowsAffected()
			if err != nil {
				return fmt.Errorf("[login_history] failed to get rows affected (fallback): %w", err)
			}
		}

		if rowsAffected > 0 {
			log.Printf("[login_history] recorded logout for user %s, private IP %s", username, privateIP)
		} else {
			log.Printf("[login_history] no active login record found for user %s, private IP %s", username, privateIP)
		}

		return nil
	})
}

// RecordUserLogoutBySessionID records a user logout event by session ID
func RecordUserLogoutBySessionID(sessionID int) error {
	logoutTime := time.Now()

	return GetDBManager().GetDB(func(db *sqlx.DB) error {
		result, err := db.Exec(sqlUpdateLogoutTimeBySessionID, logoutTime, sessionID)
		if err != nil {
			return fmt.Errorf("[login_history] failed to update logout time by session ID: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("[login_history] failed to get rows affected by session ID: %w", err)
		}

		if rowsAffected > 0 {
			log.Printf("[login_history] recorded logout by session ID %d", sessionID)
		} else {
			log.Printf("[login_history] no active login record found for session ID %d", sessionID)
		}

		return nil
	})
}

// GetLoginHistory retrieves login history with pagination and optional filters
func GetLoginHistory(query LoginHistoryQuery) ([]LoginHistory, int64, error) {
	var records []LoginHistory
	var totalCount int64

	err := GetDBManager().GetDB(func(db *sqlx.DB) error {
		// Validate and sanitize parameters
		if query.Limit <= 0 {
			query.Limit = LoginHistoryDefaultLimit
		}
		if query.Limit > LoginHistoryMaxLimit {
			query.Limit = LoginHistoryMaxLimit
		}
		if query.Offset < 0 {
			query.Offset = 0
		}

		// Sanitize username
		query.Username = strings.TrimSpace(query.Username)

		// Calculate time range
		var queryStartTime, queryEndTime time.Time
		if query.UseCustomRange {
			queryStartTime = query.StartTime
			queryEndTime = query.EndTime
		} else {
			timeBack := query.TimeBack
			if timeBack <= 0 {
				timeBack = time.Duration(LoginHistoryDefaultTimeBackDays) * 24 * time.Hour
			}
			if timeBack > time.Duration(LoginHistoryMaxTimeBackDays)*24*time.Hour {
				timeBack = time.Duration(LoginHistoryMaxTimeBackDays) * 24 * time.Hour
			}
			queryStartTime = time.Now().Add(-timeBack)
			queryEndTime = time.Time{}
		}

		whereClause, args := buildLoginHistoryWhereClause(queryStartTime, queryEndTime, query.Username, query.ShowActiveOnly)

		// Get total count
		countQuery := fmt.Sprintf(sqlSelectLoginHistoryCountWithFilter, whereClause)
		countCtx, cancel := context.WithTimeout(context.Background(), LoginHistoryQueryTimeout)
		err := db.QueryRowContext(countCtx, countQuery, args...).Scan(&totalCount)
		cancel()
		if err != nil {
			return fmt.Errorf("[login_history] count query failed: %w", err)
		}

		// Get paginated results
		dataQuery := fmt.Sprintf(sqlSelectLoginHistoryWithFilter, whereClause)
		dataArgs := append(args, query.Limit, query.Offset)

		dataCtx, cancel := context.WithTimeout(context.Background(), LoginHistoryQueryTimeout)
		defer cancel()

		err = db.SelectContext(dataCtx, &records, dataQuery, dataArgs...)
		if err != nil {
			return fmt.Errorf("[login_history] query failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return records, totalCount, nil
}

// ClearOldLoginRecords removes login records older than specified duration
func ClearOldLoginRecords(keepDuration time.Duration) error {
	return GetDBManager().GetDB(func(db *sqlx.DB) error {
		if keepDuration <= 0 {
			keepDuration = time.Duration(LoginHistoryMaxTimeBackDays) * 24 * time.Hour
		}

		cutoffTime := time.Now().Add(-keepDuration)

		ctx, cancel := context.WithTimeout(context.Background(), LoginHistoryQueryTimeout)
		defer cancel()

		result, err := db.ExecContext(ctx, sqlDeleteOldLoginRecords, cutoffTime)
		if err != nil {
			return fmt.Errorf("[login_history] failed to delete old login records: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("[login_history] failed to get rows affected: %w", err)
		}

		if rowsAffected > 0 {
			log.Printf("[login_history] cleared %d old login records before %s", rowsAffected, cutoffTime)
		}

		return nil
	})
}

// generateLoginHistoryID generates a proper UUID v4 for records
func generateLoginHistoryID() string {
	return uuid.New().String()
}

// buildLoginHistoryWhereClause builds SQL WHERE clause with dynamic filter conditions
func buildLoginHistoryWhereClause(startTime, endTime time.Time, username string, showActiveOnly bool) (string, []interface{}) {
	var whereConditions []string
	var args []interface{}

	// Handle timestamp conditions
	if endTime.IsZero() {
		whereConditions = append(whereConditions, "login_time >= ?")
		args = append(args, startTime)
	} else {
		whereConditions = append(whereConditions, "login_time >= ? AND login_time < ?")
		args = append(args, startTime, endTime)
	}

	if username != "" {
		whereConditions = append(whereConditions, "username = ?")
		args = append(args, username)
	}

	if showActiveOnly {
		whereConditions = append(whereConditions, "logout_time IS NULL")
	}

	whereClause := strings.Join(whereConditions, " AND ")
	return whereClause, args
}
