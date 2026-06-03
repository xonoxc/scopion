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

	_ "github.com/mattn/go-sqlite3"

	"github.com/xonoxc/scopion/internal/api/middleware"
	"github.com/xonoxc/scopion/internal/app/appcontext"
	"github.com/xonoxc/scopion/internal/demo"
	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/store/migrations"
	"github.com/xonoxc/scopion/internal/store/sqlite"
	"github.com/xonoxc/scopion/ui"

	appstorage "github.com/xonoxc/scopion/internal/store"
	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
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

	migrator := migrations.NewMigrator(migrations.GetAll())
	if err := migrator.Migrate(db, migrateable.SQLITE); err != nil {
		return err
	}

	store := sqlite.NewWithDB(db)

	as := appcontext.NewAtomicAppState(store, appstorage.SINGLE_PRIMARY)

	broadcaster := live.New()

	if config.Mode == DEMO_MODE {
		log.Println("Demo mode enabled - generating sample telemetry data")
		demo.Start(as, broadcaster)
	}

	router := NewAppRouter(as, broadcaster, config)
	mux := router.Setup()

	sub, err := fs.Sub(ui.FS, "dist")
	if err != nil {
		return err
	}

	mux.Handle("/", middleware.LoggingMiddleware(http.FileServer(http.FS(sub))))

	server := &http.Server{Addr: ":" + port, Handler: mux}

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
