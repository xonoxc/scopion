package cli

import (
	"context"
	"database/sql"
	"encoding/json"
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
	"github.com/xonoxc/scopion/internal/benchmark"
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
	http.Handle("/api/throughput", loggingMiddleware(api.ThroughputHandler(store)))
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
	port          string
	enableDemo    bool
	benchWorkers  int
	benchDuration time.Duration
	benchRate     int
	benchOutput   string
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

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run database benchmarks and performance tests",
	Long:  `Run comprehensive benchmarks to test SQLite database limits and performance.`,
}

var benchStandardCmd = &cobra.Command{
	Use:   "standard",
	Short: "Run standard load benchmark",
	Long:  `Run a standard load benchmark to test database performance under normal conditions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := benchmark.BenchmarkConfig{
			DatabasePath: "./scopion.db",
			Duration:     benchDuration,
			Workers:      benchWorkers,
			EventRate:    benchRate,
			BatchSize:    10,
		}

		runner := benchmark.NewBenchmarkRunner(config)
		result, err := runner.RunStandardBenchmark()
		if err != nil {
			return fmt.Errorf("benchmark failed: %w", err)
		}

		fmt.Println("=== Standard Benchmark Results ===")
		fmt.Printf("Duration: %v\n", result.Duration)
		fmt.Printf("Total Events: %d\n", result.TotalEvents)
		fmt.Printf("Events/Second: %.2f\n", result.EventsPerSecond)
		fmt.Printf("Avg Latency: %v\n", result.AvgLatency)
		fmt.Printf("Memory Usage: %.2f MB\n", result.MemoryUsageMB)
		fmt.Printf("Database Size: %.2f MB\n", result.DatabaseSizeMB)
		fmt.Printf("Errors: %d\n", result.ErrorCount)

		if benchOutput != "" {
			return saveBenchmarkResult(result, benchOutput)
		}

		return nil
	},
}

var benchStressCmd = &cobra.Command{
	Use:   "stress",
	Short: "Run progressive stress test",
	Long:  `Run a progressive stress test that gradually increases load to find breaking points.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := benchmark.BenchmarkConfig{
			DatabasePath: "./scopion.db",
			Duration:     benchDuration,
			Workers:      benchWorkers,
			EventRate:    benchRate,
			BatchSize:    10,
		}

		runner := benchmark.NewBenchmarkRunner(config)
		results, err := runner.RunStressTest()
		if err != nil {
			return fmt.Errorf("stress test failed: %w", err)
		}

		fmt.Println("=== Stress Test Results ===")
		for i, result := range results {
			fmt.Printf("\nPhase %d:\n", i+1)
			fmt.Printf("  Events/Second: %.2f\n", result.EventsPerSecond)
			fmt.Printf("  Memory Usage: %.2f MB\n", result.MemoryUsageMB)
			fmt.Printf("  Errors: %d\n", result.ErrorCount)
		}

		// Analyze results
		analyzer := benchmark.NewPerformanceAnalyzer(results)
		recommendation := analyzer.Analyze()

		fmt.Println("\n=== Performance Analysis ===")
		fmt.Printf("Best Performance: %.2f events/sec\n", recommendation.BestPerformance.EventsPerSecond)
		fmt.Printf("Worst Performance: %.2f events/sec\n", recommendation.WorstPerformance.EventsPerSecond)
		fmt.Printf("Average Performance: %.2f events/sec\n", recommendation.AveragePerformance.EventsPerSecond)
		fmt.Printf("Recommendation: %s\n", recommendation.Recommendation)
		fmt.Printf("Confidence: %s\n", recommendation.Confidence)

		if benchOutput != "" {
			report := &benchmark.DatabaseLimitsReport{
				StartTime: results[0].StartTime,
				EndTime:   results[len(results)-1].EndTime,
				Tests:     make([]benchmark.TestResult, len(results)),
				Analysis:  benchmark.DatabaseAnalysis{}, // Would need to generate
			}

			for i, result := range results {
				report.Tests[i] = benchmark.TestResult{
					Name:    fmt.Sprintf("Stress Phase %d", i+1),
					Result:  &result,
					Success: result.ErrorCount == 0,
				}
			}

			return report.SaveReport(benchOutput)
		}

		return nil
	},
}

var benchLimitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Test database limits and find breaking points",
	Long:  `Run comprehensive tests to find SQLite database limits and provide migration recommendations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := benchmark.BenchmarkConfig{
			DatabasePath: "./scopion.db",
			Duration:     benchDuration,
			Workers:      benchWorkers,
			EventRate:    benchRate,
			BatchSize:    10,
		}

		runner := benchmark.NewBenchmarkRunner(config)
		report, err := runner.RunDatabaseLimitsTest()
		if err != nil {
			return fmt.Errorf("limits test failed: %w", err)
		}

		report.PrintReport()

		if benchOutput != "" {
			return report.SaveReport(benchOutput)
		}

		return nil
	},
}

var benchMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Run continuous performance monitoring",
	Long:  `Start continuous monitoring of database performance with periodic benchmarks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := benchmark.BenchmarkConfig{
			DatabasePath: "./scopion.db",
			Duration:     benchDuration,
			Workers:      benchWorkers,
			EventRate:    benchRate,
			BatchSize:    10,
		}

		monitor := benchmark.NewContinuousMonitor(config, 1*time.Minute)

		fmt.Println("Starting continuous performance monitoring...")
		fmt.Println("Press Ctrl+C to stop")

		if err := monitor.Start(); err != nil {
			return fmt.Errorf("failed to start monitor: %w", err)
		}

		// Wait for interrupt
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		monitor.Stop()

		// Generate final report
		results := monitor.GetResults()
		if len(results) > 0 {
			trend := monitor.GenerateTrendReport()
			if trend != nil {
				fmt.Println("\n=== Performance Trend Report ===")
				fmt.Printf("Samples: %d\n", trend.TotalSamples)
				fmt.Printf("Average EPS: %.2f\n", trend.AverageEPS)
				fmt.Printf("Min/Max EPS: %.2f / %.2f\n", trend.MinEPS, trend.MaxEPS)
				fmt.Printf("Trend: %s\n", trend.TrendDirection)
			}
		}

		return nil
	},
}

func saveBenchmarkResult(result *benchmark.BenchmarkResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func init() {
	startCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	startCmd.Flags().BoolVar(&enableDemo, "demo", true, "Enable demo data generation")

	// Benchmark command flags
	benchStandardCmd.Flags().IntVarP(&benchWorkers, "workers", "w", 10, "Number of concurrent workers")
	benchStandardCmd.Flags().DurationVarP(&benchDuration, "duration", "d", 30*time.Second, "Benchmark duration")
	benchStandardCmd.Flags().IntVarP(&benchRate, "rate", "r", 0, "Target events per second (0 for unlimited)")
	benchStandardCmd.Flags().StringVarP(&benchOutput, "output", "o", "", "Output file for results (JSON)")

	benchStressCmd.Flags().IntVarP(&benchWorkers, "workers", "w", 10, "Starting number of concurrent workers")
	benchStressCmd.Flags().DurationVarP(&benchDuration, "duration", "d", 30*time.Second, "Duration per stress phase")
	benchStressCmd.Flags().IntVarP(&benchRate, "rate", "r", 0, "Starting target events per second")
	benchStressCmd.Flags().StringVarP(&benchOutput, "output", "o", "", "Output file for results (JSON)")

	benchLimitsCmd.Flags().IntVarP(&benchWorkers, "workers", "w", 10, "Starting number of concurrent workers")
	benchLimitsCmd.Flags().DurationVarP(&benchDuration, "duration", "d", 30*time.Second, "Duration per test phase")
	benchLimitsCmd.Flags().IntVarP(&benchRate, "rate", "r", 0, "Starting target events per second")
	benchLimitsCmd.Flags().StringVarP(&benchOutput, "output", "o", "", "Output file for results (JSON)")

	benchMonitorCmd.Flags().IntVarP(&benchWorkers, "workers", "w", 5, "Number of workers for monitoring tests")
	benchMonitorCmd.Flags().DurationVarP(&benchDuration, "duration", "d", 10*time.Second, "Duration per monitoring test")
	benchMonitorCmd.Flags().IntVarP(&benchRate, "rate", "r", 100, "Target events per second for monitoring")

	// Add benchmark subcommands
	benchmarkCmd.AddCommand(benchStandardCmd)
	benchmarkCmd.AddCommand(benchStressCmd)
	benchmarkCmd.AddCommand(benchLimitsCmd)
	benchmarkCmd.AddCommand(benchMonitorCmd)

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(benchmarkCmd)
}

// Execute runs the CLI application
func Execute() error {
	return rootCmd.Execute()
}
