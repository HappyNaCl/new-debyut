package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/HappyNaCl/new-debyut/internal/common/cache"
	"github.com/HappyNaCl/new-debyut/internal/common/database"
	"github.com/HappyNaCl/new-debyut/internal/common/grpcclient"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/otp"
	"github.com/HappyNaCl/new-debyut/internal/common/rabbitmq"
	"github.com/HappyNaCl/new-debyut/internal/common/storage"
	topics "github.com/HappyNaCl/new-debyut/internal/constants"
	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
	"github.com/HappyNaCl/new-debyut/internal/users/handlers/consumers"
	"github.com/HappyNaCl/new-debyut/internal/users/repositories"
	"github.com/HappyNaCl/new-debyut/internal/users/services"
	"google.golang.org/grpc"
)

func main() {
	if err := l.Init(); err != nil {
		panic("Failed to initialize logger.")
	}
	defer l.Sync()

	db, err := database.ConnectPostgres(os.Getenv("USER_SERVICE_DB"))
	if err != nil {
		panic("Failed to connect to user service DB.")
	}

	redis, err := cache.ConnectRedis()
	if err != nil {
		panic("Failed to connect to redis")
	}

	conn, ch, err := rabbitmq.Connect()
	if err != nil {
		l.Logger.Fatal("Failed to connect to RabbitMQ.")
	}
	defer conn.Close()
	defer ch.Close()

	pubCh, err := conn.Channel()
	if err != nil {
		l.Logger.Fatal("Failed to open publisher channel: " + err.Error())
	}
	defer pubCh.Close()
	publisher := rabbitmq.NewPublisher(pubCh)

	store, err := storage.Connect(context.Background())
	if err != nil {
		l.Logger.Fatal("Failed to connect to object storage: " + err.Error())
	}

	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository, redis, store)
	applicationRepository := repositories.NewOrganizerApplicationRepository(db)
	emailOtpCache := otp.NewCache(redis, "organization-email")
	emailTokenCache := repositories.NewOrganizationEmailTokenCache(redis)
	organizationRepository := repositories.NewOrganizationRepository(db)
	applicationService := services.NewOrganizerApplicationService(applicationRepository, organizationRepository, emailOtpCache, emailTokenCache, publisher, store)
	invitationRepository := repositories.NewInvitationRepository(db)
	authConn, err := grpcclient.NewGrpcClient(os.Getenv("AUTH_SERVICE_ADDR"))
	if err != nil {
		l.Logger.Fatal("Failed to connect to auth service: " + err.Error())
	}
	defer authConn.Close()
	accounts := services.NewAuthAccountDirectory(auth.NewAuthServiceClient(authConn))

	organizationService := services.NewOrganizationService(organizationRepository, invitationRepository, userRepository, accounts, publisher, store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := consumers.NewUserRegisteredConsumer(userService)
	go func() {
		if err := rabbitmq.Consume(ctx, ch, "user-service.user-registered", topics.USER_REGISTERED_TOPIC, consumer.Handle); err != nil && ctx.Err() == nil {
			l.Logger.Warn("Consumer stopped: " + err.Error())
		}
	}()

	grpcServer := grpc.NewServer()

	go func() {
		server := NewUserGrpcServer(os.Getenv("USER_SERVICE_ADDR"))
		l.Logger.Info("User gRPC server is running on" + os.Getenv("USER_SERVICE_ADDR"))

		server.Run(grpcServer, userService, applicationService, organizationService)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	grpcServer.GracefulStop()
}
