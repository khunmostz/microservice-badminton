package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	bookingv1 "github.com/you/badminton-booking/proto/booking/v1"
	"github.com/you/badminton-booking/services/api-gateway/internal/clients"
)

type BookingHandler struct {
	c *clients.Clients
}

func NewBookingHandler(c *clients.Clients) *BookingHandler {
	return &BookingHandler{c: c}
}

// POST /v1/bookings
func (h *BookingHandler) Create(c *gin.Context) {
	var in struct {
		CourtID  string `json:"court_id" binding:"required"`
		StartISO string `json:"start_iso" binding:"required"` // RFC3339
		EndISO   string `json:"end_iso"   binding:"required"` // RFC3339
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sub, _ := c.Get("sub") // set by JWTAuth middleware
	userID, _ := sub.(string)
	res, err := h.c.Book.CreateBooking(c, &bookingv1.CreateBookingRequest{
		UserId:   userID,
		CourtId:  in.CourtID,
		StartIso: in.StartISO,
		EndIso:   in.EndISO,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// POST /v1/bookings/:id/confirm (OWNER/ADMIN)
func (h *BookingHandler) Confirm(c *gin.Context) {
	id := c.Param("id")
	res, err := h.c.Book.ConfirmBooking(c, &bookingv1.ConfirmBookingRequest{Id: id})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// POST /v1/bookings/:id/cancel (owner of booking or ADMIN — MVP allow any authenticated)
func (h *BookingHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	res, err := h.c.Book.CancelBooking(c, &bookingv1.CancelBookingRequest{Id: id})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /v1/bookings/:id
func (h *BookingHandler) Get(c *gin.Context) {
	id := c.Param("id")
	res, err := h.c.Book.GetBooking(c, &bookingv1.GetBookingRequest{Id: id})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /v1/bookings?page=1&page_size=20&user_id=...&court_id=...&day=RFC3339
func (h *BookingHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}

	req := &bookingv1.ListBookingRequest{
		Page:     int32(page - 1),
		PageSize: int32(size),
		UserId:   c.Query("user_id"),
		CourtId:  c.Query("court_id"),
		DayIso:   c.Query("day"),
	}
	res, err := h.c.Book.ListBooking(c, req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
