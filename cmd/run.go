package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"pipego/runner"
	"pipego/runner/storage"
)

// Run executes the 'run' command
func Run(configPath string) error {

	// Determine database path (its stored in data directory in current working directory)
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	dataDir := filepath.Join(cwd, "data")

	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	dbPath := filepath.Join(dataDir, "pipego.db")

	// Initialize storage	
	store, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()	

	// Run pipeline with storage and streaming to terminal
	result, err := runner.RunPipelineWithOptions(configPath, runner.RunPipelineOptions{
		Storage:          store,
		StreamToTerminal: true, // Always stream to console for local development
	})

	if err != nil {
		log.Fatalf("Pipeline failed: %v", err)
	}

	fmt.Printf("\n📊 Run ID: %d | Status: %s | Duration: %s\n", result.RunID, result.Status, result.Duration)

	return nil
}

