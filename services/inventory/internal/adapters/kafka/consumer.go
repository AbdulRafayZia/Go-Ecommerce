package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"gocommerce/pkg/logger"
	"gocommerce/services/inventory/internal/application"
)

// Consumer handles consuming messages from Kafka
type Consumer struct {
	reader           *kafka.Reader
	inventoryService *application.InventoryService
	logger           *logger.Logger
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(
	brokers []string,
	groupID string,
	topics []string,
	inventoryService *application.InventoryService,
	log *logger.Logger,
) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topics[0], // For now, we'll handle multiple topics differently if needed
		MinBytes:       10e3,      // 10KB
		MaxBytes:       10e6,      // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &Consumer{
		reader:           reader,
		inventoryService: inventoryService,
		logger:           log,
	}
}

// Start begins consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("Starting Kafka consumer...")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping Kafka consumer...")
			return c.reader.Close()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				c.logger.ErrorWithErr(err, "Failed to fetch message")
				continue
			}

			c.logger.Infof("Received message from topic %s, partition %d, offset %d",
				msg.Topic, msg.Partition, msg.Offset)

			// Process message
			if err := c.processMessage(ctx, msg); err != nil {
				c.logger.ErrorWithErr(err, "Failed to process message")
				// Don't commit if processing failed
				continue
			}

			// Commit message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.ErrorWithErr(err, "Failed to commit message")
			}
		}
	}
}

// processMessage routes the message to the appropriate handler based on topic
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case "order.created":
		return c.handleOrderCreated(ctx, msg.Value)
	case "order.paid":
		return c.handleOrderPaid(ctx, msg.Value)
	case "order.cancelled":
		return c.handleOrderCancelled(ctx, msg.Value)
	default:
		c.logger.Warnf("Unknown topic: %s", msg.Topic)
		return nil
	}
}

// handleOrderCreated processes order.created events by reserving stock
func (c *Consumer) handleOrderCreated(ctx context.Context, payload []byte) error {
	var event OrderCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order.created event: %w", err)
	}

	c.logger.Infof("Processing order.created event for order %s", event.OrderID)

	// Reserve stock for each item in the order
	for _, item := range event.Items {
		_, err := c.inventoryService.ReserveStockForOrder(
			ctx,
			event.OrderID,
			item.ProductID,
			item.Quantity,
		)
		if err != nil {
			c.logger.ErrorWithErr(err, fmt.Sprintf(
				"Failed to reserve stock for order %s, product %s",
				event.OrderID, item.ProductID,
			))
			// TODO: Publish a stock.reservation.failed event so Order Service can cancel the order
			return err
		}
	}

	c.logger.Infof("Successfully reserved stock for order %s", event.OrderID)
	return nil
}

// handleOrderPaid processes order.paid events by fulfilling reservations
func (c *Consumer) handleOrderPaid(ctx context.Context, payload []byte) error {
	var event OrderPaidEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order.paid event: %w", err)
	}

	c.logger.Infof("Processing order.paid event for order %s", event.OrderID)

	// Fulfill the reservation (deduct from reserved stock)
	if err := c.inventoryService.FulfillReservation(ctx, event.OrderID); err != nil {
		c.logger.ErrorWithErr(err, fmt.Sprintf(
			"Failed to fulfill reservation for order %s",
			event.OrderID,
		))
		return err
	}

	c.logger.Infof("Successfully fulfilled reservation for order %s", event.OrderID)
	return nil
}

// handleOrderCancelled processes order.cancelled events by releasing reservations
func (c *Consumer) handleOrderCancelled(ctx context.Context, payload []byte) error {
	var event OrderCancelledEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order.cancelled event: %w", err)
	}

	c.logger.Infof("Processing order.cancelled event for order %s", event.OrderID)

	// Cancel the reservation (release reserved stock back to available)
	if err := c.inventoryService.CancelReservation(ctx, event.OrderID); err != nil {
		c.logger.ErrorWithErr(err, fmt.Sprintf(
			"Failed to cancel reservation for order %s",
			event.OrderID,
		))
		return err
	}

	c.logger.Infof("Successfully cancelled reservation for order %s", event.OrderID)
	return nil
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
