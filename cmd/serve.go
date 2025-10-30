package cmd

import (
    "fmt"
    "html/template"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/joho/godotenv"
    "pipego/api"
    "pipego/runner/storage"
)

// Serve starts the HTTP server
func Serve() error {
	// Load .env file if it exists (ignore errors if it doesn't)
	_ = godotenv.Load()

	// Get port from environment variable or use default
	port := getEnv("PORT", "8080")

	// Get database path
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

    // Parse templates - each page is now self-contained
    indexTmpl, err := template.ParseFiles(
        filepath.Join(cwd, "templates", "index.html"),
    )
    if err != nil {
        log.Fatalf("Failed to load index template: %v", err)
    }

    runTmpl, err := template.ParseFiles(
        filepath.Join(cwd, "templates", "run.html"),
    )
    if err != nil {
        log.Fatalf("Failed to load run template: %v", err)
    }

    // Setup HTTP routes
    mux := http.NewServeMux()

    // Dashboard: list runs
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
        }
        runs, err := store.GetRuns(100)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to load runs: %v", err), http.StatusInternalServerError)
            return
        }
        data := struct {
            Runs []*storage.Run
        }{Runs: runs}
        if err := indexTmpl.Execute(w, data); err != nil {
            log.Printf("template render error (index): %v", err)
            return
        }
    })

    // Run details: /runs/:id
    mux.HandleFunc("/runs/", func(w http.ResponseWriter, r *http.Request) {
        parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
        if len(parts) != 2 || parts[0] != "runs" {
            http.NotFound(w, r)
            return
        }
        id, err := strconv.Atoi(parts[1])
        if err != nil {
            http.Error(w, "Invalid run id", http.StatusBadRequest)
            return
        }
        run, err := store.GetRun(id)
        if err != nil {
            http.Error(w, "Run not found", http.StatusNotFound)
            return
        }
        steps, err := store.GetStepExecutions(id)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to load steps: %v", err), http.StatusInternalServerError)
            return
        }
        data := struct {
            Run   *storage.Run
            Steps []*storage.StepExecution
        }{Run: run, Steps: steps}
        if err := runTmpl.Execute(w, data); err != nil {
            log.Printf("template render error (run): %v", err)
            return
        }
    })

	// API endpoints
	mux.HandleFunc("/api/runs", api.GetRuns(store))
	mux.HandleFunc("/api/runs/", api.GetRun(store)) 
	mux.HandleFunc("/api/run", api.PostRun(store))

	// Start HTTP server
	serverAddr := ":" + port
	log.Printf("🚀 Starting PipeGo server on port %s...", port)
	log.Printf("📊 Dashboard: http://localhost:%s", port)
	
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

