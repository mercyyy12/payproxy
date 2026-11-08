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

	go func() {
		err := serverVar.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("can't connect to server", "err", err.Error())
			log.Fatal(err)
		}
	}()

	// Wait for SIGINT (Ctrl+C) or SIGTERM (Linux kill command)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	// give the server a 10-second deadline to finish
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	serverVar.Shutdown(ctxTimeout)
	slog.Info("Server stopped cleanly")
}
