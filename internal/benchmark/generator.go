package benchmark

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/store/sqlite"
)

type BenchmarkConfig struct {
	Duration     time.Duration
	Workers      int
	BatchSize    int
	EventRate    int
	MaxEvents    int
	EnableWAL    bool
	EnableSync   bool
	DatabasePath string
}

type BenchmarkResult struct {
	TotalEvents     int64
	EventsPerSecond float64
	AvgLatency      time.Duration
	MaxLatency      time.Duration
	MinLatency      time.Duration
	ErrorCount      int64
	DatabaseSizeMB  float64
	MemoryUsageMB   float64
	Duration        time.Duration
	StartTime       time.Time
	EndTime         time.Time
	SystemStats     SystemStats
}

type SystemStats struct {
	CPUUsage    float64
	MemoryMB    float64
	DiskReads   int64
	DiskWrites  int64
	Connections int
}

type LoadGenerator struct {
	config     BenchmarkConfig
	store      *sqlite.SqliteStore
	results    *BenchmarkResult
	stopChan   chan struct{}
	wg         sync.WaitGroup
	latencies  []time.Duration
	latencyMux sync.Mutex
}

func NewLoadGenerator(config BenchmarkConfig) (*LoadGenerator, error) {
	s, err := sqlite.New(config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	return &LoadGenerator{
		config: config,
		store:  s,
		results: &BenchmarkResult{
			StartTime: time.Now(),
		},
		stopChan:  make(chan struct{}),
		latencies: make([]time.Duration, 0, 10000),
	}, nil
}

func (lg *LoadGenerator) Run(ctx context.Context) (*BenchmarkResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	log.Printf("Starting benchmark: %d workers, target %d events/sec, duration %v", lg.config.Workers, lg.config.EventRate, lg.config.Duration)

	monitor := NewMonitor(lg.config.DatabasePath)
	go monitor.Start(ctx, lg.results)

	for i := 0; i < lg.config.Workers; i++ {
		lg.wg.Add(1)
		go lg.worker(ctx, i)
	}

	if lg.config.EventRate > 0 {
		// TODO: Implement rate limiting
		_ = NewRateLimiter(lg.config.EventRate)
	}

	timer := time.NewTimer(lg.config.Duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		log.Println("Benchmark duration completed")
	case <-ctx.Done():
		log.Println("Benchmark cancelled by context")
	}

	close(lg.stopChan)
	lg.wg.Wait()

	if monitor != nil {
		monitor.Stop()
	}

	lg.results.EndTime = time.Now()
	lg.results.Duration = lg.results.EndTime.Sub(lg.results.StartTime)

	lg.calculateStats()

	if lg.store != nil {
		lg.store.Close()
	}

	log.Printf("Benchmark completed: %d events, %.2f events/sec",
		lg.results.TotalEvents, lg.results.EventsPerSecond)

	return lg.results, nil
}

func (lg *LoadGenerator) worker(ctx context.Context, workerID int) {
	defer lg.wg.Done()

	batch := make([]model.Event, 0, lg.config.BatchSize)

	for {
		select {
		case <-lg.stopChan:
			if len(batch) > 0 {
				lg.processBatch(batch)
			}
			return
		case <-ctx.Done():
			return
		default:
		}

		event := lg.generateEvent(workerID)

		if lg.config.BatchSize > 1 {
			batch = append(batch, event)
			if len(batch) >= lg.config.BatchSize {
				lg.processBatch(batch)
				batch = batch[:0]
			}
		} else {
			lg.processEvent(event)
		}

		if lg.config.MaxEvents > 0 && atomic.LoadInt64(&lg.results.TotalEvents) >= int64(lg.config.MaxEvents) {
			return
		}
	}
}

func (lg *LoadGenerator) processEvent(event model.Event) {
	if lg.store == nil {
		atomic.AddInt64(&lg.results.ErrorCount, 1)
		return
	}

	start := time.Now()

	err := lg.store.Append(event)
	duration := time.Since(start)

	lg.latencyMux.Lock()
	lg.latencies = append(lg.latencies, duration)
	lg.latencyMux.Unlock()

	if err != nil {
		atomic.AddInt64(&lg.results.ErrorCount, 1)
		log.Printf("Error appending event: %v", err)
	} else {
		atomic.AddInt64(&lg.results.TotalEvents, 1)
	}
}

func (lg *LoadGenerator) processBatch(batch []model.Event) {
	if lg.store == nil {
		atomic.AddInt64(&lg.results.ErrorCount, int64(len(batch)))
		return
	}

	start := time.Now()

	// For batch processing, we'd need to extend the store to support batch inserts
	// For now, process individually but track batch latency
	for _, event := range batch {
		err := lg.store.Append(event)
		if err != nil {
			atomic.AddInt64(&lg.results.ErrorCount, 1)
		} else {
			atomic.AddInt64(&lg.results.TotalEvents, 1)
		}
	}

	duration := time.Since(start)
	batchLatency := duration / time.Duration(len(batch))

	lg.latencyMux.Lock()
	for range batch {
		lg.latencies = append(lg.latencies, batchLatency)
	}
	lg.latencyMux.Unlock()
}

func (lg *LoadGenerator) generateEvent(workerID int) model.Event {
	return model.Event{
		ID:        fmt.Sprintf("bench-%d-%d-%d", workerID, time.Now().UnixNano(), rand.Int63()),
		Timestamp: time.Now(),
		Service:   fmt.Sprintf("bench-service-%d", rand.Intn(10)),
		Name:      fmt.Sprintf("bench-operation-%d", rand.Intn(20)),
		TraceID:   fmt.Sprintf("bench-trace-%d-%d", workerID, rand.Int63()),
		Level:     []string{"info", "warn", "error"}[rand.Intn(3)],
		Data: map[string]any{
			"worker_id": workerID,
			"batch_id":  rand.Intn(1000),
			"payload":   fmt.Sprintf("benchmark data %d", rand.Intn(10000)),
			"metadata": map[string]any{
				"cpu":       runtime.NumCPU(),
				"goroutine": runtime.NumGoroutine(),
			},
		},
	}
}

func (lg *LoadGenerator) calculateStats() {
	totalEvents := atomic.LoadInt64(&lg.results.TotalEvents)

	if totalEvents == 0 {
		return
	}

	lg.results.EventsPerSecond = float64(totalEvents) / lg.results.Duration.Seconds()

	lg.latencyMux.Lock()
	defer lg.latencyMux.Unlock()

	if len(lg.latencies) == 0 {
		return
	}

	var totalLatency time.Duration
	minLatency := lg.latencies[0]
	maxLatency := lg.latencies[0]

	for _, latency := range lg.latencies {
		totalLatency += latency
		if latency < minLatency {
			minLatency = latency
		}
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	lg.results.AvgLatency = totalLatency / time.Duration(len(lg.latencies))
	lg.results.MinLatency = minLatency
	lg.results.MaxLatency = maxLatency
}

// StressTest runs increasingly aggressive benchmarks to find breaking points
func (lg *LoadGenerator) StressTest(ctx context.Context, maxWorkers int, maxDuration time.Duration) ([]BenchmarkResult, error) {
	results := make([]BenchmarkResult, 0)

	// Start with conservative settings and gradually increase load
	configs := []BenchmarkConfig{
		{Duration: 30 * time.Second, Workers: 1, BatchSize: 1, EventRate: 100},
		{Duration: 30 * time.Second, Workers: 5, BatchSize: 1, EventRate: 500},
		{Duration: 30 * time.Second, Workers: 10, BatchSize: 10, EventRate: 1000},
		{Duration: 30 * time.Second, Workers: 20, BatchSize: 50, EventRate: 2000},
		{Duration: 30 * time.Second, Workers: 50, BatchSize: 100, EventRate: 5000},
		{Duration: 60 * time.Second, Workers: maxWorkers, BatchSize: 200, EventRate: 10000},
	}

	for i, config := range configs {
		log.Printf("Running stress test phase %d/%d: %d workers, %d events/sec",
			i+1, len(configs), config.Workers, config.EventRate)

		config.DatabasePath = lg.config.DatabasePath
		config.Duration = maxDuration / time.Duration(len(configs))

		generator, err := NewLoadGenerator(config)
		if err != nil {
			log.Printf("Failed to create generator for phase %d: %v", i+1, err)
			continue
		}

		result, err := generator.Run(ctx)
		if err != nil {
			log.Printf("Stress test phase %d failed: %v", i+1, err)
			results = append(results, *result) // Still record partial results
		} else {
			results = append(results, *result)
		}

		// Brief pause between phases
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return results, ctx.Err()
		}
	}

	return results, nil
}
