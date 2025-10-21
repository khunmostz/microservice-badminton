package main

import (
	"log"
	"net"

	"github.com/you/badminton-booking/pkg/config"
	"github.com/you/badminton-booking/pkg/db"
	authv1 "github.com/you/badminton-booking/proto/auth/v1"
	"github.com/you/badminton-booking/services/auth-service/internal/repository"
	"github.com/you/badminton-booking/services/auth-service/internal/service"
	tgrpc "github.com/you/badminton-booking/services/auth-service/internal/transport/grpc"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	gdb := db.Open(cfg.PGAuthDSN)

	repo := repository.NewUserRepo(gdb)
	if err := repo.Migrate(); err != nil {
		log.Fatal(err)
	}
	svc := service.NewAuthSvc(repo)

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, tgrpc.NewServer(svc))

	lis, err := net.Listen("tcp", cfg.AuthGRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[auth] gRPC on %s", cfg.AuthGRPCAddr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
