package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/xonoxc/scopion/internal/app"
	"github.com/xonoxc/scopion/internal/benchmark"
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
		return app.StartServerWithConfig(ctx, port, app.ServerConfig{
			Mode: app.DEMO_MODE,
		})
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
				Analysis:  benchmark.DatabaseAnalysis{},
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

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		monitor.Stop()

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

	benchmarkCmd.AddCommand(benchStandardCmd)
	benchmarkCmd.AddCommand(benchStressCmd)
	benchmarkCmd.AddCommand(benchLimitsCmd)
	benchmarkCmd.AddCommand(benchMonitorCmd)

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(benchmarkCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
