package clients

import (
	"log"

	authv1 "github.com/you/badminton-booking/proto/auth/v1"
	bookingv1 "github.com/you/badminton-booking/proto/booking/v1"
	courtv1 "github.com/you/badminton-booking/proto/court/v1"
	paymentv1 "github.com/you/badminton-booking/proto/payment/v1"
	userv1 "github.com/you/badminton-booking/proto/user/v1"

	"google.golang.org/grpc"
)

type Clients struct {
	Auth  authv1.AuthServiceClient
	Court courtv1.CourtServiceClient
	Book  bookingv1.BookingServiceClient
	Pay   paymentv1.PaymentServiceClient
	User  userv1.UserServiceClient
}

func New(authAddr, courtAddr, bookingAddr, paymentAddr, userAddr string) *Clients {
	a, err := grpc.Dial(authAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	c, err := grpc.Dial(courtAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	b, err := grpc.Dial(bookingAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	p, err := grpc.Dial(paymentAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	u, err := grpc.Dial(userAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return &Clients{
		Auth:  authv1.NewAuthServiceClient(a),
		Court: courtv1.NewCourtServiceClient(c),
		Book:  bookingv1.NewBookingServiceClient(b),
		Pay:   paymentv1.NewPaymentServiceClient(p),
		User:  userv1.NewUserServiceClient(u),
	}
}
