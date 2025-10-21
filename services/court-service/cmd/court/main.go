package main

import (
	"fmt"
	"log"
	"net"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/you/badminton-booking/pkg/db"
	"github.com/you/badminton-booking/services/court-service/internal/repository"
	"github.com/you/badminton-booking/services/court-service/internal/service"
	"google.golang.org/grpc"

	courtv1 "github.com/you/badminton-booking/proto/court/v1"
	tgrpc "github.com/you/badminton-booking/services/court-service/internal/transport/grpc"
)

type CourtCfg struct {
	PGCourtDSN    string `envconfig:"PG_COURT_DSN" required:"true"`
	CourtGRPCAddr string `envconfig:"COURT_GRPC_ADDR" default:":50052"`
}

func loadCfg() (CourtCfg, error) {
	var c CourtCfg
	err := envconfig.Process("", &c)
	return c, err
}

func main() {
	_ = godotenv.Load(".env")
	cfg, err := loadCfg()
	if err != nil {
		logErr := fmt.Sprintf("error is %s", err)
		log.Fatal(logErr)
	}
	gdb := db.Open(cfg.PGCourtDSN)

	repo := repository.NewCourtRepo(gdb)
	if err := repo.Migrate(); err != nil {
		log.Fatal(err)
	}

	svc := service.NewCourtSvc(repo)
	lis, _ := net.Listen("tcp", cfg.CourtGRPCAddr) // หรือ cfg.CourtGRPCAddr
	s := grpc.NewServer()
	courtv1.RegisterCourtServiceServer(s, tgrpc.NewServer(svc))
	log.Printf("court-service listening on %s\n", cfg.CourtGRPCAddr)
	log.Fatal(s.Serve(lis))
}
