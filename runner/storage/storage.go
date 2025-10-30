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

// initSchema creates the database tables and handles migrations
func (s *Storage) initSchema() error {
	// Create tables with new schema
	queries := []string{
		`CREATE TABLE IF NOT EXISTS runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			status TEXT NOT NULL,
			config_path TEXT NOT NULL,
			project_name TEXT NOT NULL DEFAULT '',
			part TEXT NOT NULL DEFAULT 'default',
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
			part TEXT NOT NULL DEFAULT 'default',
			category TEXT NOT NULL DEFAULT '',
			started_at DATETIME NOT NULL,
			finished_at DATETIME,
			duration TEXT,
			FOREIGN KEY(run_id) REFERENCES runs(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_started_at ON runs(started_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_project_name ON runs(project_name)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_part ON runs(part)`,
		`CREATE INDEX IF NOT EXISTS idx_step_executions_run_id ON step_executions(run_id)`,
		`CREATE INDEX IF NOT EXISTS idx_step_executions_part ON step_executions(part)`,
		`CREATE INDEX IF NOT EXISTS idx_step_executions_category ON step_executions(category)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	// Migrate existing tables if needed
	if err := s.migrateSchema(); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}

	return nil
}

// migrateSchema adds new columns to existing tables if they don't exist
func (s *Storage) migrateSchema() error {
	migrations := []string{
		// Add project_name to runs if it doesn't exist
		`ALTER TABLE runs ADD COLUMN project_name TEXT NOT NULL DEFAULT ''`,
		// Add part to runs if it doesn't exist
		`ALTER TABLE runs ADD COLUMN part TEXT NOT NULL DEFAULT 'default'`,
		// Add part to step_executions if it doesn't exist
		`ALTER TABLE step_executions ADD COLUMN part TEXT NOT NULL DEFAULT 'default'`,
		// Add category to step_executions if it doesn't exist
		`ALTER TABLE step_executions ADD COLUMN category TEXT NOT NULL DEFAULT ''`,
	}

	for _, migration := range migrations {
		// Ignore errors if column already exists
		s.db.Exec(migration)
	}

	return nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

