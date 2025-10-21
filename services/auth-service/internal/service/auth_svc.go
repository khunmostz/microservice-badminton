package service

import (
	"context"
	"time"

	"github.com/you/badminton-booking/pkg/auth"
	"github.com/you/badminton-booking/services/auth-service/internal/domain"
	"github.com/you/badminton-booking/services/auth-service/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthSvc struct{ repo *repository.UserRepo }

func NewAuthSvc(r *repository.UserRepo) *AuthSvc { return &AuthSvc{repo: r} }

func (s *AuthSvc) Register(ctx context.Context, email, password, name, role string) (*domain.User, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := &domain.User{Email: email, PasswordHash: string(hash), Name: name, Role: domain.Role(role)}
	return u, s.repo.Create(ctx, u)
}

func (s *AuthSvc) Login(ctx context.Context, email, password string) (*domain.User, string, string, error) {
	u, err := s.repo.ByEmail(ctx, email)
	if err != nil {
		return nil, "", "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, "", "", err
	}
	access, _ := auth.CreateAccessToken(u.ID, string(u.Role), u.Email, time.Duration(60)*time.Minute)
	refresh, _ := auth.CreateAccessToken(u.ID, string(u.Role), u.Email, time.Duration(720)*time.Hour)
	return u, access, refresh, nil
}
