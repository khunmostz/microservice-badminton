package grpc

import (
	"context"

	paymentv1 "github.com/you/badminton-booking/proto/payment/v1"
	"github.com/you/badminton-booking/services/payment-service/internal/service"
)

type Server struct {
	paymentv1.UnimplementedPaymentServiceServer
	svc *service.PaymentSvc
}

func NewServer(s *service.PaymentSvc) *Server { return &Server{svc: s} }

// ---------- Card ----------
func (s *Server) CreateCardCharge(ctx context.Context, in *paymentv1.CreateCardChargeRequest) (*paymentv1.CreateCardChargeResponse, error) {
	ch, err := s.svc.CreateCardCharge(service.CreateCardChargeInput{
		BookingID: in.BookingId,
		Amount:    in.Amount,
		Currency:  in.Currency,
		CardToken: in.CardToken,
	})
	if err != nil {
		return nil, err
	}
	resp := &paymentv1.CreateCardChargeResponse{
		ChargeId:     ch.ID,
		Status:       string(ch.Status), // omise.ChargeStatus -> string
		AuthorizeUri: ch.AuthorizeURI,   // ถ้ามี 3DS/redirect จะมีค่านี้
	}
	return resp, nil
}

// ---------- Source (client ส่ง source_id + return_uri มา; return_uri ไม่ได้ใช้ตอน charge) ----------
func (s *Server) CreateSourceCharge(ctx context.Context, in *paymentv1.CreateSourceChargeRequest) (*paymentv1.CreateSourceChargeResponse, error) {
	// 1) ได้ source (จาก id หรือสร้างใหม่จาก type)
	src, err := s.svc.CreateSourceOrUseExisting(ctx, in.Amount, in.Currency, in.SourceId, in.SourceType, in.ReturnUri)
	if err != nil {
		return nil, err
	}

	// 2) Charge ด้วย src.ID
	ch, err := s.svc.CreateChargeWithSourceID(service.CreateChargeWithSourceIDInput{
		BookingID: in.BookingId,
		Amount:    in.Amount,
		Currency:  in.Currency,
		SourceID:  src.ID,
	})
	if err != nil {
		return nil, err
	}

	return &paymentv1.CreateSourceChargeResponse{
		ChargeId:     ch.ID,
		Status:       string(ch.Status),
		AuthorizeUri: ch.AuthorizeURI, // ถ้าต้อง redirect จะมีค่านี้
	}, nil
}

// ---------- Get ----------
func (s *Server) GetCharge(ctx context.Context, in *paymentv1.GetChargeRequest) (*paymentv1.GetChargeResponse, error) {
	ch, err := s.svc.GetCharge(in.ChargeId)
	if err != nil {
		return nil, err
	}
	var fc, fm string
	if ch.FailureCode != nil {
		fc = *ch.FailureCode
	}
	if ch.FailureMessage != nil {
		fm = *ch.FailureMessage
	}
	return &paymentv1.GetChargeResponse{
		ChargeId:       ch.ID,
		Status:         string(ch.Status),
		FailureCode:    fc,
		FailureMessage: fm,
	}, nil
}
