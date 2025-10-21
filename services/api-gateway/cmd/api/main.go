package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/you/badminton-booking/pkg/config"
	"github.com/you/badminton-booking/services/api-gateway/internal/clients"
	"github.com/you/badminton-booking/services/api-gateway/internal/handlers"
	"github.com/you/badminton-booking/services/api-gateway/internal/middlewares"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	c := clients.New(cfg.AuthGRPCAddr, cfg.CourtGRPCAddr, cfg.BookingGRPCAddr, cfg.PaymentGRPCAddr, cfg.UserGRPCAddr)
	r := gin.Default()

	r.GET("/payments/return", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, `
		  <html><body>
			<h3>Payment flow completed (client-side)</h3>
			<p>charge_id: %s</p>
			<p>Note: Final status will be confirmed via Webhook or polling.</p>
		  </body></html>
		`, c.Query("charge_id"))
	})

	a := handlers.NewAuthHandler(c)
	v1 := r.Group("/v1")
	{
		v1.POST("/auth/register", a.Register)
		v1.POST("/auth/login", a.Login)

		uh := handlers.NewUserHandler(c)
		{
			me := v1.Group("/users/me")
			me.Use(middlewares.JWTAuth())
			me.GET("", uh.GetMe)
			me.PUT("", uh.UpdateMe)

			admin := v1.Group("/users")
			admin.Use(middlewares.JWTAuth(), middlewares.RequireRole("ADMIN"))
			admin.GET("", uh.List)
			admin.GET("/:id", uh.GetByID)
		}

		ch := handlers.NewCourtHandler(c)
		v1.GET("/courts", ch.List)
		v1.POST(
			"/courts",
			middlewares.JWTAuth(),
			middlewares.RequireRole("OWNER", "ADMIN"),
			ch.Create,
		)

		bh := handlers.NewBookingHandler(c)
		secured := v1.Group("")
		secured.Use(middlewares.JWTAuth())
		{
			secured.POST("/bookings", bh.Create)
			secured.GET("/bookings", bh.List)
			secured.GET("/bookings/:id", bh.Get)

			owner := secured.Group("")
			owner.Use(middlewares.RequireRole("OWNER", "ADMIN"))
			owner.POST("/bookings/:id/confirm", bh.Confirm)

			secured.POST("/bookings/:id/cancel", bh.Cancel)
		}
		pay := v1.Group("/payments")
		pay.Use(middlewares.JWTAuth())
		{
			ph := handlers.NewPaymentHandler(c)
			pay.POST("/charges/card", ph.CreateCardCharge)
			pay.POST("/charges/source", ph.CreateSourceCharge)
			pay.GET("/charges/:id", ph.GetCharge)
		}

	}

	log.Println("api-gateway on", cfg.GatewayHTTPAddr)
	r.Run(cfg.GatewayHTTPAddr)
}
