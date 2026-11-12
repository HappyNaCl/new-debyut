FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /event-service ./cmd/event-service

FROM gcr.io/distroless/static-debian12

COPY --from=builder /event-service /event-service

ENTRYPOINT ["/event-service"]
