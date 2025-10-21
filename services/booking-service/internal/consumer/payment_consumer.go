package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/you/badminton-booking/pkg/mq"
	"github.com/you/badminton-booking/services/booking-service/internal/repository"
)

type PaymentPaid struct {
	Event   string `json:"event"`   // "payment.paid"
	Version int    `json:"version"` // 1
	Data    struct {
		PaymentID string `json:"payment_id"`
		BookingID string `json:"booking_id"`
		Amount    int64  `json:"amount"`
		Currency  string `json:"currency"`
		Method    string `json:"method"`
		IdemKey   string `json:"idempotency_key"`
	} `json:"data"`
}

type PaymentConsumer struct {
	repo *repository.BookingRepo
	cons *mq.Consumer
}

func NewPaymentConsumer(repo *repository.BookingRepo, cons *mq.Consumer) *PaymentConsumer {
	return &PaymentConsumer{repo: repo, cons: cons}
}

func (pc *PaymentConsumer) Run(ctx context.Context) error {
	msgs, err := pc.cons.Deliveries(ctx)
	if err != nil {
		return err
	}
	go func() {
		for d := range msgs {
			switch d.RoutingKey {
			case "payment.paid":
				var evt PaymentPaid
				if err := json.Unmarshal(d.Body, &evt); err != nil {
					log.Printf("[booking-consumer] unmarshal error: %v", err)
					_ = d.Nack(false, false)
					continue
				}
				if evt.Data.BookingID == "" || evt.Data.PaymentID == "" {
					log.Printf("[booking-consumer] invalid event payload")
					_ = d.Ack(false)
					continue
				}
				if _, err := pc.repo.ConfirmIfNotProcessed(ctx, evt.Data.BookingID, evt.Data.PaymentID, "payment.paid"); err != nil {
					log.Printf("[booking-consumer] confirm error: %v", err)
					_ = d.Nack(false, true)
					continue
				}
				_ = d.Ack(false)
			default:
				// ignore others
				_ = d.Ack(false)
			}
		}
	}()
	return nil
}
