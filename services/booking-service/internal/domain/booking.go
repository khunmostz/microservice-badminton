package domain

import "time"

type Booking struct {
	ID        string    `gorm:"primaryKey"`
	UserID    string    `gorm:"index"`
	CourtID   string    `gorm:"index"`
	StartTime time.Time `gorm:"index"`
	EndTime   time.Time `gorm:"index"`
	Status    string    `gorm:"index"` // PENDING|CONFIRMED|CANCELLED
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EventConsumed struct {
	ID          string `gorm:"primaryKey"` // event unique id (e.g. payment_id or composed key)
	EventKey    string `gorm:"index"`      // e.g. payment.paid
	ProcessedAt time.Time
}
