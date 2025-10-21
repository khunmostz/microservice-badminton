package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/you/badminton-booking/services/booking-service/internal/domain"
)

var ErrOverlap = errors.New("slot_overlapped")

type BookingRepo struct{ db *gorm.DB }

func NewBookingRepo(db *gorm.DB) *BookingRepo {
	return &BookingRepo{db: db}
}
func (r *BookingRepo) Migrate() error {
	return r.db.AutoMigrate(&domain.Booking{}, &domain.EventConsumed{})
}

// CreateWithNoOverlap runs in a txn and prevents overlapping bookings by locking rows.
func (r *BookingRepo) CreateWithNoOverlap(ctx context.Context, b *domain.Booking) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock any candidate rows that would overlap to avoid races
		var existing domain.Booking
		err := tx.Model(&domain.Booking{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("court_id = ? AND status IN ?", b.CourtID, []string{"PENDING", "CONFIRMED"}).
			Where("start_time < ? AND end_time > ?", b.EndTime, b.StartTime). // overlap condition
			Take(&existing).Error

		if err == nil {
			return ErrOverlap
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if b.ID == "" {
			b.ID = uuid.NewString()
		}
		return tx.Create(b).Error
	})
}

func (r *BookingRepo) ByID(ctx context.Context, id string) (*domain.Booking, error) {
	var b domain.Booking
	if err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, id, to string) (*domain.Booking, error) {
	var b domain.Booking
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.First(&b, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	b.Status = to
	if err := tx.Save(&b).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return &b, tx.Commit().Error
}

// IdempotentConfirm: ใช้ตอน consume payment.paid
func (r *BookingRepo) ConfirmIfNotProcessed(ctx context.Context, bookingID, eventID, eventKey string) (*domain.Booking, error) {
	var b domain.Booking
	tx := r.db.WithContext(ctx).Begin()

	// 1) ถ้า event เคยถูกกินแล้ว → ข้าม (idempotent)
	var exists int64
	if err := tx.Model(&domain.EventConsumed{}).Where("id = ?", eventID).Count(&exists).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if exists > 0 {
		// คืนสถานะปัจจุบัน
		if err := tx.First(&b, "id = ?", bookingID).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		tx.Commit()
		return &b, nil
	}

	// 2) อัปเดต booking -> CONFIRMED
	if err := tx.First(&b, "id = ?", bookingID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	// ถ้า Cancelled ไปแล้ว อาจเลือกข้าม (business rule)
	if b.Status != "CONFIRMED" {
		b.Status = "CONFIRMED"
		if err := tx.Save(&b).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 3) บันทึก event_consumed กันซ้ำ
	rec := domain.EventConsumed{ID: eventID, EventKey: eventKey, ProcessedAt: time.Now().UTC()}
	if err := tx.Create(&rec).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &b, tx.Commit().Error
}

func (r *BookingRepo) List(ctx context.Context, page, size int32, userID, courtID, dayISO string) ([]domain.Booking, int64, error) {
	if size <= 0 {
		size = 20
	}
	if page < 0 {
		page = 0
	}
	qb := r.db.WithContext(ctx).Model(&domain.Booking{})
	if userID != "" {
		qb = qb.Where("user_id = ?", userID)
	}
	if courtID != "" {
		qb = qb.Where("court_id = ?", courtID)
	}
	if dayISO != "" {
		if d, err := time.Parse(time.RFC3339, dayISO); err == nil {
			from := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
			to := from.Add(24 * time.Hour)
			qb = qb.Where("start_time < ? AND end_time > ?", to, from) // any overlap with that day
		}
	}
	var total int64
	if err := qb.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []domain.Booking
	if err := qb.Order("start_time ASC").Limit(int(size)).Offset(int(page * size)).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
