package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	RequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)

	// Ad click metrics
	ClickTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ad_clicks_total",
			Help: "Total number of ad clicks",
		},
		[]string{"ad_id"},
	)

	ClickProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "click_processing_duration_seconds",
			Help: "Duration of click processing",
		},
	)

	// Database metrics
	DatabaseOperationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "status"},
	)

	DatabaseOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "database_operation_duration_seconds",
			Help: "Duration of database operations",
		},
		[]string{"operation"},
	)

	// Redis metrics
	RedisOperationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"},
	)

	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "redis_operation_duration_seconds",
			Help: "Duration of Redis operations",
		},
		[]string{"operation"},
	)

	// Queue metrics
	QueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_size",
			Help: "Current size of processing queues",
		},
		[]string{"queue_name"},
	)

	QueueProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "queue_processing_duration_seconds",
			Help: "Duration of queue processing",
		},
		[]string{"queue_name"},
	)

	// System metrics
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)

	ErrorTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"error_type", "component"},
	)
)

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, endpoint, statusCode string, duration float64) {
	RequestTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	RequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// RecordClick records ad click metrics
func RecordClick(adID string, duration float64) {
	ClickTotal.WithLabelValues(adID).Inc()
	ClickProcessingDuration.Observe(duration)
}

// RecordDatabaseOperation records database operation metrics
func RecordDatabaseOperation(operation, status string, duration float64) {
	DatabaseOperationTotal.WithLabelValues(operation, status).Inc()
	DatabaseOperationDuration.WithLabelValues(operation).Observe(duration)
}

// RecordRedisOperation records Redis operation metrics
func RecordRedisOperation(operation, status string, duration float64) {
	RedisOperationTotal.WithLabelValues(operation, status).Inc()
	RedisOperationDuration.WithLabelValues(operation).Observe(duration)
}

// RecordError records error metrics
func RecordError(errorType, component string) {
	ErrorTotal.WithLabelValues(errorType, component).Inc()
}

// UpdateQueueSize updates queue size metrics
func UpdateQueueSize(queueName string, size float64) {
	QueueSize.WithLabelValues(queueName).Set(size)
}

// RecordQueueProcessing records queue processing metrics
func RecordQueueProcessing(queueName string, duration float64) {
	QueueProcessingDuration.WithLabelValues(queueName).Observe(duration)
}
