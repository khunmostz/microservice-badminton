package grpc

import (
	"context"

	courtv1 "github.com/you/badminton-booking/proto/court/v1"
	"github.com/you/badminton-booking/services/court-service/internal/domain"
	"github.com/you/badminton-booking/services/court-service/internal/service"
)

type Server struct {
	courtv1.UnimplementedCourtServiceServer
	svc *service.CourtSvc
}

func NewServer(s *service.CourtSvc) *Server {
	return &Server{svc: s}
}

func (s *Server) CreateCourt(ctx context.Context, in *courtv1.CreateCourtRequest) (*courtv1.CreateCourtResponse, error) {
	d := domain.Court{
		Venue:        in.Venue,
		CourtNo:      in.CourtNo,
		PricePerHour: in.PricePerHour,
		OpenFrom:     in.OpenFrom,
		OpenTo:       in.OpenTo,
	}
	out, err := s.svc.Create(ctx, d)
	if err != nil {
		return nil, err
	}
	return &courtv1.CreateCourtResponse{Court: toPB(out)}, nil
}

func (s *Server) GetCourt(ctx context.Context, in *courtv1.GetCourtRequest) (*courtv1.GetCourtResponse, error) {
	c, err := s.svc.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &courtv1.GetCourtResponse{Court: toPB(c)}, nil
}
func (s *Server) ListCourts(ctx context.Context, in *courtv1.ListCourtsRequest) (*courtv1.ListCourtsResponse, error) {
	list, err := s.svc.List(ctx, in.Page, in.PageSize, in.VenueQuery)
	if err != nil {
		return nil, err
	}
	resp := &courtv1.ListCourtsResponse{}
	for i := range list {
		resp.Courts = append(resp.Courts, toPB(&list[i]))
	}
	return resp, nil
}
func (s *Server) UpdateCourt(ctx context.Context, in *courtv1.UpdateCourtRequest) (*courtv1.UpdateCourtResponse, error) {
	d := domain.Court{ID: in.Id, Venue: in.Venue, CourtNo: in.CourtNo, PricePerHour: in.PricePerHour, OpenFrom: in.OpenFrom, OpenTo: in.OpenTo}
	out, err := s.svc.Update(ctx, d)
	if err != nil {
		return nil, err
	}
	return &courtv1.UpdateCourtResponse{Court: toPB(out)}, nil
}
func (s *Server) DeleteCourt(ctx context.Context, in *courtv1.DeleteCourtRequest) (*courtv1.DeleteCourtResponse, error) {
	return &courtv1.DeleteCourtResponse{}, s.svc.Delete(ctx, in.Id)
}

func toPB(c *domain.Court) *courtv1.Court {
	return &courtv1.Court{Id: c.ID, Venue: c.Venue, CourtNo: c.CourtNo, PricePerHour: c.PricePerHour, OpenFrom: c.OpenFrom, OpenTo: c.OpenTo, OwnerId: c.OwnerID}
}
