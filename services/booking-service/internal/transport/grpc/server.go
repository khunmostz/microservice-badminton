package grpc

import (
	"context"
	"time"

	bookingv1 "github.com/you/badminton-booking/proto/booking/v1"
	"github.com/you/badminton-booking/services/booking-service/internal/domain"
	"github.com/you/badminton-booking/services/booking-service/internal/service"
)

type Server struct {
	bookingv1.UnimplementedBookingServiceServer
	svc *service.BookingSvc
}

func NewServer(s *service.BookingSvc) *Server {
	return &Server{svc: s}
}

func toPB(b *domain.Booking) *bookingv1.Booking {
	return &bookingv1.Booking{
		Id:       b.ID,
		UserId:   b.UserID,
		CourtId:  b.CourtID,
		StartIso: b.StartTime.UTC().Format(time.RFC3339),
		EndIso:   b.EndTime.UTC().Format(time.RFC3339),
		Status:   statusToEnum(b.Status),
	}
}

func statusToEnum(s string) bookingv1.BookingStatus {
	switch s {
	case "PENDING":
		return bookingv1.BookingStatus_PENDING
	case "CONFIRMED":
		return bookingv1.BookingStatus_CONFIRMED
	case "CANCELLED":
		return bookingv1.BookingStatus_CANCELLED
	default:
		return bookingv1.BookingStatus_BOOKING_STATUS_UNSPECIFIED
	}
}

func (s *Server) CreateBooking(ctx context.Context, in *bookingv1.CreateBookingRequest) (*bookingv1.CreateBookingResponse, error) {
	b, err := s.svc.Create(ctx, in.UserId, in.CourtId, in.StartIso, in.EndIso)
	if err != nil {
		return nil, err
	}
	return &bookingv1.CreateBookingResponse{Booking: toPB(b)}, nil
}

func (s *Server) GetBooking(ctx context.Context, in *bookingv1.GetBookingRequest) (*bookingv1.GetBookingResponse, error) {
	b, err := s.svc.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &bookingv1.GetBookingResponse{Booking: toPB(b)}, nil
}

func (s *Server) ListBooking(ctx context.Context, in *bookingv1.ListBookingRequest) (*bookingv1.ListBookingResponse, error) {
	list, total, err := s.svc.List(ctx, in.Page, in.PageSize, in.UserId, in.CourtId, in.DayIso)
	if err != nil {
		return nil, err
	}
	resp := &bookingv1.ListBookingResponse{Total: total}
	for i := range list {
		resp.Bookings = append(resp.Bookings, toPB(&list[i]))
	}
	return resp, nil
}

func (s *Server) ConfirmBooking(ctx context.Context, in *bookingv1.ConfirmBookingRequest) (*bookingv1.ConfirmBookingResponse, error) {
	b, err := s.svc.Confirm(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &bookingv1.ConfirmBookingResponse{Booking: toPB(b)}, nil
}

func (s *Server) CancelBooking(ctx context.Context, in *bookingv1.CancelBookingRequest) (*bookingv1.CancelBookingResponse, error) {
	b, err := s.svc.Cancel(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &bookingv1.CancelBookingResponse{Booking: toPB(b)}, nil
}
