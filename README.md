# GoKit Tracing

<p align="center">
  <a href="https://github.com/goxkit/tracing/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License">
  </a>
  <a href="https://pkg.go.dev/github.com/goxkit/tracing">
    <img src="https://godoc.org/github.com/goxkit/tracing?status.svg" alt="Go Doc">
  </a>
  <a href="https://goreportcard.com/report/github.com/goxkit/tracing">
    <img src="https://goreportcard.com/badge/github.com/goxkit/tracing" alt="Go Report Card">
  </a>
  <a href="https://github.com/goxkit/tracing/actions">
    <img src="https://github.com/goxkit/tracing/actions/workflows/action.yml/badge.svg?branch=main" alt="Build Status">
  </a>
</p>

A comprehensive Go tracing package built on OpenTelemetry, providing distributed tracing capabilities for microservices and complex systems. This package enables end-to-end transaction monitoring across service boundaries, including both synchronous HTTP calls and asynchronous messaging systems.

## Features

- **OpenTelemetry Integration**:
  - Full OpenTelemetry SDK implementation
  - Support for trace sampling, resource attributes, and span processors
  - W3C TraceContext propagation format for interoperability

- **Multiple Transport Options**:
  - OTLP (OpenTelemetry Protocol) export over gRPC
  - Configurable connection parameters (timeout, reconnection, compression)
  - Support for secure (TLS) and insecure connections

- **Cross-Protocol Propagation**:
  - HTTP header propagation (via standard OpenTelemetry mechanisms)
  - AMQP message header propagation for message queues
  - Seamless tracing across synchronous and asynchronous communication

- **Observability Integration**:
  - Structured logging integration with Zap for trace context
  - Automatic correlation between traces and logs
  - Configuration via shared configs package for consistent settings

## Usage

### Using with ConfigsBuilder (Recommended)

The easiest way to set up tracing is through the ConfigsBuilder, which handles all the necessary configuration:

```go
package main

import (
	"context"
	
	"github.com/goxkit/configs_builder"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func main() {
	// Create configuration with OTLP enabled
	cfgs, err := configsBuilder.NewConfigsBuilder().Otlp().Build()
	if err != nil {
		panic(err)
	}
	
	// Get a tracer from the global provider
	tracer := otel.GetTracerProvider().Tracer("my-service")
	
	// Create a span
	ctx, span := tracer.Start(context.Background(), "main-operation")
	defer span.End()
	
	// Add attributes to the span
	span.SetAttributes(
		attribute.String("user.id", "12345"),
		attribute.String("operation.type", "checkout"),
	)
	
	// Log with trace context (correlates logs and traces)
	cfgs.Logger.Info("Starting operation", 
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("operation", "checkout"))
		
	// Perform work within the traced context
	processOrder(ctx, tracer)
}
```

### Manual Setup

If you need more control over the tracing setup:

```go
package main

import (
	"context"
	
	"github.com/goxkit/configs"
	"github.com/goxkit/tracing"
	"go.opentelemetry.io/otel"
)

func main() {
	// Create application configs
	cfgs := &configs.Configs{
		AppConfigs: &configs.AppConfigs{
			Name:        "MyService",
			Namespace:   "my-namespace", 
			Environment: configs.DevelopmentEnv,
		},
		OTLPConfigs: &configs.OTLPConfigs{
			Enabled:  true,
			Endpoint: "localhost:4317",
			ExporterTimeout: 10 * time.Second,
		},
	}
	
	// Install tracing
	tracerProvider, err := tracing.Install(cfgs)
	if err != nil {
		panic(err)
	}
	
	// Ensure proper shutdown
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			// Handle shutdown error
		}
	}()
	
	// Your application code
}
```

### Tracing with AMQP Messages

Propagate trace context through message queues to maintain end-to-end tracing:

```go
import (
	"context"
	
	"github.com/goxkit/tracing/amqp"
	amqplib "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

// Producer: Inject trace context into outgoing messages
func sendMessage(ctx context.Context, msg amqplib.Publishing) error {
	// Get tracer
	tracer := otel.Tracer("messaging")
	
	// Create a span for the publish operation
	ctx, span := tracer.Start(ctx, "publish.order-created")
	defer span.End()
	
	// Initialize headers if needed
	if msg.Headers == nil {
		msg.Headers = amqplib.Table{}
	}
	
	// Inject current trace context into message headers
	amqp.AMQPPropagator.Inject(ctx, amqp.AMQPHeader(msg.Headers))
	
	// Publish message (with trace context in headers)
	return channel.PublishWithContext(ctx, exchange, routingKey, false, false, msg)
}

// Consumer: Extract trace context from incoming messages
func handleMessage(delivery amqplib.Delivery) {
	// Get tracer
	tracer := otel.Tracer("messaging")
	
	// Extract trace context from headers and create a new span
	ctx, span := amqp.NewConsumerSpan(tracer, delivery.Headers, "orders-queue")
	defer span.End()
	
	// Process message with trace context
	processOrder(ctx, delivery.Body)
}
```

### Integrating Traces with Logs

Use the `zap` package to automatically add trace context to your logs:

```go
import (
	"context"
	
	"github.com/goxkit/tracing/zap"
	"go.uber.org/zap"
)

func processRequest(ctx context.Context, logger *zap.Logger) {
	// Log with trace context from the current span in ctx
	// This adds trace_id and span_id fields automatically
	logger.Info("Processing request", 
		zap.Format(ctx),
		zap.String("user_id", "user-123"),
		zap.String("request_path", "/api/orders"))
		
	// The log output will include the trace_id and span_id fields
	// making it possible to correlate logs with traces in observability platforms
}
```

### HTTP Handler Example

Complete example of an HTTP handler with tracing:

```go
package main

import (
	"net/http"
	
	"github.com/goxkit/configs_builder"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func main() {
	// Initialize with ConfigsBuilder
	cfg, _ := configsBuilder.NewConfigsBuilder().Otlp().HTTP().Build()
	
	// Create a handler with tracing
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log with trace context
		cfg.Logger.Info("Received request",
			zap.Format(r.Context()),
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method))
			
		// Process the request with the trace context
		// ...
		
		w.WriteHeader(http.StatusOK)
	})
	
	// Wrap the handler with OpenTelemetry instrumentation
	tracedHandler := otelhttp.NewHandler(
		handler,
		"http-server",
		otelhttp.WithTracerProvider(otel.GetTracerProvider()),
	)
	
	// Start the server
	http.Handle("/api/", tracedHandler)
	http.ListenAndServe(":8080", nil)
}
```

## Configuration Options

### OpenTelemetry Configuration 

When using OTLP, configure these settings in your application:

| Setting | Environment Variable | Description |
|---------|---------------------|-------------|
| Endpoint | `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP endpoint (default: `localhost:4317`) |
| Insecure | `OTEL_EXPORTER_OTLP_INSECURE` | Whether to use insecure connection (default: `true`) |
| Timeout | `OTEL_EXPORTER_OTLP_TIMEOUT` | Timeout for export operations (default: `10s`) |
| Headers | `OTEL_EXPORTER_OTLP_HEADERS` | Headers for authentication (format: `key1=value1,key2=value2`) |

### Application Configuration

| Setting | Environment Variable | Description |
|---------|---------------------|-------------|
| Name | `APP_NAME` | Service name for identification |
| Namespace | `APP_NAMESPACE` | Service namespace for grouping |
| Environment | `GO_ENV` | Application environment (`development`, `staging`, `production`) |

## Span Attributes

Standard attributes you should add to spans for better observability:

- **HTTP Requests**:
  - `http.method`: HTTP method
  - `http.url`: Full URL
  - `http.status_code`: Response status code
  - `http.user_agent`: Client user agent

- **Database Operations**:
  - `db.system`: Database type (postgres, mysql, etc.)
  - `db.operation`: Operation type (SELECT, INSERT, etc.)
  - `db.statement`: SQL statement (be careful with PII)
  - `db.name`: Database name

- **Messaging**:
  - `messaging.system`: Messaging system (rabbitmq, kafka, etc.)
  - `messaging.destination`: Queue or topic name
  - `messaging.destination_kind`: Queue, topic, etc.
  - `messaging.operation`: send or receive

## Best Practices

1. **Name spans meaningfully**:
   ```go
   // Good
   ctx, span := tracer.Start(ctx, "orders.process_payment")
   
   // Avoid
   ctx, span := tracer.Start(ctx, "function1")
   ```

2. **Add relevant attributes**:
   ```go
   // Add business-relevant attributes
   span.SetAttributes(
       attribute.String("order.id", orderID),
       attribute.Float64("order.amount", amount),
       attribute.String("payment.provider", provider),
   )
   ```

3. **Record errors properly**:
   ```go
   if err != nil {
       span.RecordError(err)
       span.SetStatus(codes.Error, err.Error())
   }
   ```

4. **Propagate context**:
   Always pass the traced context through your call chain:
   ```go
   result, err := database.QueryWithContext(ctx, query)
   ```

5. **Don't overinstrument**:
   Not every function needs a span. Focus on:
   - Service boundaries
   - Significant operations
   - Long-running operations
   - Error-prone areas

## Observability Integration

For complete observability, integrate with logging and metrics:

```go
// Configure everything with ConfigsBuilder
cfgs, _ := configsBuilder.NewConfigsBuilder().
    Otlp().    // Enables tracing, metrics, and logging with OTLP
    HTTP().    // HTTP configuration if needed
    Build()

// Now you have consistent configuration across all observability signals
```
