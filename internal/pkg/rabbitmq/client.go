package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/RuLap/sportmates-api/internal/pkg/config"
	"github.com/RuLap/sportmates-api/internal/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	log     *slog.Logger
}

func NewClient(cfg *config.RabbitMQConfig, log *slog.Logger) (*Client, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		"events",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: channel,
		log:     log,
	}, nil
}

func (c *Client) PublishEvent(event events.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.channel.PublishWithContext(
		ctx,
		"",
		"events",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Type:         event.GetType(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	c.log.Debug("event published", "type", event.GetType())
	return nil
}

func (c *Client) PublishEmailEvent(event events.EmailEvent) error {
	return c.PublishEvent(event)
}

func (c *Client) ConsumeEvents(ctx context.Context, handler func(eventType string, body []byte) error) error {
	msgs, err := c.channel.Consume(
		"events",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	c.log.Info("started consuming events")

	for {
		select {
		case <-ctx.Done():
			c.log.Info("stopping events consumer")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				c.log.Warn("events channel closed")
				return nil
			}

			if err := handler(msg.Type, msg.Body); err != nil {
				c.log.Error("failed to handle event", "type", msg.Type, "error", err)
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
				c.log.Debug("event processed", "type", msg.Type)
			}
		}
	}
}

func (c *Client) ConsumeEmailEvents(ctx context.Context, handler func(events.EmailEvent) error) error {
	return c.ConsumeEvents(ctx, func(eventType string, body []byte) error {
		if eventType != "email" {
			return nil
		}

		var event events.EmailEvent
		if err := json.Unmarshal(body, &event); err != nil {
			return fmt.Errorf("failed to unmarshal email event: %w", err)
		}

		return handler(event)
	})
}
func (c *Client) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.log.Error("failed to close channel", "error", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.log.Error("failed to close connection", "error", err)
		}
	}
	return nil
}

func (c *Client) HealthCheck() error {
	if c.conn == nil || c.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	return nil
}
