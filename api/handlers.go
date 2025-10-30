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
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Method not allowed",
			})
			return
		}

		// Parse request body
		var req struct {
			ConfigPath string `json:"config_path"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("Invalid request: %v", err),
			})
			return
		}

		if req.ConfigPath == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "config_path is required",
			})
			return
		}

		// Make path absolute if relative
		configPath := req.ConfigPath
		if !filepath.IsAbs(configPath) {
			cwd, err := os.Getwd()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": fmt.Sprintf("Failed to get working directory: %v", err),
				})
				return
			}
			configPath = filepath.Join(cwd, configPath)
		}

		// Check if file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("Config file not found: %s", configPath),
			})
			return
		}

		// Run pipeline in background (minimal for now - no goroutine yet)
		log.Printf("🚀 Triggering pipeline: %s", configPath)

		result, err := runner.RunPipelineWithOptions(configPath, runner.RunPipelineOptions{
			Storage:          store,
			StreamToTerminal: false, // Don't stream when triggered via API
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":  err.Error(),
				"run_id": result.RunID,
			})
			return
		}

		// Return success response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"run_id":  result.RunID,
			"status":  result.Status,
			"message": "Pipeline started successfully",
		})
	}
}

// GetProjects returns all configured projects
func GetProjects(projectsConfig *runner.ProjectsConfig, baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Validate each project
		type ProjectResponse struct {
			runner.Project
			Valid bool   `json:"valid"`
			Error string `json:"error,omitempty"`
		}

		projects := make([]ProjectResponse, 0, len(projectsConfig.Projects))
		for _, project := range projectsConfig.Projects {
			pr := ProjectResponse{Project: project, Valid: true}
			if err := project.Validate(baseDir); err != nil {
				pr.Valid = false
				pr.Error = err.Error()
			}
			projects = append(projects, pr)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	}
}

// GetProjectRuns returns runs for a specific project
func GetProjectRuns(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse project name from URL: /api/projects/:name/runs
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		projectName := pathParts[2]

		// Get all runs and filter by project name
		runs, err := store.GetRuns(100)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get runs: %v", err), http.StatusInternalServerError)
			return
		}

		// Filter runs for this project
		projectRuns := make([]*storage.Run, 0)
		for _, run := range runs {
			if run.ProjectName == projectName {
				projectRuns = append(projectRuns, run)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projectRuns)
	}
}

// PostProjectRun triggers a pipeline run for a specific project
func PostProjectRun(store *storage.Storage, projectsConfig *runner.ProjectsConfig, baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Method not allowed",
			})
			return
		}

		// Parse project name from URL: /api/projects/:name/run
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Invalid path",
			})
			return
		}

		projectName := pathParts[2]

		// Get project
		project, err := projectsConfig.GetProject(projectName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("Project not found: %v", err),
			})
			return
		}

		// Validate project
		if err := project.Validate(baseDir); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("Invalid project: %v", err),
			})
			return
		}

		// Get pipego.yml path
		configPath := project.GetPipegoPath(baseDir)

		// Run pipeline in background
		log.Printf("🚀 Triggering pipeline for project %s: %s", projectName, configPath)

		result, err := runner.RunPipelineWithOptions(configPath, runner.RunPipelineOptions{
			Storage:          store,
			StreamToTerminal: false,
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":  err.Error(),
				"run_id": result.RunID,
			})
			return
		}

		// Return success response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"run_id":  result.RunID,
			"status":  result.Status,
			"message": fmt.Sprintf("Pipeline started successfully for %s", projectName),
		})
	}
}

// GetProjectStats returns latest runs grouped by part for a project
func GetProjectStats(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse project name from URL: /api/projects/:name/stats
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		projectName := pathParts[2]

		// Get latest 5 runs per part
		stats, err := store.GetLatestRunsByPart(projectName, 5)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get project stats: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

