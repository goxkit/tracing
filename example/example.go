// Copyright (c) 2025 The GoKit Authors
// MIT License
// All rights reserved.

// Package main provides a simple example of using the Goxkit tracing package.
// This example shows how to set up basic tracing with OpenTelemetry.
package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	// Initialize basic OpenTelemetry setup
	// In a real application, you would configure exporters and providers
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)

	// Get a tracer from the provider
	tracer := otel.GetTracerProvider().Tracer("example-service")

	// Create a root span
	ctx, rootSpan := tracer.Start(context.Background(), "main-operation")
	defer rootSpan.End()

	// Add attributes to the span for better filtering and analysis
	rootSpan.SetAttributes(
		attribute.String("environment", "development"),
		attribute.String("version", "1.0.0"),
	)

	// Log with trace context for correlation in observability platforms
	log.Printf("Application started")

	// Create a child span for a sub-operation
	ctx, childSpan := tracer.Start(ctx, "sub-operation")

	// Simulate some work
	result, err := performOperation(ctx)

	// Record the result in the span
	if err != nil {
		// Record error and set error status on the span
		childSpan.RecordError(err)
		childSpan.SetStatus(codes.Error, err.Error())

		log.Printf("Operation failed: %v", err)
	} else {
		// Add result as an attribute
		childSpan.SetAttributes(attribute.String("result", result))

		log.Printf("Operation succeeded")
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

	// Simulate work time
	time.Sleep(100 * time.Millisecond)

	// Simulate work result
	result := "operation-completed-successfully"

	// Add another event
	span.AddEvent("operation.completed")

	return result, nil
}