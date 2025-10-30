package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// CreateStepExecution creates a new step execution record
func (s *Storage) CreateStepExecution(runID int, name, command, groupName, part, category string) (*StepExecution, error) {
	now := time.Now()
	result, err := s.db.Exec(
		`INSERT INTO step_executions (run_id, name, status, command, "group", part, category, started_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		runID, name, "running", command, groupName, part, category, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create step execution: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get step execution ID: %w", err)
	}

	return &StepExecution{
		ID:        int(id),
		RunID:     runID,
		Name:      name,
		Status:    "running",
		Command:   command,
		Group:     groupName,
		Part:      part,
		Category:  category,
		StartedAt: now,
	}, nil
}

// UpdateStepExecution updates step execution with output, status, and finish time
func (s *Storage) UpdateStepExecution(stepID int, status, output string, duration time.Duration) error {
	now := time.Now()
	durationStr := duration.String()
	_, err := s.db.Exec(
		"UPDATE step_executions SET status = ?, output = ?, finished_at = ?, duration = ? WHERE id = ?",
		status, output, now, durationStr, stepID,
	)
	if err != nil {
		return fmt.Errorf("failed to update step execution: %w", err)
	}
	return nil
}

// GetStepExecutions retrieves all step executions for a run
func (s *Storage) GetStepExecutions(runID int) ([]*StepExecution, error) {
	rows, err := s.db.Query(
		`SELECT id, run_id, name, status, command, output, "group", part, category, started_at, finished_at, duration FROM step_executions WHERE run_id = ? ORDER BY id ASC`,
		runID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query step executions: %w", err)
	}
	defer rows.Close()

	var steps []*StepExecution
	for rows.Next() {
		var step StepExecution
		var output sql.NullString
		var finishedAt sql.NullTime
		var duration sql.NullString

		err := rows.Scan(&step.ID, &step.RunID, &step.Name, &step.Status, &step.Command, &output, &step.Group, &step.Part, &step.Category, &step.StartedAt, &finishedAt, &duration)
		if err != nil {
			return nil, fmt.Errorf("failed to scan step execution: %w", err)
		}

		if output.Valid {
			step.Output = output.String
		}
		if finishedAt.Valid {
			step.FinishedAt = &finishedAt.Time
		}
		if duration.Valid {
			durationStr := duration.String
			step.Duration = &durationStr
		}

		steps = append(steps, &step)
	}

	return steps, rows.Err()
}

