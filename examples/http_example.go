// Copyright (c) 2025 The GoKit Authors
// MIT License
// All rights reserved.

// Package examples provides usage examples for the Goxkit tracing module.
package examples

import (
	"context"
	"fmt"
	"net/http"

	configsBuilder "github.com/goxkit/configs_builder"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// HTTPServerExample demonstrates how to set up an HTTP server with
// automatic tracing instrumentation.
func HTTPServerExample() {
	// Initialize configurations with tracing enabled
	cfg, err := configsBuilder.NewConfigsBuilder().
		Otlp(). // Enable OpenTelemetry
		HTTP(). // Configure HTTP server
		Build()
	if err != nil {
		panic(err)
	}

	// Create a handler for the orders endpoint
	ordersHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A span is automatically created by the otelhttp middleware
		// Get the current span from the request context
		span := trace.SpanFromContext(r.Context())

		// Extract order ID from request (simplified example)
		orderID := r.URL.Query().Get("id")
		if orderID == "" {
			orderID = "unknown"
		}

		// Add custom attributes to the span
		span.SetAttributes(
			attribute.String("order.id", orderID),
			attribute.String("user.id", r.Header.Get("X-User-ID")),
		)

		// Log with trace context
		cfg.Logger.Info("Processing order request",
			zap.String("order_id", orderID),
			zap.String("method", r.Method),
			zap.Any("context", r.Context()),
		)

		// Create a child span for database operation
		ctx, dbSpan := otel.Tracer("orders").Start(r.Context(), "db.query")
		defer dbSpan.End()

		// Simulate database query
		order, err := getOrderDetails(ctx, orderID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		// Return response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Order %s processed successfully: %v", orderID, order)
	})

	// Wrap handler with OpenTelemetry instrumentation
	tracedOrdersHandler := otelhttp.NewHandler(ordersHandler, "orders.get",
		otelhttp.WithTracerProvider(otel.GetTracerProvider()),
	)

	// Register handler
	http.Handle("/api/orders", tracedOrdersHandler)

	// Start server
	port := cfg.AppConfigs.Port
	cfg.Logger.Info("Starting HTTP server", zap.Int("port", port))
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		cfg.Logger.Error("Server failed", zap.Error(err))
	}
}

// getOrderDetails is a mock function to simulate a database query
func getOrderDetails(ctx context.Context, orderID string) (map[string]interface{}, error) {
	// In real code, you would perform a database query, passing the context
	// db.QueryRowContext(ctx, "SELECT * FROM orders WHERE id = ?", orderID)

	// Return mock data
	return map[string]interface{}{
		"id":     orderID,
		"status": "shipped",
		"items":  []string{"item1", "item2"},
		"total":  125.50,
	}, nil
}
