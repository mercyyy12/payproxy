package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Request structs
type Amount struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type PurchaseUnit struct {
	Amount      Amount `json:"amount"`
	Description string `json:"description"`
}

type CreateOrderRequest struct {
	Intent        string         `json:"intent"`
	PurchaseUnits []PurchaseUnit `json:"purchase_units"`
}

// Response structs
type Link struct {
	Href   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

type OrderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Links  []Link `json:"links"`
}

// calls PayPal's Orders API to create a new transaction
func (c *Client) CreateOrder(ctx context.Context, amount, currency, description string) (*OrderResponse, error) {
	// get access token
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// build order payload
	orderReq := CreateOrderRequest{
		Intent: "CAPTURE",
		PurchaseUnits: []PurchaseUnit{
			{
				Amount: Amount{
					CurrencyCode: currency,
					Value:        amount,
				},
				Description: description,
			},
		},
	}

	// serialize struct to json
	jsonData, err := json.Marshal(orderReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order request: %w", err)
	}

	// build http request
	url := c.baseURL + "/v2/checkout/orders"
	body := bytes.NewReader(jsonData)

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create order request: %w", err)
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send order request: %w", err)
	}
	defer resp.Body.Close()

	// check if order creation succeeded (201 Created)
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("paypal order error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// decode response
	var orderResp OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, fmt.Errorf("failed to decode order response: %w", err)
	}

	return &orderResp, nil
}

// extracts the buyer checkout url from response links
func (o *OrderResponse) GetApprovalLink() string {
	for _, link := range o.Links {
		if link.Rel == "approve" || link.Rel == "payer-action" {
			return link.Href
		}
	}
	return ""
}
