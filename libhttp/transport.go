package libhttp

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	methodKey = "http_client_method"
	pathKey   = "http_client_path"
)

type transportConfig struct {
	meter                metric.MeterProvider
	serviceName          string
	timeout              time.Duration
	maxRetries           int
	retryableStatusCodes []int
	logger               *slog.Logger
}

type TransportOption func(c *transportConfig)

func WithMeterProvider(meter metric.MeterProvider) func(c *transportConfig) {
	return func(c *transportConfig) {
		c.meter = meter
	}
}

func WithTransportName(name string) func(c *transportConfig) {
	return func(c *transportConfig) {
		c.serviceName = name
	}
}

func WithTimeout(timeout time.Duration) func(c *transportConfig) {
	return func(c *transportConfig) {
		c.timeout = timeout
	}
}

func WithRetry(maxRetries int, retryableStatusCodes []int) func(c *transportConfig) {
	return func(c *transportConfig) {
		c.maxRetries = maxRetries
		c.retryableStatusCodes = retryableStatusCodes
	}
}

func WithLogger(logger *slog.Logger) func(c *transportConfig) {
	return func(c *transportConfig) {
		c.logger = logger
	}
}

var _ http.RoundTripper = (*Transport)(nil)

type Transport struct {
	base             http.RoundTripper
	requestCounter   metric.Int64Counter
	errorsCounter    metric.Int64Counter
	latencyHistogram metric.Int64Histogram
}

func NewTransport(base http.RoundTripper, configs ...TransportOption) *Transport {
	cfg := transportConfig{
		meter:       otel.GetMeterProvider(),
		serviceName: "http_client",
	}
	for _, c := range configs {
		c(&cfg)
	}

	meterProvider := cfg.meter.Meter(cfg.serviceName)

	requestCounter, _ := meterProvider.Int64Counter("http_client_request",
		metric.WithUnit("1"),
		metric.WithDescription("count calls with methods, paths and status codes"),
	)
	errorsCounter, _ := meterProvider.Int64Counter("http_client_errors",
		metric.WithUnit("1"),
		metric.WithDescription("count of errors for http clients"),
	)
	latencyHistogram, _ := meterProvider.Int64Histogram(
		"http_client_latency",
		metric.WithUnit("ms"),
		metric.WithDescription("latency buckets for request durations with paths and methods"),
	)
	return &Transport{
		base:             base,
		requestCounter:   requestCounter,
		latencyHistogram: latencyHistogram,
		errorsCounter:    errorsCounter,
	}
}

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	startTime := time.Now()
	resp, err := t.base.RoundTrip(request)
	defer func() {
		ctx := context.Background() // we can't use request.Context(), because them can be canceled
		path := "/"
		if request.URL != nil {
			path = request.URL.Path
		}

		// Track response status code for http client call.
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		t.requestCounter.Add(ctx,
			1,
			metric.WithAttributes(
				attribute.String(methodKey, request.Method),
				attribute.Int("status_code", statusCode),
				attribute.String(pathKey, path),
			),
		)

		// Track latency for every http client call.
		t.latencyHistogram.Record(ctx,
			time.Since(startTime).Milliseconds(),
			metric.WithAttributes(
				attribute.String(methodKey, request.Method),
				attribute.String(pathKey, path),
			),
		)
		if err != nil {
			t.errorsCounter.Add(
				ctx,
				1,
				metric.WithAttributes(
					attribute.String("status", err.Error()),
				),
			)
		}
	}()
	return resp, err
}
