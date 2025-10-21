package main

import (
	"log"
	"net"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"

	userv1 "github.com/you/badminton-booking/proto/user/v1"
	"github.com/you/badminton-booking/services/user-service/internal/repository"
	"github.com/you/badminton-booking/services/user-service/internal/service"
	tgrpc "github.com/you/badminton-booking/services/user-service/internal/transport/grpc"

	"github.com/you/badminton-booking/pkg/db"
)

type Cfg struct {
	UserGRPCAddr string `envconfig:"USER_GRPC_ADDR" default:":50051"`
	PGUserDSN    string `envconfig:"PG_USER_DSN" required:"true"`
}

func main() {
	var cfg Cfg
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	gdb := db.Open(cfg.PGUserDSN)
	repo := repository.NewUserRepo(gdb)
	if err := repo.Migrate(); err != nil {
		log.Fatal(err)
	}

	svc := service.NewUserSvc(repo)

	lis, err := net.Listen("tcp", cfg.UserGRPCAddr)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	userv1.RegisterUserServiceServer(s, tgrpc.NewServer(svc))

	log.Println("user-service listening on", cfg.UserGRPCAddr)
	log.Fatal(s.Serve(lis))
}
