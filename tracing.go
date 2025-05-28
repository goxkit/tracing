// Copyright (c) 2025 The GoKit Authors
// MIT License
// All rights reserved.

// Package tracing provides distributed tracing capabilities using OpenTelemetry
// to help monitor and troubleshoot microservices-based applications.
package tracing

import (
	"github.com/goxkit/configs"
	"github.com/goxkit/tracing/noop"
	"github.com/goxkit/tracing/otlp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func Install(cfgs *configs.Configs) (*sdktrace.TracerProvider, error) {
	if cfgs.OTLPConfigs.Enabled {
		return otlp.Install(cfgs)
	}

	return noop.Install(cfgs)
}
