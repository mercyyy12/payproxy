package paypal

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestCreateOrder(t *testing.T) {
	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	clientSecret := os.Getenv("PAYPAL_CLIENT_SECRET")

	// skip test if environment variables are not set
	if clientID == "" || clientSecret == "" {
		t.Skip("skipping test; PAYPAL_CLIENT_ID or PAYPAL_CLIENT_SECRET not set")
	}

	client := NewClient(clientID, clientSecret, "https://api-m.sandbox.paypal.com")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// call CreateOrder
	order, err := client.CreateOrder(ctx, "10.00", "USD", "Cloud Server Hosting")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// verify response fields
	if order.ID == "" {
		t.Errorf("expected non-empty order ID")
	}

	if order.Status != "CREATED" {
		t.Errorf("expected status 'CREATED', got: %s", order.Status)
	}

	approvalLink := order.GetApprovalLink()
	if approvalLink == "" {
		t.Errorf("expected valid approval link, got empty string")
	}

	t.Logf("Test Passed! Order ID: %s", order.ID)
	t.Logf("Checkout URL: %s", approvalLink)
}
