package grpcx

import (
	"context"
	"fmt"

	userv1 "github.com/you/badminton-booking/proto/user/v1"
	"github.com/you/badminton-booking/services/user-service/internal/domain"
	"github.com/you/badminton-booking/services/user-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	svc *service.UserSvc
}

func NewServer(s *service.UserSvc) *Server {
	return &Server{svc: s}
}

func toPB(u *domain.User) *userv1.User {
	return &userv1.User{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Phone:     u.Phone,
		AvatarUrl: u.AvatarURL,
		Role:      u.Role,
	}
}

// NOTE: สำหรับ GetMe / UpdateUser (me) เราจะพึ่งพา JWT metadata จาก API Gateway ส่ง user_id เข้ามาใน context
type ctxKey string

const CtxUserID ctxKey = "x-user-id"
const CtxUserEmail ctxKey = "x-user-email"
const CtxUserRole ctxKey = "x-user-role"

func (s *Server) GetUser(ctx context.Context, in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	u, err := s.svc.GetByID(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &userv1.GetUserResponse{User: toPB(u)}, nil
}

func first(ss []string) string {
	if len(ss) > 0 {
		return ss[0]
	}
	return ""
}

func (s *Server) GetMe(ctx context.Context, _ *userv1.GetMeRequest) (*userv1.GetMeResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	userID := first(md.Get("x-user-id"))
	email := first(md.Get("x-user-email"))
	role := first(md.Get("x-user-role"))

	fmt.Printf("Data in Get Me %s, %s, %s", userID, email, role)

	// 1) ลองด้วย user_id ก่อน
	if userID != "" {
		if u, err := s.svc.GetByID(ctx, userID); err == nil {
			return &userv1.GetMeResponse{User: toPB(u)}, nil
		}
	}

	// 2) ถัดไปลองจาก email
	if email != "" {
		if u, err := s.svc.GetByEmail(ctx, email); err == nil {
			return &userv1.GetMeResponse{User: toPB(u)}, nil
		}
		// ไม่เจอ → สร้างจาก auth (auto upsert)
		if u, err := s.svc.SyncFromAuth(ctx, email, "", role); err == nil {
			return &userv1.GetMeResponse{User: toPB(u)}, nil
		}
	}

	// 3) ไม่มีข้อมูลพอ
	return nil, status.Error(codes.NotFound, "user not found")
}

func (s *Server) UpdateUser(ctx context.Context, in *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	// ถ้าไม่ส่ง id = ถือเป็น me
	id := in.Id
	if id == "" {
		if v := ctx.Value(CtxUserID); v != nil {
			id = v.(string)
		}
	}
	u, err := s.svc.Update(ctx, id, in.Name, in.Phone, in.AvatarUrl)
	if err != nil {
		return nil, err
	}
	return &userv1.UpdateUserResponse{User: toPB(u)}, nil
}

func (s *Server) ListUsers(ctx context.Context, in *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	users, total, err := s.svc.List(ctx, in.Page, in.PageSize, in.Query, in.Role)
	if err != nil {
		return nil, err
	}
	resp := &userv1.ListUsersResponse{Total: total}
	for i := range users {
		resp.Users = append(resp.Users, toPB(&users[i]))
	}
	return resp, nil
}

func (s *Server) SyncFromAuth(ctx context.Context, in *userv1.SyncFromAuthRequest) (*userv1.SyncFromAuthResponse, error) {
	u, err := s.svc.SyncFromAuth(ctx, in.Email, in.Name, in.Role)
	if err != nil {
		return nil, err
	}
	return &userv1.SyncFromAuthResponse{User: toPB(u)}, nil
}
