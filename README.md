# PayProxy

A safety layer between AI agents and PayPal.

PayProxy sits in front of PayPal and decides whether an AI agent is allowed to spend money. The agent never sees PayPal credentials or a credit card. It sends a payment request to PayProxy, and PayProxy checks it against spending limits and an AI fraud check before it creates a real PayPal order.

Built for the PayPal AI Hackathon.

## The problem

AI agents are starting to buy things on their own: servers, API credits, software subscriptions. Giving an agent a card or a PayPal login is risky. An agent can get stuck in a loop and spend too much. It can also make a bad decision, or be tricked into paying for something that has nothing to do with its task.

## The solution

PayProxy gives the agent one narrow door to go through. Every payment request must pass three checks:

1. Spending limits. A single payment cannot be more than $50, and total spending cannot go over $100.
2. AI fraud check. Google Gemini reads the merchant name and the stated purpose and decides if they make sense together. For example, "Steam Video Games" with the purpose "Database renewal" is flagged and blocked.
3. PayPal order. Only if both checks pass, PayProxy creates an order in PayPal and returns a checkout link. The money moves only when a human approves that link.

If any check fails, or the AI check cannot run, the request is rejected. PayProxy fails closed, which means "when unsure, do not pay".

## How it works

```text
AI agent
   |
   |  POST /api/v1/orders
   v
PayProxy
   |
   |-- 1. Validate the request
   |-- 2. Check spending limits
   |-- 3. Ask Gemini if the merchant matches the purpose
   |-- 4. Create a PayPal order
   |
   v
PayPal Sandbox (Orders API v2)
   |
   v
Checkout link returned to the agent, approved by a human
```

## How PayPal and AI are used

PayPal:
- OAuth2 client credentials to get an access token
- Orders API v2 to create an order with intent CAPTURE
- The approval link from the response is returned to the agent, so a person stays in control of the final payment

AI:
- Gemini is asked to act as a strict security auditor
- It returns structured JSON with a risk score from 0 to 100, a flagged true/false value, and a one-sentence reason
- A request is blocked if it is flagged or the score is above 50

## API

### Create an order

`POST /api/v1/orders`

Request body:

```json
{
  "amount": "15.00",
  "currency": "USD",
  "merchant": "Render",
  "purpose": "Deploy backend service"
}
```

Success response (HTTP 201):

```json
{
  "status": "CREATED",
  "order_id": "5O190127TN364715T",
  "checkout_url": "https://www.sandbox.paypal.com/checkoutnow?token=..."
}
```

Error responses:

| Status | Meaning |
| ------ | ------- |
| 400 | Invalid JSON, or amount or currency is missing |
| 403 | Blocked by the spending limits or by the AI fraud check |
| 503 | The AI check could not run, so the payment was halted |
| 500 | PayPal returned an error while creating the order |

### Health check

`GET /health` returns `{"status":"ok"}`.

## Demo

The script `scripts/simulate_agent.sh` sends three requests to a running server:

| Test | Request | Expected result |
| ---- | ------- | --------------- |
| 1 | Render, $15, "Deploy backend service" | Approved, PayPal order created |
| 2 | AWS, $75, "GPU server" | Blocked, over the $50 single payment limit |
| 3 | Steam Video Games, $20, "Database renewal" | Blocked by the AI, merchant does not match purpose |

## Getting started

### What you need

- Go 1.26 or newer
- A PayPal sandbox app (Client ID and Secret) from https://developer.paypal.com/
- A Google Gemini API key from https://aistudio.google.com/

### Run locally

```bash
git clone https://github.com/mercyyy12/payproxy.git
cd payproxy

export GEMINI_API_KEY="your_gemini_key"
export PAYPAL_CLIENT_ID="your_paypal_client_id"
export PAYPAL_CLIENT_SECRET="your_paypal_client_secret"

go run cmd/server/main.go
```

The server starts on port 8080. Set `PORT` to use a different port.

### Try it

In a second terminal, with the server running:

```bash
chmod +x scripts/simulate_agent.sh
./scripts/simulate_agent.sh
```

Or send a single request yourself:

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"amount":"15.00","currency":"USD","merchant":"Render","purpose":"Deploy backend service"}'
```

Open the `checkout_url` from the response in a browser and log in with a PayPal sandbox buyer account to see the order.

### Environment variables

| Name | Required | Description |
| ---- | -------- | ----------- |
| PAYPAL_CLIENT_ID | Yes | PayPal sandbox Client ID |
| PAYPAL_CLIENT_SECRET | Yes | PayPal sandbox Secret |
| GEMINI_API_KEY | Yes | Gemini API key. Without it, every payment is rejected |
| PORT | No | Port to listen on. Defaults to 8080 |

### Run the tests

```bash
PAYPAL_CLIENT_ID="your_client_id" \
PAYPAL_CLIENT_SECRET="your_secret" \
go test -v ./internal/paypal
```

This test creates a real order in the PayPal sandbox. It is skipped automatically if the credentials are not set.

## Docker

```bash
docker build -t payproxy:latest .

docker run -p 8080:8080 \
  -e GEMINI_API_KEY="..." \
  -e PAYPAL_CLIENT_ID="..." \
  -e PAYPAL_CLIENT_SECRET="..." \
  payproxy:latest
```

The image uses a multi-stage build, so the final image only contains the compiled binary on a small Alpine base.

## Project structure

```text
cmd/server/            Entry point, wiring, graceful shutdown
internal/server/       HTTP handlers (orders and health)
internal/guardrail/    Spending limit rules, safe for concurrent requests
internal/evaluator/    Gemini fraud check
internal/paypal/       PayPal OAuth2 client and Orders API
scripts/               Demo script
```

## Design choices

- Fail closed. Missing API key, a Gemini error, or a bad response all result in a rejected payment.
- Limits are checked first. They are cheap and need no network call, so obvious bad requests never reach the AI or PayPal.
- Thread safe. The spending counter is protected by a mutex, so many agents sending requests at once cannot skip the limit.
- Human approval. PayProxy creates the order but does not capture it. A person approves the final payment.
- Standard library only. No third party dependencies.
- Production basics. JSON logs with `log/slog`, timeouts on all outside calls, and a clean shutdown on SIGINT and SIGTERM.

## Current limitations

- Spending totals are kept in memory. They reset when the server restarts, and there is no automatic daily reset yet.
- Limits are fixed in code ($50 per payment, $100 total).
- Limits assume USD amounts.
- The endpoint has no authentication yet. Run it only on a trusted network.
- Orders are created but not captured by PayProxy.

## Ideas for next steps

- Per agent API keys and per agent budgets
- Store spending in a database, with a real daily reset
- Capture orders after approval and listen for PayPal webhooks
- Merchant allow lists and deny lists
- A small dashboard that shows approved and blocked requests

## License

MIT. See the LICENSE file.
