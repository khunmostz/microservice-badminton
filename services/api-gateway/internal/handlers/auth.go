package handlers

import (
	"net/http"

	authv1 "github.com/you/badminton-booking/proto/auth/v1"
	userv1 "github.com/you/badminton-booking/proto/user/v1"
	"github.com/you/badminton-booking/services/api-gateway/internal/clients"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	c *clients.Clients
}

func NewAuthHandler(c *clients.Clients) *AuthHandler {
	return &AuthHandler{c: c}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var in struct{ Email, Password, Name, Role string }
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.c.Auth.Register(c, &authv1.RegisterRequest{Email: in.Email, Password: in.Password, Name: in.Name, Role: in.Role})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, _ = h.c.User.SyncFromAuth(c, &userv1.SyncFromAuthRequest{
		Email: res.User.Email,
		Name:  res.User.Name,
		Role:  res.User.Role,
	})
	c.JSON(http.StatusCreated, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var in struct{ Email, Password string }
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.c.Auth.Login(c, &authv1.LoginRequest{Email: in.Email, Password: in.Password})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
