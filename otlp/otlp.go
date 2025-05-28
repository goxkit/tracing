// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

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
