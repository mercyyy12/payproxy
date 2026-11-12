package main

import (
	"context"
	_ "embed"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mercyyy12/payproxy/internal/evaluator"
	"github.com/mercyyy12/payproxy/internal/guardrail"
	"github.com/mercyyy12/payproxy/internal/paypal"
	"github.com/mercyyy12/payproxy/internal/server"
)

//go:embed index.html
var indexHTML []byte

//go:embed styles.css
var stylesCSS []byte

//go:embed script.js
var scriptJS []byte

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

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})

	mux.HandleFunc("GET /styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write(stylesCSS)
	})

	mux.HandleFunc("GET /script.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Write(scriptJS)
	})

	// initialize paypal client
	paypalClient := paypal.NewClient(
		os.Getenv("PAYPAL_CLIENT_ID"),
		os.Getenv("PAYPAL_CLIENT_SECRET"),
		"https://api-m.sandbox.paypal.com",
	)
	guard := guardrail.NewManager(50.00, 100.00)
	ai := evaluator.NewEvaluator(os.Getenv("GEMINI_API_KEY"))
	mux.HandleFunc("POST /api/v1/orders", server.HandleCreateOrder(paypalClient, guard, ai))

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
