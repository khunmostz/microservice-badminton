package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/omise/omise-go"
	"google.golang.org/grpc"

	"github.com/you/badminton-booking/pkg/mq"
	paymentv1 "github.com/you/badminton-booking/proto/payment/v1"

	httpx "github.com/you/badminton-booking/services/payment-service/internal/http"
	omisecli "github.com/you/badminton-booking/services/payment-service/internal/omise"
	paysvc "github.com/you/badminton-booking/services/payment-service/internal/service"
	tgrpc "github.com/you/badminton-booking/services/payment-service/internal/transport/grpc"
)

type Cfg struct {
	PaymentGRPCAddr string `envconfig:"PAYMENT_GRPC_ADDR" default:":50054"`
	WebhookHTTPAddr string `envconfig:"PAYMENT_WEBHOOK_HTTP_ADDR" default:":8081"`
	OmisePub        string `envconfig:"OMISE_PUBLIC_KEY" required:"true"`
	OmiseSec        string `envconfig:"OMISE_SECRET_KEY" required:"true"`
	OmiseVer        string `envconfig:"OMISE_API_VERSION" default:""`
	RabbitURL       string `envconfig:"RABBIT_URL" required:"true"`
	PaymentExchange string `envconfig:"PAYMENT_EXCHANGE" default:"payment.exchange"`
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

	// Omise client
	var omc *omise.Client = must(omisecli.NewOmiseClient(cfg.OmisePub, cfg.OmiseSec, cfg.OmiseVer))

	// MQ publisher
	pub := must(mq.NewPublisher(cfg.RabbitURL, cfg.PaymentExchange))
	defer pub.Close()

	// Start HTTP webhook (publish events)
	mux := http.NewServeMux()
	mux.HandleFunc("/webhooks/omise", httpx.NewWebhookServer(omc, pub, cfg.PaymentExchange).Handler)
	go func() {
		log.Println("[payment] webhook http listening on", cfg.WebhookHTTPAddr)
		log.Fatal(http.ListenAndServe(cfg.WebhookHTTPAddr, mux))
	}()

	// gRPC server (สำหรับสร้าง charge ผ่าน gateway ถ้าคุณมี proto)
	svc := paysvc.NewPaymentSvc(omc, cfg.OmiseSec, pub)
	lis := must(net.Listen("tcp", cfg.PaymentGRPCAddr))
	gs := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(gs, tgrpc.NewServer(svc))
	log.Println("[payment] gRPC listening on", cfg.PaymentGRPCAddr)

	// graceful
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		gs.GracefulStop()
	}()

	log.Fatal(gs.Serve(lis))
}
