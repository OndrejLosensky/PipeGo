package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"pipego/runner/storage"
)

// RunPipeline executes a pipeline defined in the config file
func RunPipeline(configPath string) error {
	_, err := RunPipelineWithOptions(configPath, RunPipelineOptions{StreamToTerminal: true})
	return err
}

// RunPipelineWithOptions executes a pipeline with options for storage and streaming
func RunPipelineWithOptions(configPath string, opts RunPipelineOptions) (*PipelineResult, error) {
	startTime := time.Now()

	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Change to the directory where config file is located
	configDir := filepath.Dir(configPath)
	originalDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	err = os.Chdir(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to change to config directory: %w", err)
	}
	defer os.Chdir(originalDir)

	// Extract project name from config path (directory name)
	projectName := filepath.Base(configDir)

	// Get all parts from config
	allParts := cfg.GetAllParts()

	result := &PipelineResult{
		RunID:  0,
		Steps:  make([]StepResult, 0),
		Status: "running",
	}

	// Execute each part
	for partName, steps := range allParts {
		// Skip this part if filter is set and doesn't match
		if opts.PartFilter != "" && partName != opts.PartFilter {
			continue
		}

		if opts.StreamToTerminal {
			if partName != "default" {
				fmt.Printf("\n📦 Part: %s\n", partName)
			}
		}

		// Create run in database for this part if storage is provided
		var run *storage.Run
		if opts.Storage != nil {
			run, err = opts.Storage.CreateRun(configPath, projectName, partName)
			if err != nil {
				return nil, fmt.Errorf("failed to create run: %w", err)
			}
			result.RunID = run.ID
		}

		// Execute each step in the part
		for _, step := range steps {
			stepResult, err := executeStep(step, partName, result.RunID, opts)
			
			result.Steps = append(result.Steps, stepResult)
			
			if err != nil {
				result.Status = "failed"
				result.Duration = time.Since(startTime)
				result.Error = err

				// Update run status in database
				if opts.Storage != nil {
					_ = opts.Storage.UpdateRunStatus(result.RunID, "failed", result.Duration)
				}

				return result, err
			}
		}

		// Update run status for this part
		if opts.Storage != nil {
			partDuration := time.Since(startTime)
			err = opts.Storage.UpdateRunStatus(result.RunID, "success", partDuration)
			if err != nil {
				return nil, fmt.Errorf("failed to update run status: %w", err)
			}
		}
	}

	result.Status = "success"
	result.Duration = time.Since(startTime)

	if opts.StreamToTerminal {
		fmt.Println("\n🏁 All steps finished successfully.")
	}

	return result, nil
}

// executeStep executes a single step and returns its result
func executeStep(step Step, partName string, runID int, opts RunPipelineOptions) (StepResult, error) {
	stepStart := time.Now()

	if opts.StreamToTerminal {
		fmt.Println("→", step.Name)
	}

	// Use empty string for category if not set
	category := step.Category
	if category == "" {
		category = ""
	}

	// Create step execution record if storage is provided
	var stepExec *storage.StepExecution
	var err error
	if opts.Storage != nil {
		stepExec, err = opts.Storage.CreateStepExecution(runID, step.Name, step.Run, partName, category)
		if err != nil {
			return StepResult{}, fmt.Errorf("failed to create step execution: %w", err)
		}
	}

	// Execute the command and capture output
	output, err := executeShellCommand(step.Run, opts.StreamToTerminal)
	stepDuration := time.Since(stepStart)

	stepResult := StepResult{
		Name:     step.Name,
		Output:   output,
		Duration: stepDuration,
	}

	if err != nil {
		stepResult.Status = "failed"
		stepResult.Error = err

		if opts.StreamToTerminal {
			fmt.Println("❌ Step failed:", err)
		}

		// Update step execution in database
		if opts.Storage != nil && stepExec != nil {
			_ = opts.Storage.UpdateStepExecution(stepExec.ID, "failed", output, stepDuration)
		}

		return stepResult, fmt.Errorf("step '%s' failed: %w", step.Name, err)
	}

	stepResult.Status = "success"

	if opts.StreamToTerminal {
		fmt.Println("✅ Done:", step.Name)
	}

	// Update step execution in database
	if opts.Storage != nil && stepExec != nil {
		err = opts.Storage.UpdateStepExecution(stepExec.ID, "success", output, stepDuration)
		if err != nil {
			return StepResult{}, fmt.Errorf("failed to update step execution: %w", err)
		}
	}

	return stepResult, nil
}

// executeShellCommand executes a shell command and captures its output
func executeShellCommand(command string, streamToTerminal bool) (string, error) {
	cmd := exec.Command("bash", "-c", command)

	var stdout, stderr bytes.Buffer
	var stdoutWriters []io.Writer
	var stderrWriters []io.Writer

	// Always capture output
	stdoutWriters = append(stdoutWriters, &stdout)
	stderrWriters = append(stderrWriters, &stderr)

	// Optionally also stream to terminal
	if streamToTerminal {
		stdoutWriters = append(stdoutWriters, os.Stdout)
		stderrWriters = append(stderrWriters, os.Stderr)
	}

	cmd.Stdout = io.MultiWriter(stdoutWriters...)
	cmd.Stderr = io.MultiWriter(stderrWriters...)

	err := cmd.Run()

	// Combine stdout and stderr
	combinedOutput := stdout.String() + stderr.String()
	if len(combinedOutput) > 0 && combinedOutput[len(combinedOutput)-1] != '\n' {
		combinedOutput += "\n"
	}

	return combinedOutput, err
}
