package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/you/badminton-booking/services/court-service/internal/domain"
	"gorm.io/gorm"
)

type CourtRepo struct {
	db *gorm.DB
}

func NewCourtRepo(db *gorm.DB) *CourtRepo {
	return &CourtRepo{db: db}
}
func (r *CourtRepo) Migrate() error {
	return r.db.AutoMigrate(&domain.Court{})
}

func (r *CourtRepo) Create(ctx context.Context, c *domain.Court) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return r.db.WithContext(ctx).Create(c).Error
}
func (r *CourtRepo) ByID(ctx context.Context, id string) (*domain.Court, error) {
	var c domain.Court
	if err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
func (r *CourtRepo) List(ctx context.Context, page, size int32, venue string) ([]domain.Court, error) {
	if size <= 0 {
		size = 20
	}
	if page < 0 {
		page = 0
	}
	qb := r.db.WithContext(ctx).Model(&domain.Court{})
	if venue != "" {
		qb = qb.Where("venue ILIKE ?", "%"+venue+"%")
	}
	var out []domain.Court
	if err := qb.Limit(int(size)).Offset(int(page * size)).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
func (r *CourtRepo) Update(ctx context.Context, c *domain.Court) error {
	return r.db.WithContext(ctx).Model(&domain.Court{}).Where("id = ?", c.ID).Updates(c).Error
}
func (r *CourtRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Court{}, "id = ?", id).Error
}
