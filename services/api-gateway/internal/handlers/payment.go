package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	paymentv1 "github.com/you/badminton-booking/proto/payment/v1"
	"github.com/you/badminton-booking/services/api-gateway/internal/clients"
)

type PaymentHandler struct {
	c *clients.Clients
}

func NewPaymentHandler(c *clients.Clients) *PaymentHandler { return &PaymentHandler{c: c} }

// ---------- Card ----------
type createCardChargeBody struct {
	BookingID string `json:"booking_id" binding:"required"`
	Amount    int64  `json:"amount" binding:"required"`
	Currency  string `json:"currency" binding:"required"` // "THB"
	CardToken string `json:"card_token" binding:"required"`
}

func (h *PaymentHandler) CreateCardCharge(c *gin.Context) {
	var body createCardChargeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.c.Pay.CreateCardCharge(c, &paymentv1.CreateCardChargeRequest{
		BookingId: body.BookingID,
		Amount:    body.Amount,
		Currency:  body.Currency,
		CardToken: body.CardToken,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ---------- Source ----------
// รองรับ 2 เคส:
//  1. client ส่ง source_id มาเลย
//  2. client ไม่ส่ง source_id แต่ส่ง source_type (+return_uri ถ้าจำเป็น) -> ให้ payment-service สร้าง source เอง
type createSourceChargeBody struct {
	BookingID string `json:"booking_id" binding:"required"`
	Amount    int64  `json:"amount" binding:"required"`
	Currency  string `json:"currency" binding:"required"` // "THB"

	// ทางเลือกที่ 1: ส่งมาเลย
	SourceID string `json:"source_id"`

	// ทางเลือกที่ 2: ให้ server สร้าง source ให้ (ต้องมีใน proto ของคุณด้วย)
	SourceType string `json:"source_type"`
	ReturnURI  string `json:"return_uri"`
}

func (h *PaymentHandler) CreateSourceCharge(c *gin.Context) {
	var body createSourceChargeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := &paymentv1.CreateSourceChargeRequest{
		BookingId:  body.BookingID,
		Amount:     body.Amount,
		Currency:   body.Currency,
		SourceId:   body.SourceID,   // อาจว่างได้
		ReturnUri:  body.ReturnURI,  // server อาจใช้เมื่อต้องสร้าง source แบบ redirect
		SourceType: body.SourceType, // <-- ต้องมี field นี้ใน proto (ถ้าใช้วิธีที่ 2)
	}

	// หมายเหตุ:
	// - ถ้า proto ของคุณ 'ยังไม่มี' SourceType: ให้ลบบรรทัดตั้งค่า SourceType ออก และบังคับว่าต้องมี SourceId
	// - ถ้ามีแล้ว: ฝั่ง payment-service จะตรวจเองว่า ถ้า SourceId ว่างแต่มี SourceType ก็ไปเรียก CreateSourceOrUseExisting

	resp, err := h.c.Pay.CreateSourceCharge(c, req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PaymentHandler) GetCharge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing charge id"})
		return
	}
	resp, err := h.c.Pay.GetCharge(c, &paymentv1.GetChargeRequest{ChargeId: id})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
