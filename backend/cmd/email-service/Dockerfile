FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /email-service ./cmd/email-service

FROM gcr.io/distroless/static-debian12

COPY --from=builder /email-service /email-service

ENTRYPOINT ["/email-service"]
