package internal

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// ============================================================================
// 统一数据库管理器
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
		return fmt.Errorf("[database] database not initialized")
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

// 全局统一数据库管理器实例
var dbManager = &DBManager{}

// 统一数据库路径
var dbPath = "data/ocserv_data.db"

// PRAGMA 设置 - 用于统一数据库
var sqlPragmas = []string{
	"PRAGMA synchronous=NORMAL",
	"PRAGMA cache_size=20000",
	"PRAGMA temp_store=MEMORY",
	"PRAGMA foreign_keys=ON",
	"PRAGMA query_only=FALSE",
	"PRAGMA busy_timeout=5000",
}

// InitializeDB initializes the SQLite database
func InitializeDB(dataPath string) error {
	if dataPath != "" {
		dbPath = dataPath + "/ocserv_data.db"
	}

	var err error
	db, err := sqlx.Open("sqlite3", dbPath+"?cache=shared&mode=rwc&_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("[database] failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		return fmt.Errorf("[database] failed to ping database: %w", err)
	}

	// Set connection pool parameters for SQLite with optimized settings
	db.SetMaxOpenConns(10) // Increased from 5 to handle both login history and violations
	db.SetMaxIdleConns(3)  // Increased from 2
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Set pragmas for performance and safety
	for _, pragma := range sqlPragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("[database] failed to set pragma: %w", err)
		}
	}

	// Create tables with proper schema (both login_history and violations tables)
	if err := createAllTables(db); err != nil {
		return fmt.Errorf("[database] failed to create tables: %w", err)
	}

	// Create indexes for both tables
	if err := createAllIndexes(db); err != nil {
		return fmt.Errorf("[database] failed to create indexes: %w", err)
	}

	// Set the database in manager
	dbManager.SetDB(db)

	log.Printf("[database] SQLite3 database initialized at: %s", dbPath)
	return nil
}

// CloseDB closes the database
func CloseDB() error {
	return dbManager.Close()
}

// VacuumDB optimizes database file
func VacuumDB() error {
	return dbManager.GetDB(func(db *sqlx.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if _, err := db.ExecContext(ctx, "VACUUM"); err != nil {
			return fmt.Errorf("[database] vacuum failed: %w", err)
		}

		log.Println("[database] database vacuumed")
		return nil
	})
}

// Helper functions for table and index creation

func createAllTables(db *sqlx.DB) error {
	// Create login_history table
	if _, err := db.Exec(sqlCreateLoginHistoryTable); err != nil {
		return fmt.Errorf("failed to create login_history table: %w", err)
	}

	// Create violations table
	if _, err := db.Exec(sqlCreateViolationsTable); err != nil {
		return fmt.Errorf("failed to create violations table: %w", err)
	}

	return nil
}

func createAllIndexes(db *sqlx.DB) error {
	// Login history indexes
	loginIndexes := []string{
		sqlCreateIndexLoginTime,
		sqlCreateIndexUsernameLogin,
		sqlCreateIndexUsernameLoginTime,
		sqlCreateIndexLogoutTime,
	}

	for _, idx := range loginIndexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create login history index: %w", err)
		}
	}

	// Violations indexes
	violationIndexes := []string{
		sqlCreateIndexTimestamp,
		sqlCreateIndexUsername,
		sqlCreateIndexUsernameTimestamp,
		sqlCreateIndexAction,
		sqlCreateIndexActionTimestamp,
	}

	for _, idx := range violationIndexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create violation index: %w", err)
		}
	}

	return nil
}

// Utility functions to access the database manager from other packages

// GetDBManager returns the database manager instance
func GetDBManager() *DBManager {
	return dbManager
}

// GetDBPath returns the current database path
func GetDBPath() string {
	return dbPath
}
