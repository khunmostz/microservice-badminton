package mq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	exchange string
	queue    string
	keys     []string
}

func NewConsumer(url, exchange, queue string, keys []string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}
	for _, rk := range keys {
		if err := ch.QueueBind(q.Name, rk, exchange, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return nil, fmt.Errorf("bind %s: %w", rk, err)
		}
	}
	return &Consumer{conn: conn, ch: ch, exchange: exchange, queue: q.Name, keys: keys}, nil
}

func (c *Consumer) Deliveries(ctx context.Context) (<-chan amqp.Delivery, error) {
	return c.ch.ConsumeWithContext(ctx, c.queue, "", false, false, false, false, nil)
}

func (c *Consumer) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
