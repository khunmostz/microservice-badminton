package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/you/badminton-booking/services/notification-service/internal/events"
	"github.com/you/badminton-booking/services/notification-service/internal/notifier"
)

type Config struct {
	RabbitURL string
	// เดิมมี Exchange เดียว; เพิ่ม Exchanges เพื่อรองรับหลายตัว
	Exchange    string   // ใช้เพื่อความเข้ากันได้ย้อนหลัง
	Exchanges   []string // ถ้ามีค่า จะใช้ค่านี้แทน Exchange
	Queue       string
	Bindings    []string
	Prefetch    int
	UseDLX      bool
	DLXName     string
	DLXQueue    string
	ServiceName string
}

type Consumer struct {
	cfg      Config
	notifier notifier.Notifier

	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewConsumer(cfg Config, n notifier.Notifier) *Consumer {
	return &Consumer{cfg: cfg, notifier: n}
}

func (c *Consumer) RabbitURL() string {
	if v := os.Getenv("RABBIT_URL"); v != "" {
		return v
	}
	return c.cfg.RabbitURL
}

func (c *Consumer) Connect() error {
	conn, err := amqp.Dial(c.RabbitURL())
	if err != nil {
		return fmt.Errorf("rabbit dial failed: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open channel failed: %w", err)
	}

	// เลือกรายการ exchanges ที่จะใช้
	exchanges := c.cfg.Exchanges
	if len(exchanges) == 0 {
		if c.cfg.Exchange != "" {
			exchanges = []string{c.cfg.Exchange}
		} else {
			exchanges = []string{"booking.exchange"}
		}
	}

	// declare queue (พร้อม DLX ถ้าต้องการ)
	args := amqp.Table{}
	if c.cfg.UseDLX {
		args["x-dead-letter-exchange"] = c.cfg.DLXName
	}

	q, err := ch.QueueDeclare(c.cfg.Queue, true, false, false, false, args)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("declare queue failed: %w", err)
	}

	// declare + bind กับ "ทุก exchange"
	for _, ex := range exchanges {
		if err := ch.ExchangeDeclare(ex, "topic", true, false, false, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return fmt.Errorf("declare exchange %s failed: %w", ex, err)
		}
		for _, key := range c.cfg.Bindings {
			if err := ch.QueueBind(q.Name, key, ex, false, nil); err != nil {
				_ = ch.Close()
				_ = conn.Close()
				return fmt.Errorf("bind queue to exchange=%s key=%s failed: %w", ex, key, err)
			}
		}
	}

	// DLX/ DLQ (optional)
	if c.cfg.UseDLX {
		if err := ch.ExchangeDeclare(c.cfg.DLXName, "topic", true, false, false, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return fmt.Errorf("declare dlx failed: %w", err)
		}
		if _, err := ch.QueueDeclare(c.cfg.DLXQueue, true, false, false, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return fmt.Errorf("declare dlq failed: %w", err)
		}
		if err := ch.QueueBind(c.cfg.DLXQueue, "#", c.cfg.DLXName, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return fmt.Errorf("bind dlq failed: %w", err)
		}
	}

	if c.cfg.Prefetch <= 0 {
		c.cfg.Prefetch = 8
	}
	if err := ch.Qos(c.cfg.Prefetch, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("set qos failed: %w", err)
	}

	c.conn = conn
	c.ch = ch
	return nil
}

func (c *Consumer) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	msgs, err := c.ch.ConsumeWithContext(ctx, c.cfg.Queue, c.cfg.ServiceName, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume failed: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := c.handleDelivery(d); err != nil {
				log.Printf("[notify] handle error key=%s err=%v -> Nack&requeue", d.RoutingKey, err)
				_ = d.Nack(false, true)
				continue
			}
			_ = d.Ack(false)
		}
	}
}

func (c *Consumer) handleDelivery(d amqp.Delivery) error {
	key := d.RoutingKey
	body := d.Body

	switch key {
	case events.RKBookingCreated:
		ev, err := events.MustUnmarshal[events.BookingCreated](body)
		if err != nil {
			return err
		}
		return c.notifier.Notify("📅 Booking Created",
			fmt.Sprintf("Booking %s (court=%s) %s", ev.BookingID, ev.CourtID, notifier.HumanTimeRange(ev.Start, ev.End)))

	case events.RKBookingConfirmed:
		ev, err := events.MustUnmarshal[events.BookingSimple](body)
		if err != nil {
			return err
		}
		return c.notifier.Notify("✅ Booking Confirmed",
			fmt.Sprintf("Booking %s has been confirmed.", ev.BookingID))

	case events.RKBookingCancelled:
		ev, err := events.MustUnmarshal[events.BookingSimple](body)
		if err != nil {
			return err
		}
		return c.notifier.Notify("❌ Booking Cancelled",
			fmt.Sprintf("Booking %s has been cancelled.", ev.BookingID))

	case events.RKPaymentPaid:
		ev, err := events.MustUnmarshal[events.PaymentPaid](body)
		if err != nil {
			return err
		}
		return c.notifier.Notify("💰 Payment Paid",
			fmt.Sprintf("Booking %s paid %d %s (charge=%s).", ev.BookingID, ev.Amount, strings.ToUpper(ev.Currency), ev.ChargeID))

	case events.RKPaymentFailed:
		ev, err := events.MustUnmarshal[events.PaymentFailed](body)
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("Payment failed for booking %s (charge=%s).", ev.BookingID, ev.ChargeID)
		if ev.FailureCode != "" || ev.FailureMessage != "" {
			msg = fmt.Sprintf("%s Reason: %s %s", msg, ev.FailureCode, ev.FailureMessage)
		}
		return c.notifier.Notify("⚠️ Payment Failed", msg)

	default:
		// ไม่รู้จัก key — แค่บันทึกแล้วรับไว้ (หรือจะ Nack ก็ได้)
		log.Printf("[notify] skip unknown key=%s", key)
	}
	return nil
}
