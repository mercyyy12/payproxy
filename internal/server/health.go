package server

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Health string `json:"status"`
}

type ErrorResponse struct {
	Err string `json:"error"`
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
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
