package service

import (
	"context"
	"errors"
	"strings"

	"github.com/you/badminton-booking/services/user-service/internal/domain"
	"github.com/you/badminton-booking/services/user-service/internal/repository"
)

type UserSvc struct{ repo *repository.UserRepo }

func NewUserSvc(r *repository.UserRepo) *UserSvc { return &UserSvc{repo: r} }

// SyncFromAuth: เรียกตอน register/login สำเร็จ เพื่อ upsert โปรไฟล์พื้นฐาน
func (s *UserSvc) SyncFromAuth(ctx context.Context, email, name, role string) (*domain.User, error) {
	u := &domain.User{Email: strings.ToLower(email), Name: name, Role: strings.ToUpper(role)}
	if err := s.repo.UpsertByEmail(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserSvc) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.ByID(ctx, id)
}

func (s *UserSvc) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.ByEmail(ctx, strings.ToLower(email))
}

func (s *UserSvc) Update(ctx context.Context, id, name, phone, avatar string) (*domain.User, error) {
	if id == "" {
		return nil, errors.New("missing id")
	}
	fields := map[string]any{}
	if name != "" {
		fields["name"] = name
	}
	if phone != "" {
		fields["phone"] = phone
	}
	if avatar != "" {
		fields["avatar_url"] = avatar
	}
	return s.repo.UpdateFields(ctx, id, fields)
}

func (s *UserSvc) List(ctx context.Context, page, size int32, query, role string) ([]domain.User, int64, error) {
	return s.repo.List(ctx, page, size, query, role)
}
