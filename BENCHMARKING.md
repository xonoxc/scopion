# Scopion Benchmarking System

This document describes the comprehensive benchmarking system for testing SQLite database limits and performance characteristics.

## Overview

The benchmarking system provides multiple tools to stress test your SQLite database and determine when you should consider switching to a more scalable database solution like PostgreSQL or MySQL.

## Benchmark Commands

### Standard Benchmark

Run a standard load test with configurable parameters:

```bash
# Basic benchmark with 10 workers for 30 seconds
./scopion benchmark standard

# Custom configuration
./scopion benchmark standard --workers 50 --duration 60s --rate 1000 --output results.json
```

**Flags:**
- `--workers, -w`: Number of concurrent workers (default: 10)
- `--duration, -d`: Test duration (default: 30s)
- `--rate, -r`: Target events per second (0 for unlimited, default: 0)
- `--output, -o`: Save results to JSON file

### Stress Test

Run a progressive stress test that gradually increases load to find breaking points:

```bash
./scopion benchmark stress --workers 10 --duration 30s --output stress-results.json
```

This test runs multiple phases with increasing concurrency and load.

### Database Limits Test

Comprehensive test suite that pushes SQLite to its limits and provides migration recommendations:

```bash
./scopion benchmark limits --output limits-report.json
```

This test includes:
- Memory exhaustion testing
- Concurrent write limits
- Large transaction handling
- Read-write contention analysis

### Continuous Monitoring

Run ongoing performance monitoring with periodic benchmarks:

```bash
./scopion benchmark monitor --workers 5 --rate 100
```

This runs continuous monitoring and alerts on performance degradation. Press Ctrl+C to stop.

## Interpreting Results

### Key Metrics

- **Events/Second**: Throughput capacity
- **Average Latency**: Response time under load
- **Memory Usage**: RAM consumption
- **Database Size**: Storage requirements
- **Error Count**: Reliability under load

### Performance Thresholds

| Metric | Excellent | Good | Concerning | Critical |
|--------|-----------|------|------------|----------|
| Events/sec | >1000 | 500-1000 | 100-500 | <100 |
| Avg Latency | <5ms | 5-20ms | 20-100ms | >100ms |
| Error Rate | 0% | <1% | 1-5% | >5% |

### Migration Recommendations

**Switch to PostgreSQL/MySQL when:**
- Events/second consistently below 500
- Average latency above 20ms under normal load
- High error rates (>1%) under moderate concurrency
- Database size approaches SQLite's practical limits (~1GB)

**SQLite is acceptable for:**
- Development and testing environments
- Low-traffic applications (<100 concurrent users)
- Read-heavy workloads
- Simple data models

## Example Usage

### Quick Performance Check

```bash
# 10-second test with 5 workers
./scopion benchmark standard --workers 5 --duration 10s

# Expected output:
# === Standard Benchmark Results ===
# Duration: 10.05s
# Total Events: 2847
# Events/Second: 283.28
# Avg Latency: 17.52ms
# Memory Usage: 5.2 MB
# Database Size: 2.1 MB
# Errors: 0
```

### Production Load Testing

```bash
# Simulate production load
./scopion benchmark standard --workers 20 --duration 5m --rate 500 --output prod-test.json
```

### Database Limits Assessment

```bash
# Comprehensive limits testing
./scopion benchmark limits --output migration-assessment.json

# This will show:
# - Maximum safe concurrency
# - Performance degradation points
# - Migration recommendations
```

### Continuous Monitoring

```bash
# Monitor performance during deployment
./scopion benchmark monitor --workers 10 --rate 200

# Will show real-time metrics and alert on degradation
```

## Troubleshooting

### Common Issues

1. **Database locked errors**: Reduce concurrency or use WAL mode
2. **Out of memory**: Lower worker count or batch size
3. **Slow performance**: Check disk I/O, consider SSD upgrade
4. **High latency**: Monitor system resources, check for contention

### Optimizing SQLite Performance

```sql
-- Enable WAL mode for better concurrency
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=1000000; -- 1GB cache
PRAGMA temp_store=MEMORY;
```

### Alternative Database Options

**PostgreSQL:**
- Better for high concurrency
- Advanced features (JSON, full-text search)
- ACID compliance
- Good for complex queries

**MySQL/MariaDB:**
- Excellent performance
- Wide ecosystem support
- Good for web applications
- Strong replication features

**ClickHouse:**
- Excellent for analytics workloads
- High ingestion rates
- Column-oriented storage
- Good for time-series data

## Architecture

The benchmarking system consists of:

- **LoadGenerator**: Core benchmarking engine
- **Monitor**: Real-time performance monitoring
- **StressTest**: Progressive load testing
- **Analyzer**: Results analysis and recommendations
- **CLI Interface**: Command-line tools

## Extending the System

### Adding Custom Tests

```go
type CustomTest struct {
    *DatabaseStressTest
}

func (ct *CustomTest) RunCustomLoadTest(ctx context.Context) (*BenchmarkResult, error) {
    // Implement custom test logic
    config := ct.config
    config.Workers = 100 // Extreme concurrency

    generator, err := NewLoadGenerator(config)
    if err != nil {
        return nil, err
    }

    return generator.Run(ctx)
}
```

### Custom Metrics

```go
type CustomMonitor struct {
    *Monitor
    customMetrics map[string]float64
}

func (cm *CustomMonitor) collectCustomStats() {
    // Add custom monitoring logic
    cm.customMetrics["custom_metric"] = someCalculation()
}
```

## Best Practices

1. **Run benchmarks on production-like hardware**
2. **Test with realistic data patterns**
3. **Monitor system resources during tests**
4. **Run tests multiple times for consistency**
5. **Document your performance requirements**
6. **Set up automated performance regression testing**

## Contributing

When adding new benchmark tests:

1. Follow the existing patterns in `internal/benchmark/`
2. Add comprehensive error handling
3. Include performance analysis
4. Update this documentation
5. Add CLI commands for new tests