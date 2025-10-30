package cmd

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"

    "github.com/joho/godotenv"
    "pipego/api"
    "pipego/runner"
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

	// Load projects configuration
	projectsPath := filepath.Join(cwd, "projects.yml")
	projectsConfig, err := runner.LoadProjects(projectsPath)
	if err != nil {
		log.Printf("Warning: Failed to load projects config: %v", err)
		projectsConfig = &runner.ProjectsConfig{Projects: []runner.Project{}}
	} else {
		log.Printf("📁 Loaded %d project(s)", len(projectsConfig.Projects))
	}

    // Setup HTTP routes
    mux := http.NewServeMux()

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Serve React
	webDir := filepath.Join(cwd, "web", "dist")
	fileServer := http.FileServer(http.Dir(webDir))
	
	// Serve static files from React build
	mux.Handle("/assets/", fileServer)
	
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		
		indexPath := filepath.Join(webDir, "index.html")
		http.ServeFile(w, r, indexPath)
	})

	// API endpoints
	mux.HandleFunc("/api/runs", api.GetRuns(store))
	mux.HandleFunc("/api/runs/", api.GetRun(store)) 
	mux.HandleFunc("/api/run", api.PostRun(store))
	
	mux.HandleFunc("/api/projects", api.GetProjects(projectsConfig, cwd))
	mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/runs") {
			api.GetProjectRuns(store)(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/run") {
			api.PostProjectRun(store, projectsConfig, cwd)(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/stats") {
			api.GetProjectStats(store)(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// Start HTTP server with CORS
	serverAddr := ":" + port
	log.Printf("🚀 Starting PipeGo server on port %s...", port)
	log.Printf("📊 Dashboard: http://localhost:%s", port)
	
	if err := http.ListenAndServe(serverAddr, corsMiddleware(mux)); err != nil {
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

