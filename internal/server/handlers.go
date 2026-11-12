package server

import (
	"encoding/json"
	"net/http"

	"github.com/mercyyy12/payproxy/internal/evaluator"
	"github.com/mercyyy12/payproxy/internal/guardrail"
	"github.com/mercyyy12/payproxy/internal/paypal"
)

// JSON payload sent by the AI agent
type AgentOrderRequest struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Merchant string `json:"merchant"`
	Purpose  string `json:"purpose"`
}

// JSON returned back to the agent
type AgentOrderResponse struct {
	Status      string `json:"status"`
	OrderID     string `json:"order_id"`
	CheckoutURL string `json:"checkout_url"`
}

// receives incoming agent requests and creates a PayPal order
func HandleCreateOrder(paypalClient *paypal.Client, guard *guardrail.Manager, ai *evaluator.Evaluator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// decode incoming JSON
		var req AgentOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Err: "Invalid JSON",
			})
			return
		}

		// validate required fields
		if req.Amount == "" || req.Currency == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Err: "Amount and Currency is required.",
			})
			return
		}

		if err := guard.CheckAndReserve(req.Amount); err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(&ErrorResponse{
				Err: "Guardrail blocked transaction: " + err.Error(),
			})
			return
		}

		// run AI fraud inspection
		assessment, err := ai.Assess(r.Context(), req.Merchant, req.Purpose)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(&ErrorResponse{
				Err: "Transaction halted: " + err.Error(),
			})
			return
		}

		if assessment.Flagged || assessment.RiskScore > 50 {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(&ErrorResponse{
				Err: "AI Security Block: " + assessment.Reason,
			})
			return
		}

		// create order via PayPal
		order, err := paypalClient.CreateOrder(r.Context(), req.Amount, req.Currency, req.Purpose)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Err: "Failed to create paypal order " + err.Error(),
			})
			return
		}

		// return the approval checkout URL
		resp := AgentOrderResponse{
			Status:      order.Status,
			OrderID:     order.ID,
			CheckoutURL: order.GetApprovalLink(),
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}
