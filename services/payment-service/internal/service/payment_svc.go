// services/payment-service/internal/service/payment_svc.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"

	"github.com/you/badminton-booking/pkg/mq"
)

type PaymentSvc struct {
	omc       *omise.Client
	secretKey string // ใช้ยิง REST (Basic Auth) -> skey_xxx
	pub       *mq.Publisher
}

func NewPaymentSvc(omc *omise.Client, secretKey string, pub *mq.Publisher) *PaymentSvc {
	return &PaymentSvc{omc: omc, secretKey: secretKey, pub: pub}
}

// ---------- helpers (publish events) ----------
func (s *PaymentSvc) publishPaid(ctx context.Context, bookingID, chargeID string, amount int64, currency string) {
	if s.pub == nil {
		return
	}
	_ = s.pub.PublishJSON(ctx, "payment.paid", map[string]any{
		"booking_id": bookingID,
		"charge_id":  chargeID,
		"amount":     amount,
		"currency":   currency,
	})
}

func (s *PaymentSvc) publishFailed(ctx context.Context, bookingID, chargeID, code, message string) {
	if s.pub == nil {
		return
	}
	_ = s.pub.PublishJSON(ctx, "payment.failed", map[string]any{
		"booking_id":      bookingID,
		"charge_id":       chargeID,
		"failure_code":    code,
		"failure_message": message,
	})
}

// ---------- Card ----------
type CreateCardChargeInput struct {
	BookingID string
	Amount    int64
	Currency  string
	CardToken string
}

func (s *PaymentSvc) CreateCardCharge(in CreateCardChargeInput) (*omise.Charge, error) {
	if in.Amount <= 0 || in.CardToken == "" || in.Currency == "" {
		return nil, errors.New("invalid params")
	}
	ch := &omise.Charge{}
	req := &operations.CreateCharge{
		Amount:   in.Amount,
		Currency: in.Currency,
		Card:     in.CardToken,
		Metadata: map[string]any{"booking_id": in.BookingID},
	}

	if err := s.omc.Do(ch, req); err != nil {
		// publish failed without charge_id (ยังสร้าง charge ไม่สำเร็จ)
		s.publishFailed(context.Background(), in.BookingID, "", "create_charge_error", err.Error())
		return nil, err
	}

	fmt.Printf("Payment status %s\n", string(ch.Status))

	// สถานะจาก Omise: pending / successful / failed / awaiting_authorize
	switch string(ch.Status) {
	case "successful":
		s.publishPaid(context.Background(), in.BookingID, ch.ID, ch.Amount, ch.Currency)
	case "failed":
		var fc, fm string
		if ch.FailureCode != nil {
			fc = *ch.FailureCode
		}
		if ch.FailureMessage != nil {
			fm = *ch.FailureMessage
		}
		s.publishFailed(context.Background(), in.BookingID, ch.ID, fc, fm)
		// pending/awaiting_authorize ไม่ยิง paid/failed ณ จุดนี้ (รอ webhook ยืนยันผลสุดท้าย)
	}

	return ch, nil
}

// ---------- Source (ใช้ source_id ตรง ๆ) ----------
type CreateChargeWithSourceIDInput struct {
	BookingID string
	Amount    int64
	Currency  string
	SourceID  string
}

func (s *PaymentSvc) CreateChargeWithSourceID(in CreateChargeWithSourceIDInput) (*omise.Charge, error) {
	if in.Amount <= 0 || in.Currency == "" || in.SourceID == "" {
		return nil, errors.New("invalid params")
	}
	ch := &omise.Charge{}
	req := &operations.CreateCharge{
		Amount:   in.Amount,
		Currency: in.Currency,
		Source:   in.SourceID,
		Metadata: map[string]any{"booking_id": in.BookingID},
	}

	if err := s.omc.Do(ch, req); err != nil {
		s.publishFailed(context.Background(), in.BookingID, "", "create_charge_error", err.Error())
		return nil, err
	}

	switch string(ch.Status) {
	case "successful":
		s.publishPaid(context.Background(), in.BookingID, ch.ID, ch.Amount, ch.Currency)
	case "failed":
		var fc, fm string
		if ch.FailureCode != nil {
			fc = *ch.FailureCode
		}
		if ch.FailureMessage != nil {
			fm = *ch.FailureMessage
		}
		s.publishFailed(context.Background(), in.BookingID, ch.ID, fc, fm)
	}

	return ch, nil
}

// ---------- Source helper (ถ้า client ไม่ส่ง source_id) ----------

// CreateSourceOrUseExisting: ถ้ามี source_id แล้ว => ใช้เลย; ถ้าไม่มีแต่ให้ source_type (+return_uri ถ้าจำเป็น)
// จะสร้าง source ผ่าน SDK (promptpay) หรือ REST fallback (ช่องทางที่ต้อง return_uri)
func (s *PaymentSvc) CreateSourceOrUseExisting(
	ctx context.Context,
	amount int64, currency, sourceID, sourceType, returnURI string,
) (*omise.Source, error) {

	if sourceID != "" {
		src := &omise.Source{}
		// NOTE: ถ้า SDK ของคุณใช้ field ชื่อ ID ให้เปลี่ยนเป็น &operations.RetrieveSource{ID: sourceID}
		if err := s.omc.Do(src, &operations.RetrieveSource{SourceID: sourceID}); err != nil {
			return nil, err
		}
		return src, nil
	}

	if sourceType == "" {
		return nil, errors.New("either source_id or source_type is required")
	}
	if amount <= 0 || currency == "" {
		return nil, errors.New("invalid params")
	}

	// promptpay ไม่ต้อง return_uri -> ใช้ SDK ได้เลย
	if strings.EqualFold(sourceType, "promptpay") {
		src := &omise.Source{}
		req := &operations.CreateSource{
			Type:     sourceType,
			Amount:   amount,
			Currency: currency,
		}
		if err := s.omc.Do(src, req); err != nil {
			return nil, err
		}
		return src, nil
	}

	// ช่องทางที่ต้อง redirect (เช่น mobile_banking_kbank / internet_banking_*)
	return s.createSourceViaREST(ctx, sourceType, amount, currency, returnURI)
}

// createSourceViaREST: ยิง REST ไป Omise เพื่อสร้าง source ที่ต้องการ return_uri
func (s *PaymentSvc) createSourceViaREST(
	ctx context.Context,
	sourceType string,
	amount int64,
	currency string,
	returnURI string,
) (*omise.Source, error) {
	if s.secretKey == "" {
		return nil, errors.New("missing Omise secret key for REST call")
	}

	form := url.Values{}
	form.Set("type", sourceType)
	form.Set("amount", strconv.FormatInt(amount, 10))
	form.Set("currency", currency)
	// บางช่องทางจำเป็นต้องมี return_uri (เช่น internet_banking_*, mobile_banking_*)
	if returnURI != "" {
		form.Set("return_uri", returnURI)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.omise.co/sources", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Basic Auth: username = skey_xxx, password = "" (ว่าง)
	req.SetBasicAuth(s.secretKey, "")

	httpClient := &http.Client{Timeout: 15 * time.Second}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("omise create source failed: %s (%d)", string(body), res.StatusCode)
	}

	// ถอดรหัสเป็น omise.Source
	var src omise.Source
	if err := json.Unmarshal(body, &src); err != nil {
		return nil, fmt.Errorf("parse source json failed: %w", err)
	}
	return &src, nil
}

// ---------- Retrieve ----------
func (s *PaymentSvc) GetCharge(id string) (*omise.Charge, error) {
	ch := &omise.Charge{}
	// NOTE: ถ้า SDK ของคุณใช้ field ชื่อ ID ให้เปลี่ยนเป็น &operations.RetrieveCharge{ID: id}
	if err := s.omc.Do(ch, &operations.RetrieveCharge{ChargeID: id}); err != nil {
		return nil, err
	}
	return ch, nil
}
