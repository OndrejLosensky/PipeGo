package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Storage handles database operations
type Storage struct {
	db *sql.DB
}

// NewStorage creates a new storage instance
func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{db: db}
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the database tables
func (s *Storage) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			status TEXT NOT NULL,
			config_path TEXT NOT NULL,
			started_at DATETIME NOT NULL,
			finished_at DATETIME,
			duration TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS step_executions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			command TEXT NOT NULL,
			output TEXT,
			started_at DATETIME NOT NULL,
			finished_at DATETIME,
			duration TEXT,
			FOREIGN KEY(run_id) REFERENCES runs(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_started_at ON runs(started_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_step_executions_run_id ON step_executions(run_id)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

