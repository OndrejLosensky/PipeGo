package runner

import (
	"time"

	"pipego/runner/storage"
)

// PipelineResult represents the result of running a pipeline
type PipelineResult struct {
	Status   string        `json:"status"` // "success" or "failed"
	RunID    int           `json:"run_id"`
	Steps    []StepResult  `json:"steps"`
	Duration time.Duration `json:"duration"`
	Error    error         `json:"error,omitempty"`
}

// StepResult represents the result of executing a single step
type StepResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"` // "success" or "failed"
	Output   string        `json:"output"`
	Duration time.Duration `json:"duration"`
	Error    error         `json:"error,omitempty"`
}

// RunPipelineOptions configures how the pipeline should be executed
type RunPipelineOptions struct {
	Storage          *storage.Storage // Optional storage for database persistence
	StreamToTerminal bool             // If true, also stream output to terminal
	PartFilter       string           // Optional: run only this specific part (empty = run all)
}

