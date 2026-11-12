package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mercyyy12/payproxy/internal/guardrail"
	"github.com/mercyyy12/payproxy/internal/paypal"
	"github.com/mercyyy12/payproxy/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting on port: " + port)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.HandleHealth)

	// initialize paypal client
	paypalClient := paypal.NewClient(
		os.Getenv("PAYPAL_CLIENT_ID"),
		os.Getenv("PAYPAL_CLIENT_SECRET"),
		"https://api-m.sandbox.paypal.com",
	)
	guard := guardrail.NewManager(50.00, 100.00)
	mux.HandleFunc("POST /api/v1/orders", server.HandleCreateOrder(paypalClient, guard))

	serverVar := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		err := serverVar.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("can't connect to server", "err", err.Error())
			log.Fatal(err)
		}
	}()

	// Wait for SIGINT (Ctrl+C) or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	// give the server a 10-second deadline to finish
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	serverVar.Shutdown(ctxTimeout)
	slog.Info("Server stopped cleanly")
}
