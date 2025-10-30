package storage

import (
	"database/sql"
	"fmt"
)

// PartRunStats represents the latest runs grouped by part
type PartRunStats struct {
	Group        string  `json:"group"`        // Group name (e.g., "frontend", "backend") or empty
	Part         string  `json:"part"`         // Part name or full path (e.g., "deploy" or "frontend.deploy")
	RunID        int     `json:"run_id"`
	Status       string  `json:"status"`
	Duration     *string `json:"duration,omitempty"`
	StartedAt    string  `json:"started_at"`
	StepCount    int     `json:"step_count"`
}

// GetLatestRunsByPart returns the latest runs for each part of a project
func (s *Storage) GetLatestRunsByPart(projectName string, limit int) ([]PartRunStats, error) {
	// Simple query without window functions for better SQLite compatibility
	query := `
		SELECT 
			COALESCE(r."group", '') as "group",
			r.part,
			r.id,
			r.status,
			r.duration,
			r.started_at,
			COUNT(se.id) as step_count
		FROM runs r
		LEFT JOIN step_executions se ON r.id = se.run_id
		WHERE r.project_name = ?
		GROUP BY r.id, r."group", r.part, r.status, r.duration, r.started_at
		ORDER BY r."group", r.part, r.started_at DESC
	`

	rows, err := s.db.Query(query, projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest runs: %w", err)
	}
	defer rows.Close()

	// Group by part and limit per part
	partCounts := make(map[string]int)
	stats := make([]PartRunStats, 0)

	for rows.Next() {
		var stat PartRunStats
		var duration sql.NullString

		err := rows.Scan(
			&stat.Group,
			&stat.Part,
			&stat.RunID,
			&stat.Status,
			&duration,
			&stat.StartedAt,
			&stat.StepCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run stats: %w", err)
		}

		// Limit runs per part
		partKey := stat.Group + "." + stat.Part
		if partCounts[partKey] >= limit {
			continue
		}
		partCounts[partKey]++

		if duration.Valid {
			durationStr := duration.String
			stat.Duration = &durationStr
		}

		stats = append(stats, stat)
	}

	return stats, rows.Err()
}
