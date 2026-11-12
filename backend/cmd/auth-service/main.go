package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/HappyNaCl/new-debyut/internal/auth/handlers/consumers"
	"github.com/HappyNaCl/new-debyut/internal/auth/repositories"
	"github.com/HappyNaCl/new-debyut/internal/auth/services"
	"github.com/HappyNaCl/new-debyut/internal/common/cache"
	"github.com/HappyNaCl/new-debyut/internal/common/database"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/otp"
	"github.com/HappyNaCl/new-debyut/internal/common/rabbitmq"
	topics "github.com/HappyNaCl/new-debyut/internal/constants"
	"google.golang.org/grpc"
)

func main() {
	if err := l.Init(); err != nil {
		panic("Failed to initialize logger!")
	}
	defer l.Sync()

	db, err := database.ConnectPostgres(os.Getenv("AUTH_SERVICE_DB"))
	if err != nil {
		l.Logger.Fatal("Failed to connect to DB: " + err.Error())
	}

	redis, err := cache.ConnectRedis()
	if err != nil {
		l.Logger.Fatal("Failed to connect to Redis: " + err.Error())
	}

	conn, ch, err := rabbitmq.Connect()
	if err != nil {
		l.Logger.Fatal(err.Error())
	}
	defer conn.Close()
	defer ch.Close()

	consumerCh, err := conn.Channel()
	if err != nil {
		l.Logger.Fatal("Failed to open consumer channel: " + err.Error())
	}
	defer consumerCh.Close()

	publisher := rabbitmq.NewPublisher(ch)

	registerOtpCache := otp.NewCache(redis, "register")
	loginOtpCache := otp.NewCache(redis, "login")
	registrationTokenCache := repositories.NewRegistrationTokenCache(redis)
	credentialRepo := repositories.NewCredentialRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)

	tokenService := services.NewTokenService(tokenRepo)
	authService := services.NewAuthService(registerOtpCache, loginOtpCache, registrationTokenCache, credentialRepo, tokenService, publisher)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := consumers.NewOrganizerStatusChangedConsumer(authService)
	go func() {
		if err := rabbitmq.Consume(ctx, consumerCh, "auth-service.organizer-status-changed", topics.ORGANIZER_STATUS_CHANGED_TOPIC, consumer.Handle); err != nil && ctx.Err() == nil {
			l.Logger.Warn("Consumer stopped: " + err.Error())
		}
	}()

	grpcServer := grpc.NewServer()
	addr := os.Getenv("AUTH_SERVICE_ADDR")
	go func() {
		server := NewAuthGrpcServer(addr)
		l.Logger.Info("Auth gRPC server is running on " + addr)

		server.Run(grpcServer, authService, tokenService)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	grpcServer.GracefulStop()
}
