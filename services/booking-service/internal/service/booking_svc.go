package service

import (
	"context"
	"errors"
	"time"

	"github.com/you/badminton-booking/pkg/mq"
	"github.com/you/badminton-booking/services/booking-service/internal/domain"
	"github.com/you/badminton-booking/services/booking-service/internal/repository"
)

type BookingSvc struct {
	repo *repository.BookingRepo
	pub  *mq.Publisher
}

func NewBookingSvc(r *repository.BookingRepo, pub *mq.Publisher) *BookingSvc {
	return &BookingSvc{repo: r, pub: pub}
}

func parseRFC3339UTC(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func (s *BookingSvc) Create(ctx context.Context, userID, courtID, startISO, endISO string) (*domain.Booking, error) {
	st, err := parseRFC3339UTC(startISO)
	if err != nil {
		return nil, err
	}
	et, err := parseRFC3339UTC(endISO)
	if err != nil {
		return nil, err
	}
	if !et.After(st) {
		return nil, errors.New("end must be after start")
	}

	b := &domain.Booking{UserID: userID, CourtID: courtID, StartTime: st, EndTime: et, Status: "PENDING"}
	if err := s.repo.CreateWithNoOverlap(ctx, b); err != nil {
		return nil, err
	}

	_ = s.pub.PublishJSON(ctx, "booking.created", map[string]any{
		"booking_id": b.ID, "user_id": b.UserID, "court_id": b.CourtID,
		"start": b.StartTime.Unix(), "end": b.EndTime.Unix(),
	})
	return b, nil
}

func (s *BookingSvc) Confirm(ctx context.Context, id string) (*domain.Booking, error) {
	b, err := s.repo.UpdateStatus(ctx, id, "CONFIRMED")
	if err != nil {
		return nil, err
	}
	_ = s.pub.PublishJSON(ctx, "booking.confirmed", map[string]any{"booking_id": b.ID})
	return b, nil
}

func (s *BookingSvc) Cancel(ctx context.Context, id string) (*domain.Booking, error) {
	b, err := s.repo.UpdateStatus(ctx, id, "CANCELLED")
	if err != nil {
		return nil, err
	}
	_ = s.pub.PublishJSON(ctx, "booking.cancelled", map[string]any{"booking_id": b.ID})
	return b, nil
}

func (s *BookingSvc) Get(ctx context.Context, id string) (*domain.Booking, error) {
	return s.repo.ByID(ctx, id)
}
func (s *BookingSvc) List(ctx context.Context, page, size int32, userID, courtID, dayISO string) ([]domain.Booking, int64, error) {
	return s.repo.List(ctx, page, size, userID, courtID, dayISO)
}
