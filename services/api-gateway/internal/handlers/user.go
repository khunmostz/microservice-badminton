package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	userv1 "github.com/you/badminton-booking/proto/user/v1"
	"github.com/you/badminton-booking/services/api-gateway/internal/clients"

	"google.golang.org/grpc/metadata"
)

type UserHandler struct {
	c *clients.Clients
}

func NewUserHandler(c *clients.Clients) *UserHandler {
	return &UserHandler{c: c}
}

func injectUserMD(c *gin.Context) (ctxWithMD gin.Context, mdCtxForGRPC context.Context) {
	md := metadata.New(nil)
	if v, ok := c.Get("sub"); ok {
		md.Append("x-user-id", v.(string))
	}
	if v, ok := c.Get("email"); ok && v != "" {
		md.Append("x-user-email", v.(string))
	}
	if v, ok := c.Get("role"); ok && v != "" {
		md.Append("x-user-role", v.(string))
	}
	return *c, metadata.NewOutgoingContext(c.Request.Context(), md)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	_, ctx := injectUserMD(c)
	res, err := h.c.User.GetMe(ctx, &userv1.GetMeRequest{})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	var in struct {
		Name      string `json:"name"`
		Phone     string `json:"phone"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, ctx := injectUserMD(c)
	res, err := h.c.User.UpdateUser(ctx, &userv1.UpdateUserRequest{
		Name: in.Name, Phone: in.Phone, AvatarUrl: in.AvatarURL,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.c.User.GetUser(c, &userv1.GetUserRequest{Id: id})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	req := &userv1.ListUsersRequest{
		Page:     int32(page - 1),
		PageSize: int32(size),
		Query:    c.Query("q"),
		Role:     c.Query("role"),
	}
	res, err := h.c.User.ListUsers(c, req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
