// services/payment-service/internal/http/webhook.go
package httpx

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
	"github.com/you/badminton-booking/pkg/mq"
)

type WebhookServer struct {
	omc           *omise.Client
	publisher     *mq.Publisher
	Exchange      string
	RoutingPaid   string
	RoutingFailed string
}

func NewWebhookServer(omc *omise.Client, pub *mq.Publisher, exchange string) *WebhookServer {
	return &WebhookServer{
		omc:           omc,
		publisher:     pub,
		Exchange:      exchange,
		RoutingPaid:   "payment.paid",
		RoutingFailed: "payment.failed",
	}
}

type incomingEvent struct {
	ID   string          `json:"id"`
	Key  string          `json:"key"`
	Data json.RawMessage `json:"data"`
}

type PaymentPaid struct {
	Event      string `json:"event"`       // "payment.paid"
	Version    int    `json:"version"`     // 1
	OccurredAt string `json:"occurred_at"` // RFC3339
	Data       struct {
		PaymentID string `json:"payment_id"`
		BookingID string `json:"booking_id"`
		Amount    int64  `json:"amount"`
		Currency  string `json:"currency"`
		Method    string `json:"method"`
		IdemKey   string `json:"idempotency_key"`
	} `json:"data"`
}

type PaymentFailed struct {
	Event      string `json:"event"` // "payment.failed"
	Version    int    `json:"version"`
	OccurredAt string `json:"occurred_at"`
	Data       struct {
		PaymentID string `json:"payment_id"`
		BookingID string `json:"booking_id"`
		Reason    string `json:"reason"`
	} `json:"data"`
}

func (s *WebhookServer) Handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var inc incomingEvent
	if err := json.NewDecoder(r.Body).Decode(&inc); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 1) ยืนยันเหตุการณ์กับ Omise โดยดึง Event อีกรอบ
	ev := &omise.Event{}
	if err := s.omc.Do(ev, &operations.RetrieveEvent{EventID: inc.ID}); err != nil {
		log.Printf("[webhook] retrieve event error: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("ev.Key: %s", ev.Key)

	switch ev.Key {
	case "charge.complete":
		// ev.Data เป็น interface{} → marshal ก่อนแล้วค่อย unmarshal เป็น Charge
		raw, err := json.Marshal(ev.Data)
		if err != nil {
			log.Printf("[webhook] marshal ev.Data error: %v", err)
			w.WriteHeader(http.StatusOK)
			return
		}
		var ch omise.Charge
		if err := json.Unmarshal(raw, &ch); err != nil {
			log.Printf("[webhook] unmarshal charge error: %v", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		bookingID, _ := ch.Metadata["booking_id"].(string)

		if ch.Status == "successful" {
			evt := PaymentPaid{
				Event:      "payment.paid",
				Version:    1,
				OccurredAt: time.Now().UTC().Format(time.RFC3339),
			}
			evt.Data.PaymentID = ch.ID
			evt.Data.BookingID = bookingID
			evt.Data.Amount = ch.Amount
			evt.Data.Currency = ch.Currency
			// method: ถ้ามี source ให้ใช้ type, ไม่งั้นถือเป็น card
			if ch.Source != nil && ch.Source.Type != "" {
				evt.Data.Method = ch.Source.Type
			} else {
				evt.Data.Method = "card"
			}
			evt.Data.IdemKey = bookingID

			if err := s.publisher.PublishJSON(context.Background(), s.RoutingPaid, evt); err != nil {
				log.Printf("[webhook] publish payment.paid error: %v", err)
			}
			log.Print("publish payment.paid")
		} else {
			evt := PaymentFailed{
				Event:      "payment.failed",
				Version:    1,
				OccurredAt: time.Now().UTC().Format(time.RFC3339),
			}
			evt.Data.PaymentID = ch.ID
			evt.Data.BookingID = bookingID
			// FailureCode เป็น *string → ต้องเช็ค nil
			if ch.FailureCode != nil {
				evt.Data.Reason = *ch.FailureCode
			} else {
				evt.Data.Reason = ""
			}
			if err := s.publisher.PublishJSON(context.Background(), s.RoutingFailed, evt); err != nil {
				log.Printf("[webhook] publish payment.failed error: %v", err)
			}
			log.Print("publish payment.failed")
		}
	default:
		// ข้าม event key อื่น
		log.Print("default")
	}

	w.WriteHeader(http.StatusOK)
}
