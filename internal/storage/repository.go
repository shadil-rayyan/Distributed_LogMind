package storage

import (
	"context"
	"database/sql"
	"fmt"

	"logmind/internal/domain"
)

type LogRepository interface {
	InsertBatch(ctx context.Context, logs []domain.Log) error
	RecentErrors(ctx context.Context, sinceUnix int64) ([]domain.Log, error)
}

type SQLiteLogRepository struct {
	db *sql.DB
}

func NewSQLiteLogRepository(db *sql.DB) *SQLiteLogRepository {
	return &SQLiteLogRepository{db: db}
}

func (r *SQLiteLogRepository) InsertBatch(ctx context.Context, logs []domain.Log) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO logs(service, level, message, timestamp) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, l := range logs {
		if _, err := stmt.ExecContext(ctx, l.Service, l.Level, l.Message, l.Timestamp); err != nil {
			return fmt.Errorf("exec statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (r *SQLiteLogRepository) RecentErrors(ctx context.Context, sinceUnix int64) ([]domain.Log, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT service, level, message, timestamp
		FROM logs
		WHERE level = ? AND timestamp >= ?
		ORDER BY timestamp ASC
	`, "error", sinceUnix)
	if err != nil {
		return nil, fmt.Errorf("query recent errors: %w", err)
	}
	defer rows.Close()

	logs := make([]domain.Log, 0)
	for rows.Next() {
		var l domain.Log
		if err := rows.Scan(&l.Service, &l.Level, &l.Message, &l.Timestamp); err != nil {
			return nil, fmt.Errorf("scan recent errors: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent errors: %w", err)
	}

	return logs, nil
}
