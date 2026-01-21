package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/natefinch/lumberjack.v2"
	"server-dashboard/internal/config"
	"server-dashboard/internal/handlers"
	"server-dashboard/internal/services"
)

//go:embed web/templates
var templatesFS embed.FS

//go:embed web/static
var staticFS embed.FS

// Version information - set via ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Read version from VERSION file if not set via ldflags
	if Version == "dev" {
		if data, err := ioutil.ReadFile("VERSION"); err == nil {
			Version = strings.TrimSpace(string(data))
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Set up logging
	if err := setupLogging(cfg); err != nil {
		log.Fatalf("Error setting up logging: %v", err)
	}

	// Log configuration (safe - don't log passwords in production)
	log.Printf("Server Dashboard %s", Version)
	log.Printf("Starting in %s environment", cfg.Environment)
	log.Printf("Listening on %s", cfg.ServerAddress)
	log.Printf("Log directory: %s", cfg.Logging.Directory)
	log.Printf("Log level: %s", cfg.Logging.Level)
	if cfg.TLS.Enabled {
		log.Printf("TLS enabled")
	}
	if cfg.Auth.Enabled {
		log.Printf("Authentication enabled for user: %s", cfg.Auth.Username)
	}

	// Initialize service cache from config
	services.InitializeCache(cfg)

	// Create function map for templates
	funcMap := template.FuncMap{
		"currentYear": func() int {
			return time.Now().Year()
		},
		"appVersion": func() string {
			return Version
		},
		"buildInfo": func() string {
			if BuildTime != "unknown" {
				return fmt.Sprintf("%s (built %s)", Version, BuildTime)
			}
			return Version
		},
		"getServerCount": func() int {
			servers, _ := services.GetAllServers()
			return len(servers)
		},
		"getVMCount": func() int {
			vms, _ := services.GetAllVMs()
			return len(vms)
		},
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Load templates from embedded filesystem
	templates := template.New("").Funcs(funcMap)
	templateFiles, err := fs.Glob(templatesFS, "web/templates/*.html")
	if err != nil {
		log.Fatalf("Error finding templates: %v", err)
	}
	for _, tmplFile := range templateFiles {
		tmplContent, err := templatesFS.ReadFile(tmplFile)
		if err != nil {
			log.Fatalf("Error reading template %s: %v", tmplFile, err)
		}
		_, err = templates.New(filepath.Base(tmplFile)).Parse(string(tmplContent))
		if err != nil {
			log.Fatalf("Error parsing template %s: %v", tmplFile, err)
		}
	}

	// Initialize router
	r := mux.NewRouter()

	// Middleware - add logging and security headers
	r.Use(loggingMiddleware)
	r.Use(securityHeadersMiddleware)

	// Serve static files from embedded filesystem
	staticFilesFS, err := fs.Sub(staticFS, "web/static")
	if err != nil {
		log.Fatalf("Error creating static files sub-filesystem: %v", err)
	}
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFilesFS))))

	// Health check endpoint
	r.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Set up routes
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.DashboardHandlerWithTemplates(w, r, templates)
	}).Methods("GET")

	r.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServerHandlerWithTemplates(w, r, templates)
	}).Methods("GET")

	r.HandleFunc("/servers/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServerDetailHandlerWithTemplates(w, r, templates)
	}).Methods("GET")

	r.HandleFunc("/vms", func(w http.ResponseWriter, r *http.Request) {
		handlers.VMHandlerWithTemplates(w, r, templates)
	}).Methods("GET")

	r.HandleFunc("/vms/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.VMDetailHandlerWithTemplates(w, r, templates)
	}).Methods("GET")

	// Monitoring control API endpoints
	r.HandleFunc("/api/monitoring/status", handlers.GetMonitoringStatus).Methods("GET")
	r.HandleFunc("/api/monitoring/start", handlers.StartMonitoring).Methods("POST")
	r.HandleFunc("/api/monitoring/stop", handlers.StopMonitoring).Methods("POST")
	r.HandleFunc("/api/monitoring/restart", handlers.RestartMonitoring).Methods("POST")

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		if cfg.TLS.Enabled {
			log.Printf("Starting HTTPS server on %s", cfg.ServerAddress)
			log.Printf("TLS certificate: %s", cfg.TLS.CertFile)
			if err := server.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTPS server error: %v", err)
				serverErrors <- err
			}
		} else {
			log.Printf("Starting HTTP server on %s", cfg.ServerAddress)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTP server error: %v", err)
				serverErrors <- err
			}
		}
	}()

	log.Printf("Server started successfully - ready to accept connections")

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Starting graceful shutdown...", sig)

		// Create a context with a timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Printf("Shutting down server with 30 second timeout...")

		// Shutdown the server gracefully
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Shutdown error: %v", err)
		}

		log.Printf("Server shut down successfully")
	}
}

// setupLogging configures logging to both file and console with rotation
func setupLogging(cfg *config.Config) error {
	// Determine log directory based on environment
	logDir := cfg.Logging.Directory
	
	// In development mode, use relative path from current directory
	// In production mode, use absolute path (e.g., /var/log/server-dashboard)
	if cfg.Environment == "development" {
		// Use relative path from current directory
		if !filepath.IsAbs(logDir) {
			// Already relative, use as-is
		} else {
			// Convert to relative for development
			logDir = "./logs"
		}
	}
	
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// If we can't create the configured directory, fall back to current directory
		log.Printf("Warning: Cannot create log directory %s: %v. Using current directory.", logDir, err)
		logDir = "."
	}

	logFile := filepath.Join(logDir, "server-dashboard.log")

	// Set up file rotation
	fileLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    cfg.Logging.MaxSizeMB,    // megabytes
		MaxBackups: cfg.Logging.MaxBackups,   // number of backups
		MaxAge:     cfg.Logging.MaxAgeDays,   // days
		Compress:   cfg.Logging.Compress,     // compress old files
		LocalTime:  true,                     // use local time for filenames
	}

	// Create multi-writer for both file and console
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)

	// Set log output to both file and console
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("Logging initialized - writing to %s", logFile)
	return nil
}

// loggingMiddleware logs HTTP requests with timing information
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log the request
		log.Printf("Started %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(rw, r)

		// Log the response
		duration := time.Since(start)
		log.Printf("Completed %s %s with %d in %v", r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// securityHeadersMiddleware adds security headers to all responses
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// healthCheckHandler returns the health status of the application
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}