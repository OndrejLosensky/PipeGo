package storage

import "time"

// Run represents a pipeline execution
type Run struct {
	ID          int        `json:"id"`
	Status      string     `json:"status"` // "running", "success", "failed"
	ConfigPath  string     `json:"config_path"`
	ProjectName string     `json:"project_name"`
	Group       string     `json:"group"` // The group (e.g., "frontend", "backend") or empty for ungrouped
	Part        string     `json:"part"`  // The part being executed (e.g., "deploy", "tests" or full path "frontend.deploy")
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	Duration    *string    `json:"duration,omitempty"`
}

// StepExecution represents execution of a single step
type StepExecution struct {
	ID         int        `json:"id"`
	RunID      int        `json:"run_id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"` // "running", "success", "failed"
	Command    string     `json:"command"`
	Output     string     `json:"output"`
	Group      string     `json:"group"`    // The group this step belongs to
	Part       string     `json:"part"`     // The part this step belongs to
	Category   string     `json:"category"` // The category (tests, deploy, setup, etc.)
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Duration   *string    `json:"duration,omitempty"`
}

