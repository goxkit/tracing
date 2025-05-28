// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

// Package noop provides a no-operation tracer implementation for the tracing package.
// This package creates a minimal tracer provider that satisfies the tracing interface
// but doesn't collect or export any spans. It's useful for development environments,
// testing, or situations where tracing needs to be disabled without changing code structure.
package noop

import (
	"github.com/goxkit/configs"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Install initializes and returns a minimal no-operation tracer provider.
// The provider is created with default settings and no exporters, meaning
// it will not collect or export any spans. This is useful when tracing
// overhead should be avoided but code that expects a tracer should still work.
//
// The tracer provider is stored in the configs object for use throughout
// the application.
//
// Parameters:
//   - cfgs: Application configurations to store the tracer provider
//
// Returns:
//   - *sdktrace.TracerProvider: A minimal tracer provider with no exporters
//   - error: Always nil for the noop implementation
func Install(cfgs *configs.Configs) (*sdktrace.TracerProvider, error) {
	provider := sdktrace.NewTracerProvider()
	cfgs.TracerProvider = provider
	return provider, nil
}
