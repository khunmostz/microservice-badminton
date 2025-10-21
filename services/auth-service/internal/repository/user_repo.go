package repository

import (
	"context"

	"github.com/you/badminton-booking/services/auth-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Migrate() error {
	return r.db.AutoMigrate(&domain.User{})
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	return r.db.WithContext(ctx).Create(u).Error
}
func (r *UserRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *UserRepo) ByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
