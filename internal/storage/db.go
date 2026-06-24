package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_sync=NORMAL&_busy_timeout=5000", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1) // Single writer for SQLite
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	createTable := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service TEXT,
		level TEXT,
		message TEXT,
		timestamp INTEGER
	);
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
	`
	if _, err = db.Exec(createTable); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}
	return db, nil
}
