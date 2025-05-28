// Copyright (c) 2025 The GoKit Authors
// MIT License
// All rights reserved.

// Package examples provides usage examples for the Goxkit tracing module.
package examples

import (
	"context"
	"time"

	configsBuilder "github.com/goxkit/configs_builder"
	"github.com/goxkit/tracing/amqp"
	amqplib "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// AMQPExample demonstrates how to use tracing with AMQP (RabbitMQ)
// to propagate context between publisher and consumer.
func AMQPExample() {
	// Initialize configurations with tracing enabled
	cfg, err := configsBuilder.NewConfigsBuilder().
		Otlp().     // Enable OpenTelemetry
		RabbitMQ(). // Configure RabbitMQ
		Build()
	if err != nil {
		panic(err)
	}

	// Connect to RabbitMQ (simplified - in real code you'd use the actual connection)
	conn, err := amqplib.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		cfg.Logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		return
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		cfg.Logger.Error("Failed to open a channel", zap.Error(err))
		return
	}
	defer ch.Close()

	// Queue declaration (simplified)
	queueName := "orders"
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		cfg.Logger.Error("Failed to declare queue", zap.Error(err))
		return
	}

	// Get a tracer from the global provider
	tracer := otel.Tracer("messaging")

	// Prepare message payload
	orderPayload := []byte(`{"order_id":"123","customer":"customer-456","amount":99.99}`)

	// ----------------- PUBLISHER SIDE -----------------

	// Create a context and a span for the publish operation
	ctx, span := tracer.Start(context.Background(), "publish.order")
	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", queueName),
		attribute.String("messaging.operation", "publish"),
	)
	defer span.End()

	// Prepare AMQP message
	msg := amqplib.Publishing{
		ContentType: "application/json",
		Body:        orderPayload,
		Headers:     amqplib.Table{},
	}

	// Inject trace context into message headers
	amqp.AMQPPropagator.Inject(ctx, amqp.AMQPHeader(msg.Headers))

	// Log with trace context
	cfg.Logger.Info("Publishing message",
		zap.String("queue", queueName),
		zap.Any("context", ctx),
	)

	// Publish the message
	err = ch.PublishWithContext(
		ctx,       // context
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		msg,       // message
	)
	if err != nil {
		cfg.Logger.Error("Failed to publish message", zap.Error(err))
		return
	}

	// ----------------- CONSUMER SIDE -----------------

	// Normally this would be in a separate process/service
	// This is just for demonstration

	// Simulate message delivery
	delivery := amqplib.Delivery{
		Body:    orderPayload,
		Headers: msg.Headers,
	}

	// Extract trace context from headers and create a new span
	consumerCtx, consumerSpan := amqp.NewConsumerSpan(tracer, delivery.Headers, queueName)
	consumerSpan.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", queueName),
		attribute.String("messaging.operation", "receive"),
	)
	defer consumerSpan.End()

	// Log with trace context
	cfg.Logger.Info("Received message",
		zap.String("queue", queueName),
		zap.Any("context", ctx),
	)

	// Process the message with the traced context
	processOrder(consumerCtx, delivery.Body)
}

// processOrder simulates processing an order
func processOrder(ctx context.Context, orderData []byte) {
	// Get a tracer
	tracer := otel.Tracer("orders")

	// Create a child span for the processing
	ctx, span := tracer.Start(ctx, "process.order")
	defer span.End()

	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	// Add processing details to the span
	span.SetAttributes(
		attribute.Int("order.items", 3),
		attribute.Float64("order.total", 99.99),
	)

	// Add events to mark processing stages
	span.AddEvent("order.validated")
	time.Sleep(10 * time.Millisecond)

	span.AddEvent("payment.processed")
	time.Sleep(20 * time.Millisecond)

	span.AddEvent("order.completed")
}
