package config

import (
	"github.com/kelseyhightower/envconfig"
)

type App struct {
	// DB
	PGAuthDSN string `envconfig:"PG_AUTH_DSN" required:"true"`
	// JWT
	JWTSecret       string `envconfig:"JWT_SECRET" required:"true"`
	JWTExpireMin    int    `envconfig:"JWT_EXPIRE_MIN" default:"60"`
	RefreshExpireHr int    `envconfig:"REFRESH_EXPIRE_HR" default:"720"`
	// Network
	AuthGRPCAddr    string `envconfig:"AUTH_GRPC_ADDR" default:":50051"`
	CourtGRPCAddr   string `envconfig:"COURT_GRPC_ADDR" default:":50052"`
	BookingGRPCAddr string `envconfig:"BOOKING_GRPC_ADDR" default:":50053"`
	PaymentGRPCAddr string `envconfig:"PAYMENT_GRPC_ADDR" default:":50054"`
	UserGRPCAddr    string `envconfig:"USER_GRPC_ADDR" default:":50055"`

	GatewayHTTPAddr string `envconfig:"GATEWAY_HTTP_ADDR" default:":8080"`
}

func Load() (App, error) {
	var c App
	err := envconfig.Process("", &c)
	return c, err
}
