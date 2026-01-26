package app

import (
	"context"
	"database/sql"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pressly/goose/v3"

	"github.com/xonoxc/scopion/internal/api/middleware"
	"github.com/xonoxc/scopion/internal/app/appcontext"
	"github.com/xonoxc/scopion/internal/demo"
	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/store/sqlite"
	"github.com/xonoxc/scopion/ui"

	appstorage "github.com/xonoxc/scopion/internal/store"
)

/*
* TYPE for the modes
 */
type ServerMode string

/*
* MODES the application can be started in
 */
const (
	DEMO_MODE   ServerMode = "demo"
	NORMAL_MODE ServerMode = "normal"
)

/*
* API server config
* DEMO_MODE: enables demo mode with sample telemetry data
* NORMAL_MODE: standard operation mode
 */
type ServerConfig struct {
	Mode ServerMode
}

func (s *ServerConfig) IsDemoMode() bool {
	return s.Mode == DEMO_MODE
}

/*
* StartServer starts the API server on the specified port with the given mode.
 */
func StartServer(ctx context.Context, port string, mode ServerMode) error {
	return StartServerWithConfig(ctx, port, ServerConfig{
		Mode: mode,
	})
}

/*
* stating server with config
 */
func StartServerWithConfig(ctx context.Context, port string, config ServerConfig) error {
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

	store, err := sqlite.New("./scopion.db")
	if err != nil {
		return err
	}
	defer store.Close()

	as := appcontext.NewAtomicAppState(store, appstorage.DUAL_WRITE)

	broadcaster := live.New()

	if config.Mode == DEMO_MODE {
		log.Println("Demo mode enabled - generating sample telemetry data")
		demo.Start(store, broadcaster)
	}

	router := NewAppRouter(as, broadcaster, config)
	router.Setup()

	sub, err := fs.Sub(ui.FS, "dist")
	if err != nil {
		return err
	}

	http.Handle("/", middleware.LoggingMiddleware(http.FileServer(http.FS(sub))))

	server := &http.Server{Addr: ":" + port, Handler: nil}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	select {
	case sig := <-shutdown:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down gracefully...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		if err == context.DeadlineExceeded {
			log.Println("Shutdown timeout exceeded, server may still be shutting down...")
			return nil
		} else {
			log.Printf("Server shutdown error: %v", err)
			return err
		}
	}

	log.Println("Server stopped gracefully")
	return nil
}
