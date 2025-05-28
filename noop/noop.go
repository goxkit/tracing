// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package noop

import (
	"github.com/goxkit/configs"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func Install(cfgs *configs.Configs) (*sdktrace.TracerProvider, error) {
	provider := sdktrace.NewTracerProvider()
	cfgs.TracerProvider = provider
	return provider, nil
}
