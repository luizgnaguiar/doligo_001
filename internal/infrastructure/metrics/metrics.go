package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// EndpointTiming holds the metrics for a single endpoint.
type EndpointTiming struct {
	TotalDuration atomic.Int64
	RequestCount  atomic.Uint64
}

// Metrics holds all application metrics.
type Metrics struct {
	startTime       time.Time
	totalRequests   atomic.Uint64
	errorRequests   atomic.Uint64
	endpointTimings sync.Map // map[string]*EndpointTiming
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// IncTotalRequests increments the total number of requests.
func (m *Metrics) IncTotalRequests() {
	m.totalRequests.Add(1)
}

// IncErrorRequests increments the number of error requests (4xx/5xx).
func (m *Metrics) IncErrorRequests() {
	m.errorRequests.Add(1)
}

// AddEndpointTiming records the duration for a specific endpoint.
func (m *Metrics) AddEndpointTiming(endpoint string, duration time.Duration) {
	val, _ := m.endpointTimings.LoadOrStore(endpoint, &EndpointTiming{})
	et := val.(*EndpointTiming)
	et.RequestCount.Add(1)
	et.TotalDuration.Add(int64(duration))
}

// GetMetrics returns a snapshot of the current metrics.
func (m *Metrics) GetMetrics() map[string]interface{} {
	data := make(map[string]interface{})
	data["start_time"] = m.startTime.Format(time.RFC3339)
	data["uptime_seconds"] = time.Since(m.startTime).Seconds()
	data["total_requests"] = m.totalRequests.Load()
	data["error_requests_4xx_5xx"] = m.errorRequests.Load()

	endpoints := make(map[string]interface{})
	m.endpointTimings.Range(func(key, value interface{}) bool {
		et := value.(*EndpointTiming)
		count := et.RequestCount.Load()
		totalDuration := et.TotalDuration.Load()
		avgDuration := float64(0)
		if count > 0 {
			avgDuration = float64(totalDuration) / float64(count) / float64(time.Millisecond)
		}
		endpoints[key.(string)] = map[string]interface{}{
			"requests":             count,
			"total_duration_ns":    totalDuration,
			"avg_response_time_ms": avgDuration,
		}
		return true
	})
	data["endpoints"] = endpoints

	return data
}
