package internal

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	httpEndpointResponseTime *prometheus.SummaryVec
	httpResponseCode         *prometheus.CounterVec
}

// NewMetrics creates a new Metrics handler
func NewMetrics() *Metrics {
	// Summarize endpoint response times
	httpEndpointResponseTime := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "echo_rest_api",
		Name:      "endpoint_response_times",
		Help:      "Endpoint response times",
	}, []string{"endpoint"})
	prometheus.MustRegister(httpEndpointResponseTime)

	// Count response codes
	httpResponseCode := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "echo_rest_api",
		Name:        "response_codes",
		Help:        "Response codes",
		ConstLabels: nil,
	}, []string{"code"})
	prometheus.MustRegister(httpResponseCode)

	return &Metrics{
		httpEndpointResponseTime: httpEndpointResponseTime,
		httpResponseCode:         httpResponseCode,
	}
}

// Handle metrics processing for all incoming requests
func (m *Metrics) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Start timing
		start := time.Now()

		// Execute the request
		err := next(c)
		if err != nil {
			c.Error(err)
		}

		// Observe the time passed for execution
		end := time.Since(start).Seconds()
		m.httpEndpointResponseTime.WithLabelValues(c.Path()).Observe(end)

		// Increment the count for the response code
		m.httpResponseCode.WithLabelValues(strconv.Itoa(c.Response().Status)).Inc()

		return err
	}
}
