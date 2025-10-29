package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// CreateRun creates a new run record
func (s *Storage) CreateRun(configPath string) (*Run, error) {
	now := time.Now()
	result, err := s.db.Exec(
		"INSERT INTO runs (status, config_path, started_at) VALUES (?, ?, ?)",
		"running", configPath, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get run ID: %w", err)
	}

	return &Run{
		ID:         int(id),
		Status:     "running",
		ConfigPath: configPath,
		StartedAt:  now,
	}, nil
}

// UpdateRunStatus updates the status and finish time of a run
func (s *Storage) UpdateRunStatus(runID int, status string, duration time.Duration) error {
	now := time.Now()
	durationStr := duration.String()
	_, err := s.db.Exec(
		"UPDATE runs SET status = ?, finished_at = ?, duration = ? WHERE id = ?",
		status, now, durationStr, runID,
	)
	if err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}
	return nil
}

// GetRuns retrieves all runs, ordered by most recent first
func (s *Storage) GetRuns(limit int) ([]*Run, error) {
	query := "SELECT id, status, config_path, started_at, finished_at, duration FROM runs ORDER BY started_at DESC LIMIT ?"
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query runs: %w", err)
	}
	defer rows.Close()

	var runs []*Run
	for rows.Next() {
		var r Run
		var finishedAt sql.NullTime
		var duration sql.NullString

		err := rows.Scan(&r.ID, &r.Status, &r.ConfigPath, &r.StartedAt, &finishedAt, &duration)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}

		if finishedAt.Valid {
			r.FinishedAt = &finishedAt.Time
		}
		if duration.Valid {
			durationStr := duration.String
			r.Duration = &durationStr
		}

		runs = append(runs, &r)
	}

	return runs, rows.Err()
}

// GetRun retrieves a single run by ID
func (s *Storage) GetRun(runID int) (*Run, error) {
	var r Run
	var finishedAt sql.NullTime
	var duration sql.NullString

	err := s.db.QueryRow(
		"SELECT id, status, config_path, started_at, finished_at, duration FROM runs WHERE id = ?",
		runID,
	).Scan(&r.ID, &r.Status, &r.ConfigPath, &r.StartedAt, &finishedAt, &duration)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("run not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	if finishedAt.Valid {
		r.FinishedAt = &finishedAt.Time
	}
	if duration.Valid {
		durationStr := duration.String
		r.Duration = &durationStr
	}

	return &r, nil
}

