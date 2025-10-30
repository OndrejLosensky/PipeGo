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

// GetRunStatus returns just the status of a run (lightweight for polling)
func GetRunStatus(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse run ID from URL: /api/runs/:id/status
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

		// Get only the run (no steps for lightweight response)
		run, err := store.GetRun(runID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Run not found: %v", err), http.StatusNotFound)
			return
		}

		// Return minimal status info
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     run.ID,
			"status": run.Status,
		})
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

		// Get optional part filter from query parameter
		partFilter := r.URL.Query().Get("part")

		// Run pipeline in background (async)
		if partFilter != "" {
			log.Printf("🚀 Triggering pipeline for project %s (part: %s): %s", projectName, partFilter, configPath)
		} else {
			log.Printf("🚀 Triggering pipeline for project %s: %s", projectName, configPath)
		}

		// Start pipeline in goroutine - runs completely async
		go func() {
			_, err := runner.RunPipelineWithOptions(configPath, runner.RunPipelineOptions{
				Storage:          store,
				StreamToTerminal: false,
				PartFilter:       partFilter,
			})

			if err != nil {
				log.Printf("❌ Pipeline execution failed for %s: %v", projectName, err)
			} else {
				log.Printf("✅ Pipeline completed successfully for %s", projectName)
			}
		}()

		// Return immediately - the run will be created in DB and polling will pick it up
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": fmt.Sprintf("Pipeline started for %s", projectName),
			"status":  "starting",
		})
	}
}

// GetProjectStats returns latest runs grouped by part for a project
func GetProjectStats(store *storage.Storage, projectsConfig *runner.ProjectsConfig, baseDir string) http.HandlerFunc {
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

		// Get latest 5 runs per part from database
		stats, err := store.GetLatestRunsByPart(projectName, 5)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get project stats: %v", err), http.StatusInternalServerError)
			return
		}

		// Load project config to get all defined parts
		project, err := projectsConfig.GetProject(projectName)
		if err == nil {
			configPath := project.GetPipegoPath(baseDir)
			cfg, err := runner.LoadConfig(configPath)
			if err == nil {
				// Get all parts from config
				allParts := cfg.GetAllParts()
				
				// Track which parts have runs
				partsWithRuns := make(map[string]bool)
				for _, stat := range stats {
					partsWithRuns[stat.Part] = true
				}
				
				// Add placeholder for parts without runs
				for partName := range allParts {
					if !partsWithRuns[partName] {
						// Add empty entry for parts with no runs
						stats = append(stats, storage.PartRunStats{
							Part:      partName,
							RunID:     0,
							Status:    "",
							StepCount: 0,
						})
					}
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

