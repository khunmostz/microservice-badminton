package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"

	"github.com/you/badminton-booking/pkg/db"
	"github.com/you/badminton-booking/pkg/mq"
	bookingv1 "github.com/you/badminton-booking/proto/booking/v1"
	cons "github.com/you/badminton-booking/services/booking-service/internal/consumer"
	"github.com/you/badminton-booking/services/booking-service/internal/repository"
	"github.com/you/badminton-booking/services/booking-service/internal/service"
	tgrpc "github.com/you/badminton-booking/services/booking-service/internal/transport/grpc"
)

type Cfg struct {
	PGBookingDSN    string `envconfig:"PG_BOOKING_DSN" required:"true"`
	BookingGRPCAddr string `envconfig:"BOOKING_GRPC_ADDR" default:":50053"`

	// RabbitMQ for consuming payment events
	RabbitURL       string `envconfig:"RABBIT_URL" required:"true"`
	PaymentExchange string `envconfig:"PAYMENT_EXCHANGE" default:"payment.exchange"`
	PaymentQueue    string `envconfig:"BOOKING_PAYMENT_QUEUE" default:"booking.payment.q"`

	// RabbitMQ for publishing booking events (e.g. booking.confirmed)
	BookingExchange string `envconfig:"BOOKING_EXCHANGE" default:"booking.exchange"`
}

func must[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func main() {
	var cfg Cfg
	must(0, envconfig.Process("", &cfg))

	// DB
	gdb := db.Open(cfg.PGBookingDSN)
	repo := repository.NewBookingRepo(gdb)
	must(0, repo.Migrate())

	// Publisher (สำหรับปล่อย booking.* events)
	bookingPub := must(mq.NewPublisher(cfg.RabbitURL, cfg.BookingExchange))
	defer bookingPub.Close()

	// gRPC server ของ booking-service
	svc := service.NewBookingSvc(repo, bookingPub) // ✅ ตอนนี้ส่งครบ 2 อาร์กิวเมนต์แล้ว
	lis := must(net.Listen("tcp", cfg.BookingGRPCAddr))
	gs := grpc.NewServer()
	bookingv1.RegisterBookingServiceServer(gs, tgrpc.NewServer(svc))

	// Consumer (ฟัง payment.paid)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	paymentCons := must(mq.NewConsumer(cfg.RabbitURL, cfg.PaymentExchange, cfg.PaymentQueue, []string{"payment.paid"}))
	defer paymentCons.Close()
	pc := cons.NewPaymentConsumer(repo, paymentCons)
	must(0, pc.Run(ctx))
	log.Println("[booking] consumer started (payment.paid)")

	// start gRPC
	go func() {
		log.Println("[booking] gRPC listening on", cfg.BookingGRPCAddr)
		log.Fatal(gs.Serve(lis))
	}()

	// graceful shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
	gs.GracefulStop()
	log.Println("[booking] stopped")
}
