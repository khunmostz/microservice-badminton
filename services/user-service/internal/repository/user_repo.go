package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/you/badminton-booking/services/user-service/internal/domain"
)

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}
func (r *UserRepo) Migrate() error {
	return r.db.AutoMigrate(&domain.User{})
}

func (r *UserRepo) UpsertByEmail(ctx context.Context, u *domain.User) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	// upsert by email
	return r.db.WithContext(ctx).
		Where("email = ?", u.Email).
		Assign(map[string]any{
			"name":       u.Name,
			"phone":      u.Phone,
			"avatar_url": u.AvatarURL,
			"role":       u.Role,
		}).FirstOrCreate(u).Error
}

func (r *UserRepo) ByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpdateFields(ctx context.Context, id string, fields map[string]any) (*domain.User, error) {
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Updates(fields).Error; err != nil {
		return nil, err
	}
	return r.ByID(ctx, id)
}

func (r *UserRepo) List(ctx context.Context, page, size int32, query, role string) ([]domain.User, int64, error) {
	if size <= 0 {
		size = 20
	}
	if page < 0 {
		page = 0
	}
	qb := r.db.WithContext(ctx).Model(&domain.User{})
	if role != "" {
		qb = qb.Where("role = ?", strings.ToUpper(role))
	}
	if q := strings.TrimSpace(query); q != "" {
		qb = qb.Where("(email ILIKE ? OR name ILIKE ?)", "%"+q+"%", "%"+q+"%")
	}
	var total int64
	if err := qb.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []domain.User
	if err := qb.Order("created_at DESC").Limit(int(size)).Offset(int(page * size)).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
