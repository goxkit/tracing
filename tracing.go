// Copyright (c) 2025 The GoKit Authors
// MIT License
// All rights reserved.

// Package tracing provides distributed tracing capabilities using OpenTelemetry.
// It enables monitoring and troubleshooting of microservices-based applications
// by creating, propagating, and collecting trace data across service boundaries.
//
// The package integrates seamlessly with the configs package to provide
// environment-specific configuration and supports multiple output options:
// - OTLP export for observability platforms (Jaeger, Zipkin, etc.)
// - No-operation mode for testing and development
//
// It also provides utilities for trace context propagation in different protocols
// (HTTP, AMQP) and integration with structured logging via Zap.
package tracing

import (
	"github.com/goxkit/configs"
	"github.com/goxkit/tracing/noop"
	"github.com/goxkit/tracing/otlp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Install initializes and configures a tracer provider based on the application configuration.
// If OTLP is enabled in the configuration, it sets up an OpenTelemetry tracer with
// export to the configured endpoint. Otherwise, it sets up a no-operation tracer
// that satisfies the interface but doesn't collect or export spans.
//
// The configured tracer provider is stored in the configs object and also set as
// the global tracer provider for the application.
//
// Parameters:
//   - cfgs: Application configurations including OTLP settings
//
// Returns:
//   - *sdktrace.TracerProvider: The configured tracer provider
//   - error: Any error encountered during setup
func Install(cfgs *configs.Configs) (*sdktrace.TracerProvider, error) {
	if cfgs.OTLPConfigs.Enabled {
		return otlp.Install(cfgs)
	}

	return noop.Install(cfgs)
}
