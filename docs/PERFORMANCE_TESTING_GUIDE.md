# GoNotes Performance Testing Guide

This guide provides comprehensive information about performance testing for the GoNotes application.

## Table of Contents

1. [Overview](#overview)
2. [Performance Test Scripts](#performance-test-scripts)
3. [Test Types](#test-types)
4. [Running Tests](#running-tests)
5. [Understanding Results](#understanding-results)
6. [Performance Metrics](#performance-metrics)
7. [Optimization Tips](#optimization-tips)
8. [Troubleshooting](#troubleshooting)

## Overview

GoNotes performance testing suite provides comprehensive testing capabilities to ensure optimal application performance under various load conditions. The testing suite includes:

- **Load Testing**: Normal expected load
- **Stress Testing**: Beyond normal capacity
- **Spike Testing**: Sudden load increases
- **Volume Testing**: Large amounts of data
- **Endurance Testing**: Extended periods

## Performance Test Scripts

### 1. Comprehensive Performance Test (`performance_test.sh`)

The main performance testing script that runs comprehensive tests across all endpoints.

**Features:**
- Tests all major API endpoints
- Multiple load levels (light, medium, heavy)
- Stress and spike testing
- Detailed reporting with timestamps
- System resource monitoring
- Automatic cleanup

**Usage:**
```bash
# Standard comprehensive test
./scripts/performance_test.sh

# Quick mode (reduced load)
./scripts/performance_test.sh --quick

# High-stress mode
./scripts/performance_test.sh --stress
```

### 2. Quick Performance Test (`quick_performance_test.sh`)

A lightweight performance test for quick validation during development.

**Features:**
- 4 key endpoint tests
- 20 concurrent users
- Quick execution (~2-3 minutes)
- Basic metrics reporting
- Suitable for CI/CD pipelines

**Usage:**
```bash
# Run quick performance test
./scripts/quick_performance_test.sh
```

## Test Types

### 1. Load Testing
Tests application performance under expected normal load conditions.

**Parameters:**
- Light Load: 10 concurrent users
- Medium Load: 50 concurrent users  
- Heavy Load: 100 concurrent users
- Duration: 60 seconds per test

**Endpoints Tested:**
- Health check (`/health`)
- Authentication (`/api/v1/auth/login`)
- Profile management (`/api/v1/user/profile`)
- Notes operations (`/api/v1/notes`)
- Session management (`/api/v1/user/sessions`)

### 2. Stress Testing
Tests application behavior under extreme load conditions.

**Parameters:**
- Stress Load: 200 concurrent users
- Duration: 30 seconds
- Purpose: Find breaking points

### 3. Spike Testing
Tests application response to sudden traffic spikes.

**Parameters:**
- Spike Load: 500 concurrent users
- Duration: 15 seconds
- Purpose: Test auto-scaling and recovery

### 4. Volume Testing
Tests application with large amounts of data.

**Focus Areas:**
- Database performance with large datasets
- Memory usage with bulk operations
- Response times with complex queries

### 5. Endurance Testing
Tests application stability over extended periods.

**Parameters:**
- Moderate Load: 50 concurrent users
- Duration: 5 minutes (300 seconds)
- Purpose: Detect memory leaks and degradation

## Running Tests

### Prerequisites

1. **Server Running**: Ensure GoNotes server is running on `localhost:8080`
2. **Dependencies**: Install required tools:
   ```bash
   # macOS
   brew install bc curl

   # Ubuntu/Debian
   sudo apt-get install bc curl

   # CentOS/RHEL
   sudo yum install bc curl
   ```

### Using Make Commands

```bash
# Comprehensive performance test
make test-performance

# Quick performance test
make test-perf-quick

# Stress performance test
make test-perf-stress
```

### Using Aliases

```bash
# Source aliases first
source aliases.sh

# Run tests
gn-test-performance  # Comprehensive test
gn-test-perf        # Quick test
gn-test-stress      # Stress test
```

### Direct Script Execution

```bash
# Make scripts executable
chmod +x scripts/performance_test.sh
chmod +x scripts/quick_performance_test.sh

# Run tests
./scripts/performance_test.sh
./scripts/quick_performance_test.sh
```

## Understanding Results

### Performance Metrics

Each test provides the following metrics:

1. **Total Requests**: Number of HTTP requests made
2. **Successful Requests**: Requests with 2xx status codes
3. **Failed Requests**: Requests with non-2xx status codes
4. **Success Rate**: Percentage of successful requests
5. **Requests/Second (RPS)**: Request throughput
6. **Response Times**:
   - Average response time
   - Minimum response time
   - Maximum response time

### Sample Output

```
[PERF] Results for 50 users on GET /api/v1/notes:
  Total Requests: 2847
  Successful: 2847
  Failed: 0
  Success Rate: 100.00%
  Requests/Second: 47.45
  Avg Response Time: 0.125s
  Min Response Time: 0.045s
  Max Response Time: 0.892s
```

### Report Files

Detailed reports are saved in the `performance_results/` directory:

- **Report File**: `performance_report_YYYYMMDD_HHMMSS.txt`
- **System Resources**: Memory and CPU usage before/after tests
- **Test Configuration**: Load levels and duration settings
- **Timestamps**: Detailed timing information

## Performance Benchmarks

### Expected Performance Levels

**Health Endpoint:**
- RPS: 200-500+ (lightweight endpoint)
- Response Time: <50ms
- Success Rate: 100%

**Authentication Endpoints:**
- RPS: 50-100 (includes encryption)
- Response Time: <200ms
- Success Rate: 99%+

**Profile Endpoints (with Redis caching):**
- RPS: 100-200 (cached responses)
- Response Time: <100ms
- Success Rate: 99%+

**Notes Endpoints:**
- RPS: 50-150 (database operations)
- Response Time: <300ms
- Success Rate: 99%+

**Session Management:**
- RPS: 80-150 (Redis operations)
- Response Time: <150ms
- Success Rate: 99%+

### Performance Thresholds

**Green (Good):**
- Success Rate: >99%
- Response Time: Within expected ranges
- No significant errors

**Yellow (Warning):**
- Success Rate: 95-99%
- Response Time: 150% of expected
- Occasional timeout errors

**Red (Critical):**
- Success Rate: <95%
- Response Time: >200% of expected
- Frequent errors or timeouts

## Optimization Tips

### 1. Database Optimization

```go
// Use connection pooling
db, err := sql.Open("postgres", "...")
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 2. Redis Caching

```go
// Implement cache-aside pattern
func GetUserProfile(userID string) (*User, error) {
    // Try cache first
    if user, err := redis.Get(ctx, "user:"+userID); err == nil {
        return user, nil
    }
    
    // Fallback to database
    user, err := db.GetUser(userID)
    if err != nil {
        return nil, err
    }
    
    // Cache for future requests
    redis.Set(ctx, "user:"+userID, user, 15*time.Minute)
    return user, nil
}
```

### 3. Rate Limiting

```go
// Implement appropriate rate limits
limiter := rate.NewLimiter(rate.Every(time.Second), 100)
```

### 4. Database Indexing

```sql
-- Add indexes for frequently queried fields
CREATE INDEX idx_notes_user_id ON notes(user_id);
CREATE INDEX idx_notes_created_at ON notes(created_at);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
```

### 5. Response Compression

```go
// Enable gzip compression
app.Use(middleware.GzipWithConfig(middleware.GzipConfig{
    Level: 5,
}))
```

## Troubleshooting

### Common Issues

#### 1. Port Already in Use
```bash
# Check what's using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

#### 2. Database Connection Issues
```bash
# Check database connectivity
docker-compose -f docker-compose.dev.yaml ps
docker-compose -f docker-compose.dev.yaml logs db
```

#### 3. High Response Times
- Check database query performance
- Verify Redis cache hit rates
- Monitor system resources
- Review database indexes

#### 4. Low Success Rates
- Check application logs for errors
- Verify database and Redis connectivity
- Review rate limiting configurations
- Check for resource exhaustion

#### 5. Memory Issues
```bash
# Monitor memory usage
docker stats

# Check for memory leaks
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Performance Debugging

#### 1. Enable Profiling
```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

#### 2. Monitor Metrics
```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

#### 3. Database Query Analysis
```sql
-- Enable query logging in PostgreSQL
ALTER SYSTEM SET log_statement = 'all';
SELECT pg_reload_conf();

-- Analyze slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;
```

### Best Practices

1. **Test Environment**: Use dedicated environment for performance testing
2. **Data Isolation**: Use separate test data that doesn't affect production
3. **Baseline Testing**: Establish baseline performance metrics
4. **Regular Testing**: Run performance tests in CI/CD pipeline
5. **Monitoring**: Set up monitoring and alerting for production
6. **Documentation**: Document performance requirements and benchmarks

## Advanced Testing

### Custom Test Scenarios

Create custom test scenarios for specific use cases:

```bash
# Test specific endpoint with custom parameters
make_concurrent_requests "/api/v1/notes/search" "POST" 25 60 '{"query":"test"}' "true"
```

### Integration with CI/CD

Add performance testing to your CI/CD pipeline:

```yaml
# .github/workflows/performance.yml
name: Performance Tests
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  performance:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Setup Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
    - name: Start services
      run: |
        make dev-db
        make dev-migrate
        make dev-run &
        sleep 10
    - name: Run performance tests
      run: make test-perf-quick
```

### Performance Monitoring

Set up continuous performance monitoring:

```bash
# Use tools like Prometheus + Grafana
# Monitor key metrics:
# - Response times
# - Request rates
# - Error rates
# - Resource utilization
```

## Conclusion

The GoNotes performance testing suite provides comprehensive tools for ensuring optimal application performance. Regular performance testing helps identify bottlenecks, validate optimizations, and maintain consistent user experience under various load conditions.

For more advanced performance testing needs, consider using specialized tools like Apache JMeter, k6, or Artillery alongside the provided scripts. 