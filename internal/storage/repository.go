package storage

import (
	"context"
	"database/sql"
	"fmt"

	"logmind/internal/domain"
)

type LogRepository interface {
	InsertBatch(ctx context.Context, logs []domain.Log) error
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
