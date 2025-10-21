package grpc

import (
	"context"

	"github.com/you/badminton-booking/pkg/auth"
	authv1 "github.com/you/badminton-booking/proto/auth/v1"
	"github.com/you/badminton-booking/services/auth-service/internal/service"
)

type Server struct {
	authv1.UnimplementedAuthServiceServer
	svc *service.AuthSvc
}

func NewServer(s *service.AuthSvc) *Server {
	return &Server{svc: s}
}

func (s *Server) Register(ctx context.Context, in *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	u, err := s.svc.Register(ctx, in.Email, in.Password, in.Name, in.Role)
	if err != nil {
		return nil, err
	}
	return &authv1.RegisterResponse{User: &authv1.User{Id: u.ID, Email: u.Email, Name: u.Name, Role: string(u.Role)}}, nil
}

func (s *Server) Login(ctx context.Context, in *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	u, at, rt, err := s.svc.Login(ctx, in.Email, in.Password)
	if err != nil {
		return nil, err
	}
	return &authv1.LoginResponse{AccessToken: at, RefreshToken: rt, User: &authv1.User{Id: u.ID, Email: u.Email, Name: u.Name, Role: string(u.Role)}}, nil
}

func (s *Server) ValidateToken(ctx context.Context, in *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	// ใน MVP ตรวจโดยตรง ไม่เรียกฐานข้อมูล
	claims, err := auth.ParseValidate(in.Token)
	if err != nil {
		return &authv1.ValidateTokenResponse{Valid: false}, nil
	}
	return &authv1.ValidateTokenResponse{UserId: claims.Sub, Role: string(claims.Role), Valid: true}, nil
}
