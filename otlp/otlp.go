// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

// Package otlp provides OpenTelemetry Protocol (OTLP) integration for tracing.
// It enables trace data to be exported to observability platforms like Jaeger,
// Zipkin, or any other OpenTelemetry-compatible collector. This package handles
// the setup and configuration of the OTLP exporter and tracer provider.
package otlp

import (
	"context"
	"time"

	"github.com/goxkit/configs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

// Install configures and initializes an OpenTelemetry tracer provider that exports
// trace data via OTLP to a collector. It sets up the connection to the OTLP endpoint
// specified in the configuration and configures the tracer with proper service and
// environment attributes for better observability context.
//
// The function handles the complete setup of:
// - OTLP exporter with gRPC transport
// - Batch processing for efficient span export
// - Resource attributes for service identification
// - Global tracer provider registration
// - W3C TraceContext propagation
//
// Parameters:
//   - cfgs: Application configurations including OTLP endpoint and service information
//
// Returns:
//   - *sdktrace.TracerProvider: The configured tracer provider with OTLP export capabilities
//   - error: Any error encountered during setup
func Install(cfgs *configs.Configs) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()
	exp, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(cfgs.OTLPConfigs.Endpoint),
			otlptracegrpc.WithReconnectionPeriod(cfgs.OTLPConfigs.ExporterReconnectionPeriod),
			otlptracegrpc.WithTimeout(cfgs.OTLPConfigs.ExporterTimeout),
			otlptracegrpc.WithCompressor("gzip"),
			otlptracegrpc.WithDialOption(
				grpc.WithConnectParams(grpc.ConnectParams{
					Backoff: backoff.Config{
						BaseDelay:  1 * time.Second,
						Multiplier: 1.6,
						MaxDelay:   15 * time.Second,
					},
					MinConnectTimeout: 0,
				}),
			),
			otlptracegrpc.WithInsecure()),
	)
	if err != nil {
		cfgs.Logger.Error("failed to create OTLP trace exporter", zap.Error(err))
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfgs.AppConfigs.Name),
			semconv.ServiceNamespaceKey.String(cfgs.AppConfigs.Namespace),
			attribute.String("service.environment", cfgs.AppConfigs.Environment.String()),
			semconv.DeploymentEnvironmentKey.String(cfgs.AppConfigs.Environment.String()),
			semconv.TelemetrySDKLanguageKey.String("go"),
			semconv.TelemetrySDKLanguageGo.Key.Bool(true),
		)),
		sdktrace.WithSpanProcessor(bsp),
	)

	cfgs.TracerProvider = tracerProvider
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider, nil
}
