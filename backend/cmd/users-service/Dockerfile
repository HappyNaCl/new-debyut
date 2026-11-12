FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /users-service ./cmd/users-service

FROM gcr.io/distroless/static-debian12

COPY --from=builder /users-service /users-service

ENTRYPOINT ["/users-service"]
