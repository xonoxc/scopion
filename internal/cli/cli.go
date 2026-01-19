package cli

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"github.com/xonoxc/scopion/internal/api"
	"github.com/xonoxc/scopion/internal/demo"
	"github.com/xonoxc/scopion/internal/ingest"
	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/store"
	"github.com/xonoxc/scopion/ui"
)

const scorpionArt = `
   ___     ___
  /   \___/   \
  \___/   \___/
      \___/
     __| |__
    /  | |  \
   |   | |   |
   \   \_/   /
    \       /
     \     /
      \   /
       \ /
        V
   Scopion
Single-binary observability
`

// loggingMiddleware logs HTTP requests with method, path, status, and duration
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
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

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

type ServerConfig struct {
	DemoEnabled bool
}

func startServer(ctx context.Context, port string, enableDemo bool) error {
	return startServerWithConfig(ctx, port, ServerConfig{DemoEnabled: enableDemo})
}

func startServerWithConfig(ctx context.Context, port string, config ServerConfig) error {
	db, err := sql.Open("sqlite3", "./scopion.db")
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	store, err := store.New("./scopion.db")
	if err != nil {
		return err
	}
	defer store.Close()

	broadcaster := live.New()

	if config.DemoEnabled {
		log.Println("Demo mode enabled - generating sample telemetry data")
		demo.Start(store, broadcaster)
	}

	http.Handle("/api/live", loggingMiddleware(live.SSE(broadcaster)))
	http.Handle("/api/events", loggingMiddleware(api.EventsHandler(store)))
	http.Handle("/api/trace-events", loggingMiddleware(api.TraceEventsHandler(store)))
	http.Handle("/api/stats", loggingMiddleware(api.StatsHandler(store)))
	http.Handle("/api/errors-by-service", loggingMiddleware(api.ErrorsByServiceHandler(store)))
	http.Handle("/api/services", loggingMiddleware(api.ServicesHandler(store)))
	http.Handle("/api/traces", loggingMiddleware(api.TracesHandler(store)))
	http.Handle("/api/search", loggingMiddleware(api.SearchHandler(store)))
	http.Handle("/api/status", loggingMiddleware(api.StatusHandler(config.DemoEnabled)))
	http.Handle("/ingest", loggingMiddleware(ingest.Handler(store, broadcaster)))

	sub, err := fs.Sub(ui.FS, "dist")
	if err != nil {
		return err
	}

	http.Handle("/", loggingMiddleware(http.FileServer(http.FS(sub))))

	server := &http.Server{Addr: ":" + port, Handler: nil}

	// Channel to listen for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for shutdown signal or context cancellation
	select {
	case sig := <-shutdown:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down gracefully...")
	}

	// Create a deadline for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		if err == context.DeadlineExceeded {
			log.Println("Shutdown timeout exceeded, server may still be shutting down...")
			return nil // Don't return error for timeout, server will still shut down
		} else {
			log.Printf("Server shutdown error: %v", err)
			return err
		}
	}

	log.Println("Server stopped gracefully")
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "scopion",
	Short: "Scopion - Single-binary observability",
	Long: scorpionArt + `

Scopion is a single-binary observability tool that collects telemetry data and provides a web UI for monitoring.`,
}

var (
	port       string
	enableDemo bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Scopion server",
	Long:  `Start the Scopion server with telemetry collection and web UI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(scorpionArt)
		fmt.Println()
		ctx := context.Background()
		return startServerWithConfig(ctx, port, ServerConfig{DemoEnabled: enableDemo})
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of Scopion.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Scopion v1.0.0")
	},
}

func init() {
	startCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	startCmd.Flags().BoolVar(&enableDemo, "demo", true, "Enable demo data generation")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the CLI application
func Execute() error {
	return rootCmd.Execute()
}
