// Copyright (c) 2025 The GoKit Authors
// MIT License
// All rights reserved.

// This example demonstrates using the Goxkit tracing package with ConfigsBuilder
package examples

import (
	"context"

	configsBuilder "github.com/goxkit/configs_builder"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func main() {
	// Initialize with ConfigsBuilder
	// The Otlp() call enables OpenTelemetry for tracing, metrics and logging
	cfgs, err := configsBuilder.NewConfigsBuilder().Otlp().Build()
	if err != nil {
		panic(err)
	}

	// Get a tracer from the global provider that was configured by ConfigsBuilder
	tracer := otel.GetTracerProvider().Tracer("example-service")

	// Create a root span
	ctx, rootSpan := tracer.Start(context.Background(), "main-operation")
	defer rootSpan.End()

	// Add attributes to the span for better filtering and analysis
	rootSpan.SetAttributes(
		attribute.String("environment", cfgs.AppConfigs.Environment.String()),
		attribute.String("version", "1.0.0"),
	)

	// Log with trace context for correlation in observability platforms
	cfgs.Logger.Info("Application started",
		zap.String("service", cfgs.AppConfigs.Name),
		zap.Any("context", ctx))

	// Create a child span for a sub-operation
	ctx, childSpan := tracer.Start(ctx, "sub-operation")

	// Simulate some work
	result, err := performOperation(ctx)

	// Record the result in the span
	if err != nil {
		// Record error and set error status on the span
		childSpan.RecordError(err)
		childSpan.SetStatus(codes.Error, err.Error())

		cfgs.Logger.Error("Operation failed",
			zap.Error(err),
			zap.Any("context", ctx))
	} else {
		// Add result as an attribute
		childSpan.SetAttributes(attribute.String("result", result))

		cfgs.Logger.Info("Operation succeeded",
			zap.String("result", result),
			zap.Any("context", ctx))
	}

	// End the child span
	childSpan.End()
}

// performOperation simulates a traced operation
func performOperation(ctx context.Context) (string, error) {
	// Get tracer (same as in main)
	tracer := otel.GetTracerProvider().Tracer("example-service")

	// Create a nested span
	_, span := tracer.Start(ctx, "perform-operation")
	defer span.End()

	// Add events to the span timeline
	span.AddEvent("operation.started")

	// Simulate work result
	result := "operation-completed-successfully"

	// Add another event
	span.AddEvent("operation.completed")

	return result, nil
}
