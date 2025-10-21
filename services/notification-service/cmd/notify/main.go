package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/you/badminton-booking/services/notification-service/internal/notifier"
	"github.com/you/badminton-booking/services/notification-service/internal/worker"
)

func mustEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func main() {
	exchanges := parseCSV(os.Getenv("NOTIFY_EXCHANGES"))
	if len(exchanges) == 0 {
		// fallback: ใช้ค่าเดิม MQ_EXCHANGE ตัวเดียว
		exchanges = []string{mustEnv("MQ_EXCHANGE", "booking.exchange")}
	}

	cfg := worker.Config{
		RabbitURL: mustEnv("RABBIT_URL", "amqp://guest:guest@rabbitmq:5672/"),
		// Exchange เดิมเก็บไว้เพื่อ backward-compat (แต่เราส่ง Exchanges ให้แล้ว)
		Exchange:    "",
		Exchanges:   exchanges,
		Queue:       mustEnv("NOTIFY_QUEUE", "notification.q"),
		Bindings:    parseCSV(mustEnv("NOTIFY_BINDINGS", "booking.*,payment.*")),
		Prefetch:    16,
		UseDLX:      true,
		DLXName:     mustEnv("NOTIFY_DLX", "notification.dlx"),
		DLXQueue:    mustEnv("NOTIFY_DLQ", "notification.q.dlq"),
		ServiceName: "notification-service",
	}

	n := notifier.NewConsole()
	cons := worker.NewConsumer(cfg, n)

	for {
		if err := cons.Connect(); err != nil {
			log.Printf("[notify] connect failed: %v; retry in 2s", err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	defer cons.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := cons.Run(ctx); err != nil {
			log.Printf("[notify] run error: %v", err)
		}
	}()

	log.Printf("[notify] started. queue=%s exchanges=%v bindings=%v",
		cfg.Queue, cfg.Exchanges, cfg.Bindings)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	cancel()
	time.Sleep(200 * time.Millisecond)
}
