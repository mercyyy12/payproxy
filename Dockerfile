FROM golang:1.26-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod ./
RUN go mod download

# Copy source code and build statically linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/payproxy cmd/server/main.go

# minimal runtime image
FROM alpine:3.20

# Add SSL certificates so Go can talk HTTPS to PayPal and Gemini
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy only the compiled binary from builder
COPY --from=builder /app/payproxy .

EXPOSE 8080

CMD ["./payproxy"]
