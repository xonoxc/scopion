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
	"github.com/xonoxc/scopion/internal/demo"
	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/pipeline"
	"github.com/xonoxc/scopion/internal/store/sqlite"
	"github.com/xonoxc/scopion/ui"
)

type ServerMode string

const (
	DEMO_MODE   ServerMode = "demo"
	NORMAL_MODE ServerMode = "normal"
)

type ServerConfig struct {
	Mode            ServerMode
	RetentionDays   int
	CleanupInterval time.Duration
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Mode:            NORMAL_MODE,
		RetentionDays:   7,
		CleanupInterval: 1 * time.Hour,
	}
}

func (s *ServerConfig) IsDemoMode() bool {
	return s.Mode == DEMO_MODE
}

func StartServer(ctx context.Context, port string, mode ServerMode) error {
	return StartServerWithConfig(ctx, port, ServerConfig{
		Mode: mode,
	})
}

func StartServerWithConfig(ctx context.Context, port string, config ServerConfig) error {
	db, err := sql.Open("sqlite3", "./scopion.db")
	if err != nil {
		return err
	}
	defer db.Close()

	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")
	db.Exec("PRAGMA foreign_keys=ON")

	store := sqlite.NewWithDB(db)

	if err := store.Migrate(); err != nil {
		return err
	}

	broadcaster := live.New()
	bp := pipeline.New(store, broadcaster)

	pipelineCtx, pipelineCancel := context.WithCancel(context.Background())
	defer pipelineCancel()
	go bp.Run(pipelineCtx)

	if config.RetentionDays > 0 {
		cleanupInterval := config.CleanupInterval
		if cleanupInterval <= 0 {
			cleanupInterval = 1 * time.Hour
		}

		go func() {
			ticker := time.NewTicker(cleanupInterval)
			defer ticker.Stop()
			for {
				select {
				case <-pipelineCtx.Done():
					return
				case <-ticker.C:
					deleted, err := store.DeleteEventsOlderThan(time.Duration(config.RetentionDays) * 24 * time.Hour)
					if err != nil {
						log.Printf("retention cleanup failed: %v", err)
					} else if deleted > 0 {
						log.Printf("retention cleanup: deleted %d events older than %d days", deleted, config.RetentionDays)
					}
				}
			}
		}()
	}

	if config.Mode == DEMO_MODE {
		log.Println("Demo mode enabled - generating sample telemetry data")
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		demo.Start(ctx, store, broadcaster)
	}

	router := NewAppRouter(store, broadcaster, bp, config)
	mux := router.Setup()

	sub, err := fs.Sub(ui.FS, "dist")
	if err != nil {
		return err
	}

	mux.Handle(
		"/", middleware.LoggingMiddleware(http.FileServer(http.FS(sub))),
	)

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

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
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
