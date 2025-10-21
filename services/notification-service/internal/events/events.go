package events

import (
	"encoding/json"
	"fmt"
)

// ประเภทอีเวนต์หลัก ๆ ที่เราจะรองรับ (สามารถขยายได้)
const (
	RKBookingCreated   = "booking.created"
	RKBookingConfirmed = "booking.confirmed"
	RKBookingCancelled = "booking.cancelled"

	RKPaymentPaid   = "payment.paid"
	RKPaymentFailed = "payment.failed"
)

// BookingCreated พกข้อมูลให้พอสำหรับข้อความแจ้งเตือน
type BookingCreated struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
	CourtID   string `json:"court_id"`
	Start     int64  `json:"start"` // unix seconds
	End       int64  `json:"end"`
}

type BookingSimple struct {
	BookingID string `json:"booking_id"`
}

// PaymentPaid / PaymentFailed
type PaymentPaid struct {
	BookingID string `json:"booking_id"`
	ChargeID  string `json:"charge_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}

type PaymentFailed struct {
	BookingID      string `json:"booking_id"`
	ChargeID       string `json:"charge_id"`
	FailureCode    string `json:"failure_code,omitempty"`
	FailureMessage string `json:"failure_message,omitempty"`
}

func MustUnmarshal[T any](b []byte) (T, error) {
	var t T
	if err := json.Unmarshal(b, &t); err != nil {
		var zero T
		return zero, fmt.Errorf("decode payload failed: %w", err)
	}
	return t, nil
}
