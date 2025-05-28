// Copyright (c) 2025 The GoKit Authors
// MIT License
// All rights reserved.

// Package zap provides integration between distributed tracing and structured logging.
// It enables automatic inclusion of trace context (trace IDs and span IDs) in log entries,
// facilitating correlation between traces and logs in observability platforms.
package zap

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// traceLog is a struct that holds trace and span IDs for structured logging.
// It implements zapcore.ObjectMarshaler to enable trace context to be
// included in zap log entries as a structured object.
type traceLog struct {
	// TraceID is the string representation of the trace identifier
	TraceID string

	// SpanID is the string representation of the span identifier
	SpanID string
}

// MarshalLogObject implements zapcore.ObjectMarshaler interface for traceLog.
// This allows the trace information to be added to zap logs in a structured way,
// making it easy to correlate logs with traces in observability platforms.
//
// Parameters:
//   - enc: The zap object encoder to add the trace fields to
//
// Returns:
//   - error: Always nil for this implementation
func (u *traceLog) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("trace_id", u.TraceID)
	enc.AddString("span_id", u.SpanID)
	return nil
}

// Format extracts trace and span IDs from a context and returns them as a zap field
// for structured logging. If no active span is found in the context, it returns a Skip field.
// This function allows easy inclusion of trace context in log entries.
//
// Example usage:
//
//	logger.Info("Processing request", tracing.Format(ctx), zap.String("user_id", userID))
//
// Parameters:
//   - ctx: The context containing the trace information
//
// Returns:
//   - zapcore.Field: A zap field containing the trace and span IDs, or a Skip field if no span is present
func Format(ctx context.Context) zapcore.Field {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return zap.Skip()
	}

	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	return zap.Inline(&traceLog{traceID, spanID})
}
