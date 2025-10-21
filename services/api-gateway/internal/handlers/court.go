package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	courtv1 "github.com/you/badminton-booking/proto/court/v1"
	"github.com/you/badminton-booking/services/api-gateway/internal/clients"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CourtHandler struct {
	c *clients.Clients
}

func NewCourtHandler(c *clients.Clients) *CourtHandler {
	return &CourtHandler{c: c}
}

func (h *CourtHandler) Create(c *gin.Context) {
	var in struct {
		Venue        string `json:"venue" binding:"required"`
		CourtNo      int32  `json:"court_no" binding:"required"`
		PricePerHour int64  `json:"price_per_hour" binding:"required"`
		OpenFrom     string `json:"open_from" binding:"required"`
		OpenTo       string `json:"open_to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.c.Court.CreateCourt(c, &courtv1.CreateCourtRequest{
		Venue:        in.Venue,
		CourtNo:      in.CourtNo,
		PricePerHour: in.PricePerHour,
		OpenFrom:     in.OpenFrom,
		OpenTo:       in.OpenTo,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *CourtHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	venueQuery := c.Query("q")

	req := &courtv1.ListCourtsRequest{
		Page:       int32(page - 1),
		PageSize:   int32(size),
		VenueQuery: venueQuery,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	res, err := h.c.Court.ListCourts(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(grpcCodeToHTTP(st.Code()), gin.H{
				"error":   st.Message(),
				"details": st.Details(),
			})
		} else {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, res)
}

func grpcCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.DeadlineExceeded, codes.ResourceExhausted, codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.Internal, codes.DataLoss, codes.Unknown:
		return http.StatusBadGateway
	default:
		return http.StatusBadRequest
	}
}
