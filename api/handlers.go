package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"pipego/runner"
	"pipego/runner/storage"
)

// GetRuns returns all runs
func GetRuns(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		runs, err := store.GetRuns(100) // Limit to 100 most recent
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get runs: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(runs)
	}
}

// GetRun returns a single run with its steps
func GetRun(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse run ID from URL: /api/runs/:id
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		runID, err := strconv.Atoi(pathParts[2])
		if err != nil {
			http.Error(w, "Invalid run ID", http.StatusBadRequest)
			return
		}

		// Get run
		run, err := store.GetRun(runID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Run not found: %v", err), http.StatusNotFound)
			return
		}

		// Get steps
		steps, err := store.GetStepExecutions(runID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get steps: %v", err), http.StatusInternalServerError)
			return
		}

		// Build response
		type RunResponse struct {
			Run   *storage.Run             `json:"run"`
			Steps []*storage.StepExecution `json:"steps"`
		}

		response := RunResponse{
			Run:   run,
			Steps: steps,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// PostRun triggers a new pipeline run
func PostRun(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req struct {
			ConfigPath string `json:"config_path"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		if req.ConfigPath == "" {
			http.Error(w, "config_path is required", http.StatusBadRequest)
			return
		}

		// Make path absolute if relative
		configPath := req.ConfigPath
		if !filepath.IsAbs(configPath) {
			cwd, err := os.Getwd()
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get working directory: %v", err), http.StatusInternalServerError)
				return
			}
			configPath = filepath.Join(cwd, configPath)
		}

		// Check if file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Config file not found: %s", configPath), http.StatusNotFound)
			return
		}

		// Run pipeline in background (minimal for now - no goroutine yet)
		log.Printf("🚀 Triggering pipeline: %s", configPath)

		result, err := runner.RunPipelineWithOptions(configPath, runner.RunPipelineOptions{
			Storage:          store,
			StreamToTerminal: false, // Don't stream when triggered via API
		})

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":  err.Error(),
				"run_id": result.RunID,
			})
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"run_id":  result.RunID,
			"status":  result.Status,
			"message": "Pipeline started successfully",
		})
	}
}

