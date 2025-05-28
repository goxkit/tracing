// Copyright (c) 2025 The GoKit Authors
// MIT License
// All rights reserved.

// Package examples provides usage examples for the Goxkit tracing module.
package examples

import (
	"context"
	"time"

	configsBuilder "github.com/goxkit/configs_builder"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// ConfigsBuilderBasicExample demonstrates how to set up tracing with ConfigsBuilder
// and create simple spans for operations.
func ConfigsBuilderBasicExample() {
	// Set up environment variables (optional - typically done in deployment)
	// os.Setenv("APP_NAME", "ExampleService")
	// os.Setenv("APP_NAMESPACE", "examples")
	// os.Setenv("GO_ENV", "development")
	// os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	// Use the ConfigsBuilder to set up your application with OTLP enabled
	cfg, err := configsBuilder.NewConfigsBuilder().
		Otlp(). // Enable OpenTelemetry tracing, metrics, and logging
		HTTP(). // Add HTTP configuration if needed
		Build()
	if err != nil {
		panic(err)
	}

	// Get a tracer from the global provider configured by ConfigsBuilder
	tracer := otel.GetTracerProvider().Tracer("example-service")

	// Create a root span
	ctx, rootSpan := tracer.Start(context.Background(), "process-request")
	defer rootSpan.End()

	// Add attributes to the span
	rootSpan.SetAttributes(
		attribute.String("customer.id", "cust-123"),
		attribute.String("request.id", "req-456"),
	)

	// Log with trace context
	cfg.Logger.Info("Processing started",
		zap.String("operation", "data-processing"),
		zap.Any("context", ctx),
	)

	// Create a child span for a sub-operation
	childCtx, childSpan := tracer.Start(ctx, "database-query")

	// Simulate some work with the traced context
	time.Sleep(100 * time.Millisecond)

	// End the child span when done
	childSpan.End()

	// Add events to the root span to mark important occurrences
	rootSpan.AddEvent("request.validation.complete")

	// Record errors if they occur
	if err := simulateError(); err != nil {
		rootSpan.RecordError(err)
		rootSpan.SetStatus(codes.Error, err.Error())

		// Log with trace context
		cfg.Logger.Error("Operation failed",
			zap.Error(err),
			zap.Any("context", childCtx),
		)
	} else {
		// Mark the span as successful
		rootSpan.SetStatus(codes.Ok, "")

		cfg.Logger.Info("Operation succeeded",
			zap.Any("context", childCtx),
		)
	}
}

// simulateError is a helper function that randomly returns an error
func simulateError() error {
	return nil // For this example, always succeed
}
