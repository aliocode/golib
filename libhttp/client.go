package libhttp

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	Attributes []trace.SpanStartOption
}

type Option func(c *config)

func WithServiceName(serviceName string) func(c *config) {
	return func(c *config) {
		c.Attributes = append(
			c.Attributes,
			trace.WithAttributes(attribute.String("service.name", serviceName)),
		)
	}
}

func WithServiceVersion(serviceVersion string) func(c *config) {
	return func(c *config) {
		c.Attributes = append(
			c.Attributes,
			trace.WithAttributes(attribute.String("service.version", serviceVersion)),
		)
	}
}

// todo rework transport option

func NewClient(opts ...Option) *http.Client {
	c := config{}
	for _, o := range opts {
		o(&c)
	}
	return &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				if method := r.Header.Get("x-rpc-method-name"); method != "" {
					return method
				}
				return operation
			}),
			otelhttp.WithSpanOptions(c.Attributes...),
		),
	}
}
