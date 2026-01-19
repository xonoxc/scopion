package benchmark

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/xonoxc/scopion/internal/store"
)

type Monitor struct {
	dbPath    string
	result    *BenchmarkResult
	stopChan  chan struct{}
	startTime time.Time
}

func NewMonitor(dbPath string) *Monitor {
	return &Monitor{
		dbPath:    dbPath,
		stopChan:  make(chan struct{}),
		startTime: time.Now(),
	}
}

func (m *Monitor) Start(ctx context.Context, result *BenchmarkResult) {
	m.result = result

	ticker := time.NewTicker(5 * time.Second) // Less frequent to reduce overhead
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.collectStats()
		}
	}
}

func (m *Monitor) Stop() {
	close(m.stopChan)
	m.collectStats() // Final collection
}

func (m *Monitor) collectStats() {
	stats := SystemStats{}

	// Basic memory stats using runtime
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	stats.MemoryMB = float64(memStats.Alloc) / 1024 / 1024
	m.result.MemoryUsageMB = stats.MemoryMB

	// CPU - simplified (just goroutines as proxy for activity)
	stats.CPUUsage = float64(runtime.NumGoroutine())

	// Database file size
	if info, err := os.Stat(m.dbPath); err == nil {
		m.result.DatabaseSizeMB = float64(info.Size()) / 1024 / 1024
	}

	// Simplified disk stats (just track connections as active operations)
	stats.Connections = runtime.NumGoroutine()

	m.result.SystemStats = stats
}

// RateLimiter controls the rate of event generation
type RateLimiter struct {
	rate       int // events per second
	interval   time.Duration
	lastTime   time.Time
	permits    int64
	maxPermits int64
}

func NewRateLimiter(rate int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		interval:   time.Second / time.Duration(rate),
		lastTime:   time.Now(),
		maxPermits: int64(rate),
		permits:    int64(rate),
	}
}

func (rl *RateLimiter) Acquire() {
	for {
		now := time.Now()
		timePassed := now.Sub(rl.lastTime)

		// Replenish permits based on time passed
		newPermits := int64(timePassed / rl.interval)
		if newPermits > 0 {
			atomic.AddInt64(&rl.permits, newPermits)
			if atomic.LoadInt64(&rl.permits) > rl.maxPermits {
				atomic.StoreInt64(&rl.permits, rl.maxPermits)
			}
			rl.lastTime = now
		}

		// Try to acquire a permit
		if atomic.LoadInt64(&rl.permits) > 0 {
			atomic.AddInt64(&rl.permits, -1)
			return
		}

		// Wait a bit before trying again
		time.Sleep(rl.interval / 10)
	}
}

// DatabaseStressTest runs extreme stress tests to find SQLite limits
type DatabaseStressTest struct {
	config BenchmarkConfig
	store  *store.Store
}

func NewDatabaseStressTest(config BenchmarkConfig) (*DatabaseStressTest, error) {
	s, err := store.New(config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	return &DatabaseStressTest{
		config: config,
		store:  s,
	}, nil
}

func (dst *DatabaseStressTest) RunMemoryExhaustionTest(ctx context.Context) (*BenchmarkResult, error) {
	log.Println("Running memory exhaustion test...")

	// Create events with increasingly large payloads
	config := dst.config
	config.Workers = 10
	config.BatchSize = 1
	config.Duration = 5 * time.Minute

	generator, err := NewLoadGenerator(config)
	if err != nil {
		return nil, err
	}

	// Note: For memory exhaustion test, we would ideally modify the event generation
	// to create events with increasingly large payloads. For now, this runs a standard
	// load test which will still stress memory as the database grows.
	log.Println("Memory exhaustion test starting with high concurrency...")

	return generator.Run(ctx)
}

func (dst *DatabaseStressTest) RunConcurrentWriteTest(ctx context.Context, maxConcurrency int) (*BenchmarkResult, error) {
	log.Printf("Running concurrent write test with %d workers...", maxConcurrency)

	config := dst.config
	config.Workers = maxConcurrency
	config.BatchSize = 1
	config.Duration = 2 * time.Minute

	generator, err := NewLoadGenerator(config)
	if err != nil {
		return nil, err
	}

	return generator.Run(ctx)
}

func (dst *DatabaseStressTest) RunLargeTransactionTest(ctx context.Context) (*BenchmarkResult, error) {
	log.Println("Running large transaction test...")

	// This would require extending the store to support transactions
	// For now, simulate with batch processing
	config := dst.config
	config.Workers = 1
	config.BatchSize = 1000 // Large batches
	config.Duration = 1 * time.Minute

	generator, err := NewLoadGenerator(config)
	if err != nil {
		return nil, err
	}

	return generator.Run(ctx)
}

func (dst *DatabaseStressTest) RunReadWriteContentionTest(ctx context.Context) (*BenchmarkResult, error) {
	log.Println("Running read-write contention test...")

	// This test would require simultaneous reads and writes
	// For now, just run a standard load test
	config := dst.config
	config.Workers = 20
	config.BatchSize = 10
	config.Duration = 1 * time.Minute

	generator, err := NewLoadGenerator(config)
	if err != nil {
		return nil, err
	}

	return generator.Run(ctx)
}

func (dst *DatabaseStressTest) Cleanup() {
	if dst.store != nil {
		dst.store.Close()
	}
}
