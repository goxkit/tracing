module github.com/goxkit/tracing/examples

go 1.24.3

require (
	github.com/goxkit/tracing v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.10.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0
	go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/attribute v1.37.0
	go.opentelemetry.io/otel/codes v1.37.0
	go.opentelemetry.io/otel/trace v1.37.0
	go.uber.org/zap v1.27.0
)

replace github.com/goxkit/tracing => ../
