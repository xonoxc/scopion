package benchmark

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"time"
)

// BenchmarkRunner coordinates benchmark execution and analysis
type BenchmarkRunner struct {
	config BenchmarkConfig
}

func NewBenchmarkRunner(config BenchmarkConfig) *BenchmarkRunner {
	return &BenchmarkRunner{config: config}
}

// RunStandardBenchmark runs a standard load test
func (br *BenchmarkRunner) RunStandardBenchmark() (*BenchmarkResult, error) {
	log.Printf("Running standard benchmark: %d workers, %v duration",
		br.config.Workers, br.config.Duration)

	generator, err := NewLoadGenerator(br.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create load generator: %w", err)
	}

	return generator.Run(nil)
}

// RunStressTest runs a progressive stress test to find breaking points
func (br *BenchmarkRunner) RunStressTest() ([]BenchmarkResult, error) {
	log.Println("Running progressive stress test...")

	generator, err := NewLoadGenerator(br.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create load generator: %w", err)
	}

	return generator.StressTest(nil, 100, 10*time.Minute)
}

// RunDatabaseLimitsTest runs extreme tests to find SQLite limits
func (br *BenchmarkRunner) RunDatabaseLimitsTest() (*DatabaseLimitsReport, error) {
	log.Println("Running database limits test...")

	stressTest, err := NewDatabaseStressTest(br.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create stress test: %w", err)
	}
	defer stressTest.Cleanup()

	report := &DatabaseLimitsReport{
		StartTime: time.Now(),
		Tests:     make([]TestResult, 0),
	}

	// Test 1: Memory exhaustion
	log.Println("Test 1: Memory exhaustion...")
	result, err := stressTest.RunMemoryExhaustionTest(nil)
	if err != nil {
		log.Printf("Memory exhaustion test failed: %v", err)
	}
	report.Tests = append(report.Tests, TestResult{
		Name:    "Memory Exhaustion",
		Result:  result,
		Error:   err,
		Success: err == nil,
	})

	// Test 2: Concurrent writes
	log.Println("Test 2: Concurrent writes...")
	for _, concurrency := range []int{10, 50, 100, 200} {
		result, err := stressTest.RunConcurrentWriteTest(nil, concurrency)
		report.Tests = append(report.Tests, TestResult{
			Name:    fmt.Sprintf("Concurrent Writes (%d workers)", concurrency),
			Result:  result,
			Error:   err,
			Success: err == nil,
		})
		if err != nil {
			log.Printf("Concurrent write test (%d) failed: %v", concurrency, err)
			break // Stop at first failure
		}
	}

	// Test 3: Large transactions
	log.Println("Test 3: Large transactions...")
	result, err = stressTest.RunLargeTransactionTest(nil)
	report.Tests = append(report.Tests, TestResult{
		Name:    "Large Transactions",
		Result:  result,
		Error:   err,
		Success: err == nil,
	})

	// Test 4: Read-write contention
	log.Println("Test 4: Read-write contention...")
	result, err = stressTest.RunReadWriteContentionTest(nil)
	report.Tests = append(report.Tests, TestResult{
		Name:    "Read-Write Contention",
		Result:  result,
		Error:   err,
		Success: err == nil,
	})

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)
	report.GenerateAnalysis()

	return report, nil
}

// DatabaseLimitsReport contains results from database limits testing
type DatabaseLimitsReport struct {
	StartTime time.Time        `json:"start_time"`
	EndTime   time.Time        `json:"end_time"`
	Duration  time.Duration    `json:"duration"`
	Tests     []TestResult     `json:"tests"`
	Analysis  DatabaseAnalysis `json:"analysis"`
}

type TestResult struct {
	Name    string           `json:"name"`
	Result  *BenchmarkResult `json:"result,omitempty"`
	Error   error            `json:"error,omitempty"`
	Success bool             `json:"success"`
}

type DatabaseAnalysis struct {
	MaxSafeConcurrency     int     `json:"max_safe_concurrency"`
	MaxEventsPerSecond     float64 `json:"max_events_per_second"`
	RecommendedSwitchPoint string  `json:"recommended_switch_point"`
	PerformanceDegradation string  `json:"performance_degradation"`
	RiskAssessment         string  `json:"risk_assessment"`
}

func (dlr *DatabaseLimitsReport) GenerateAnalysis() {
	analysis := DatabaseAnalysis{
		MaxSafeConcurrency: 10, // Default safe value
	}

	var maxEPS float64
	var successfulTests int

	for _, test := range dlr.Tests {
		if test.Success && test.Result != nil {
			successfulTests++
			if test.Result.EventsPerSecond > maxEPS {
				maxEPS = test.Result.EventsPerSecond
			}

			// Extract concurrency from test name for concurrent write tests
			if test.Name == "Concurrent Writes (10 workers)" && test.Success {
				analysis.MaxSafeConcurrency = 10
			} else if test.Name == "Concurrent Writes (50 workers)" && test.Success {
				analysis.MaxSafeConcurrency = 50
			} else if test.Name == "Concurrent Writes (100 workers)" && test.Success {
				analysis.MaxSafeConcurrency = 100
			} else if test.Name == "Concurrent Writes (200 workers)" && test.Success {
				analysis.MaxSafeConcurrency = 200
			}
		}
	}

	analysis.MaxEventsPerSecond = maxEPS

	// Generate recommendations
	if analysis.MaxEventsPerSecond < 1000 {
		analysis.RecommendedSwitchPoint = "Consider switching to PostgreSQL or MySQL for better performance"
		analysis.PerformanceDegradation = "High - SQLite cannot handle moderate load efficiently"
		analysis.RiskAssessment = "High risk for production use under normal load"
	} else if analysis.MaxEventsPerSecond < 5000 {
		analysis.RecommendedSwitchPoint = "SQLite acceptable for light to moderate load, monitor closely"
		analysis.PerformanceDegradation = "Medium - SQLite works but may struggle under peak load"
		analysis.RiskAssessment = "Medium risk, acceptable for development/testing"
	} else {
		analysis.RecommendedSwitchPoint = "SQLite performing well, continue monitoring"
		analysis.PerformanceDegradation = "Low - SQLite handling load effectively"
		analysis.RiskAssessment = "Low risk for current workload"
	}

	dlr.Analysis = analysis
}

// SaveReport saves the report to a JSON file
func (dlr *DatabaseLimitsReport) SaveReport(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(dlr)
}

// PrintReport prints a human-readable version of the report
func (dlr *DatabaseLimitsReport) PrintReport() {
	fmt.Println("=== SQLite Database Limits Report ===")
	fmt.Printf("Test Duration: %v\n", dlr.Duration)
	fmt.Printf("Tests Run: %d\n", len(dlr.Tests))
	fmt.Println()

	fmt.Println("Test Results:")
	for _, test := range dlr.Tests {
		status := "✓ PASS"
		if !test.Success {
			status = "✗ FAIL"
		}
		fmt.Printf("  %s: %s\n", test.Name, status)
		if test.Success && test.Result != nil {
			fmt.Printf("    Events/sec: %.2f, Memory: %.2f MB, DB Size: %.2f MB\n",
				test.Result.EventsPerSecond, test.Result.MemoryUsageMB, test.Result.DatabaseSizeMB)
		}
		if test.Error != nil {
			fmt.Printf("    Error: %v\n", test.Error)
		}
	}
	fmt.Println()

	fmt.Println("Analysis:")
	fmt.Printf("  Max Safe Concurrency: %d workers\n", dlr.Analysis.MaxSafeConcurrency)
	fmt.Printf("  Max Events/Second: %.2f\n", dlr.Analysis.MaxEventsPerSecond)
	fmt.Println()
	fmt.Println("Recommendations:")
	fmt.Printf("  Switch Point: %s\n", dlr.Analysis.RecommendedSwitchPoint)
	fmt.Printf("  Performance: %s\n", dlr.Analysis.PerformanceDegradation)
	fmt.Printf("  Risk: %s\n", dlr.Analysis.RiskAssessment)
}

// ContinuousMonitor runs ongoing performance monitoring
type ContinuousMonitor struct {
	config    BenchmarkConfig
	interval  time.Duration
	results   []BenchmarkResult
	stopChan  chan struct{}
	isRunning bool
}

func NewContinuousMonitor(config BenchmarkConfig, interval time.Duration) *ContinuousMonitor {
	return &ContinuousMonitor{
		config:    config,
		interval:  interval,
		results:   make([]BenchmarkResult, 0),
		stopChan:  make(chan struct{}),
		isRunning: false,
	}
}

func (cm *ContinuousMonitor) Start() error {
	if cm.isRunning {
		return fmt.Errorf("monitor already running")
	}

	cm.isRunning = true
	log.Printf("Starting continuous monitoring (interval: %v)", cm.interval)

	go cm.monitorLoop()

	return nil
}

func (cm *ContinuousMonitor) Stop() {
	if !cm.isRunning {
		return
	}

	log.Println("Stopping continuous monitoring...")
	close(cm.stopChan)
	cm.isRunning = false
}

func (cm *ContinuousMonitor) monitorLoop() {
	ticker := time.NewTicker(cm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-cm.stopChan:
			return
		case <-ticker.C:
			cm.runQuickTest()
		}
	}
}

func (cm *ContinuousMonitor) runQuickTest() {
	// Run a short benchmark to check current performance
	config := cm.config
	config.Duration = 10 * time.Second // Quick test
	config.Workers = 5
	config.BatchSize = 10

	runner := NewBenchmarkRunner(config)
	result, err := runner.RunStandardBenchmark()

	if err != nil {
		log.Printf("Continuous monitor test failed: %v", err)
		return
	}

	cm.results = append(cm.results, *result)

	// Log current performance
	log.Printf("Performance check - Events/sec: %.2f, Memory: %.2f MB, DB Size: %.2f MB",
		result.EventsPerSecond, result.MemoryUsageMB, result.DatabaseSizeMB)

	// Alert if performance drops significantly
	if len(cm.results) > 1 {
		previous := cm.results[len(cm.results)-2]
		degradation := (previous.EventsPerSecond - result.EventsPerSecond) / previous.EventsPerSecond * 100

		if degradation > 20 { // 20% degradation
			log.Printf("⚠️  PERFORMANCE DEGRADATION: %.1f%% drop in events/sec (%.2f -> %.2f)",
				degradation, previous.EventsPerSecond, result.EventsPerSecond)
		}
	}
}

func (cm *ContinuousMonitor) GetResults() []BenchmarkResult {
	return cm.results
}

func (cm *ContinuousMonitor) GenerateTrendReport() *TrendReport {
	if len(cm.results) < 2 {
		return nil
	}

	report := &TrendReport{
		StartTime:      cm.results[0].StartTime,
		EndTime:        cm.results[len(cm.results)-1].EndTime,
		TotalSamples:   len(cm.results),
		AverageEPS:     0,
		MinEPS:         cm.results[0].EventsPerSecond,
		MaxEPS:         cm.results[0].EventsPerSecond,
		TrendDirection: "stable",
	}

	var totalEPS float64
	for _, result := range cm.results {
		totalEPS += result.EventsPerSecond
		if result.EventsPerSecond < report.MinEPS {
			report.MinEPS = result.EventsPerSecond
		}
		if result.EventsPerSecond > report.MaxEPS {
			report.MaxEPS = result.EventsPerSecond
		}
	}

	report.AverageEPS = totalEPS / float64(len(cm.results))

	// Determine trend
	firstHalf := cm.results[:len(cm.results)/2]
	secondHalf := cm.results[len(cm.results)/2:]

	var firstAvg, secondAvg float64
	for _, r := range firstHalf {
		firstAvg += r.EventsPerSecond
	}
	for _, r := range secondHalf {
		secondAvg += r.EventsPerSecond
	}
	firstAvg /= float64(len(firstHalf))
	secondAvg /= float64(len(secondHalf))

	if secondAvg > firstAvg*1.05 {
		report.TrendDirection = "improving"
	} else if secondAvg < firstAvg*0.95 {
		report.TrendDirection = "degrading"
	}

	return report
}

type TrendReport struct {
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	TotalSamples   int       `json:"total_samples"`
	AverageEPS     float64   `json:"average_eps"`
	MinEPS         float64   `json:"min_eps"`
	MaxEPS         float64   `json:"max_eps"`
	TrendDirection string    `json:"trend_direction"`
}

// PerformanceAnalyzer analyzes benchmark results to provide recommendations
type PerformanceAnalyzer struct {
	results []BenchmarkResult
}

func NewPerformanceAnalyzer(results []BenchmarkResult) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{results: results}
}

func (pa *PerformanceAnalyzer) Analyze() *PerformanceRecommendation {
	if len(pa.results) == 0 {
		return &PerformanceRecommendation{
			Recommendation: "No data available for analysis",
			Confidence:     "N/A",
		}
	}

	// Sort by events per second
	sorted := make([]BenchmarkResult, len(pa.results))
	copy(sorted, pa.results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].EventsPerSecond > sorted[j].EventsPerSecond
	})

	best := sorted[0]
	worst := sorted[len(sorted)-1]

	rec := &PerformanceRecommendation{
		BestPerformance:    best,
		WorstPerformance:   worst,
		AveragePerformance: pa.calculateAverage(),
	}

	// Analyze patterns
	pa.analyzePatterns(rec)

	return rec
}

func (pa *PerformanceAnalyzer) calculateAverage() BenchmarkResult {
	if len(pa.results) == 0 {
		return BenchmarkResult{}
	}

	avg := BenchmarkResult{}
	for _, result := range pa.results {
		avg.TotalEvents += result.TotalEvents
		avg.EventsPerSecond += result.EventsPerSecond
		avg.MemoryUsageMB += result.MemoryUsageMB
		avg.DatabaseSizeMB += result.DatabaseSizeMB
		avg.ErrorCount += result.ErrorCount
	}

	count := int64(len(pa.results))
	avg.TotalEvents /= count
	avg.EventsPerSecond /= float64(count)
	avg.MemoryUsageMB /= float64(count)
	avg.DatabaseSizeMB /= float64(count)
	avg.ErrorCount /= count

	return avg
}

func (pa *PerformanceAnalyzer) analyzePatterns(rec *PerformanceRecommendation) {
	maxEPS := rec.BestPerformance.EventsPerSecond

	if maxEPS < 100 {
		rec.Recommendation = "Switch to PostgreSQL immediately - SQLite cannot handle even basic load"
		rec.Confidence = "High"
		rec.Reasoning = "Maximum events/second is below acceptable threshold for any real application"
	} else if maxEPS < 1000 {
		rec.Recommendation = "Consider switching to PostgreSQL for better performance and concurrency"
		rec.Confidence = "Medium"
		rec.Reasoning = "SQLite shows performance limitations under moderate load"
	} else if maxEPS < 5000 {
		rec.Recommendation = "SQLite acceptable for development/testing, monitor production usage"
		rec.Confidence = "Medium"
		rec.Reasoning = "SQLite performing adequately but may not scale well"
	} else {
		rec.Recommendation = "SQLite performing well for current workload"
		rec.Confidence = "High"
		rec.Reasoning = "Performance metrics indicate SQLite can handle the current load effectively"
	}

	// Check for memory issues
	if rec.BestPerformance.MemoryUsageMB > 500 { // 500MB threshold
		rec.Recommendation += " (Note: High memory usage detected)"
	}

	// Check for database size issues
	if rec.BestPerformance.DatabaseSizeMB > 1000 { // 1GB threshold
		rec.Recommendation += " (Note: Large database size may impact performance)"
	}
}

type PerformanceRecommendation struct {
	BestPerformance    BenchmarkResult `json:"best_performance"`
	WorstPerformance   BenchmarkResult `json:"worst_performance"`
	AveragePerformance BenchmarkResult `json:"average_performance"`
	Recommendation     string          `json:"recommendation"`
	Confidence         string          `json:"confidence"`
	Reasoning          string          `json:"reasoning"`
}
