package storage

import "time"

// Run represents a pipeline execution
type Run struct {
	ID         int        `json:"id"`
	Status     string     `json:"status"` // "running", "success", "failed"
	ConfigPath string     `json:"config_path"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Duration   *string    `json:"duration,omitempty"`
}

// StepExecution represents execution of a single step
type StepExecution struct {
	ID         int        `json:"id"`
	RunID      int        `json:"run_id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"` // "running", "success", "failed"
	Command    string     `json:"command"`
	Output     string     `json:"output"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Duration   *string    `json:"duration,omitempty"`
}

