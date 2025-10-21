package service

import (
	"context"

	"github.com/you/badminton-booking/services/court-service/internal/domain"
	"github.com/you/badminton-booking/services/court-service/internal/repository"
)

type CourtSvc struct {
	repo *repository.CourtRepo
}

func NewCourtSvc(r *repository.CourtRepo) *CourtSvc {
	return &CourtSvc{repo: r}
}

func (s *CourtSvc) Create(ctx context.Context, in domain.Court) (*domain.Court, error) {
	if err := s.repo.Create(ctx, &in); err != nil {
		return nil, err
	}
	return &in, nil
}
func (s *CourtSvc) Get(ctx context.Context, id string) (*domain.Court, error) {
	return s.repo.ByID(ctx, id)
}
func (s *CourtSvc) List(ctx context.Context, page, size int32, venue string) ([]domain.Court, error) {
	return s.repo.List(ctx, page, size, venue)
}
func (s *CourtSvc) Update(ctx context.Context, in domain.Court) (*domain.Court, error) {
	if err := s.repo.Update(ctx, &in); err != nil {
		return nil, err
	}
	return &in, nil
}
func (s *CourtSvc) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
