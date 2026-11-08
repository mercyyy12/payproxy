package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mercyyy12/payproxy/internal/paypal"
)

type HealthResponse struct {
	Health string `json:"status"`
}

type ErrorResponse struct {
	Err string `json:"error"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	var resp HealthResponse
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Err: "Method not allowed",
		})
		return
	}

	resp = HealthResponse{
		Health: "ok",
	}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting on port: " + port)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)

	serverVar := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Initialize PayPal client with credentials
	paypalClient := paypal.NewClient(
		os.Getenv("PAYPAL_CLIENT_ID"),
		os.Getenv("PAYPAL_CLIENT_SECRET"),
		"https://api-m.sandbox.paypal.com",
	)

	authCtx, cancelAuth := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelAuth()

	//  Ask PayPal for an access token
	token, err := paypalClient.GetAccessToken(authCtx)
	if err != nil {
		slog.Error("PayPal Auth Failed", "err", err)
	} else {
		// Show the first 20 characters of the token to prove we got it
		slog.Info("PayPal Auth Successful!", "token_preview", token[:20]+"...")
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
