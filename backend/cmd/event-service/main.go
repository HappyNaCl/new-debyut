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
	"github.com/HappyNaCl/new-debyut/internal/common/meilisearch"
	"github.com/HappyNaCl/new-debyut/internal/common/rabbitmq"
	"github.com/HappyNaCl/new-debyut/internal/common/storage"
	topics "github.com/HappyNaCl/new-debyut/internal/constants"
	"github.com/HappyNaCl/new-debyut/internal/event/consumers"
	"github.com/HappyNaCl/new-debyut/internal/event/repositories"
	"github.com/HappyNaCl/new-debyut/internal/event/search"
	"github.com/HappyNaCl/new-debyut/internal/event/services"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	if err := l.Init(); err != nil {
		panic("Failed to initialize logger.")
	}
	defer l.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.ConnectPostgres(os.Getenv("EVENT_SERVICE_DB"))
	if err != nil {
		l.Logger.Fatal("Failed to connect to event service DB: " + err.Error())
	}

	redis, err := cache.ConnectRedis()
	if err != nil {
		l.Logger.Fatal("Failed to connect to Redis: " + err.Error())
	}

	store, err := storage.Connect(ctx)
	if err != nil {
		l.Logger.Fatal("Failed to connect to object storage: " + err.Error())
	}

	meili, err := meilisearch.Connect(ctx)
	if err != nil {
		l.Logger.Fatal("Failed to connect to Meilisearch: " + err.Error())
	}

	userConn, err := grpcclient.NewGrpcClient(os.Getenv("USER_SERVICE_ADDR"))
	if err != nil {
		l.Logger.Fatal("Failed to create users service client: " + err.Error())
	}
	defer userConn.Close()

	conn, ch, err := rabbitmq.Connect()
	if err != nil {
		l.Logger.Fatal("Failed to connect to RabbitMQ: " + err.Error())
	}
	defer conn.Close()
	defer ch.Close()

	categoryRepository := repositories.NewCategoryRepository(db)
	categoryService := services.NewCategoryService(categoryRepository, redis)

	eventRepository := repositories.NewEventRepository(db)
	outboxRepository := repositories.NewSearchOutboxRepository(db)
	searcher := search.NewSearcher(meili)
	eventService := services.NewEventService(eventRepository, categoryRepository, store, searcher)

	indexer := search.NewIndexer(meili, outboxRepository, eventRepository,
		search.NewOrganizationDirectory(users.NewOrganizationServiceClient(userConn)))
	if err := indexer.Setup(ctx); err != nil {
		l.Logger.Fatal("Failed to set up search: " + err.Error())
	}
	go indexer.Run(ctx)

	renamedConsumer := consumers.NewOrganizationRenamedConsumer(outboxRepository)
	go func() {
		if err := rabbitmq.Consume(ctx, ch, "event-service.organization-renamed", topics.ORGANIZATION_RENAMED_TOPIC, renamedConsumer.Handle); err != nil && ctx.Err() == nil {
			l.Logger.Warn("Organization renamed consumer stopped", zap.Error(err))
		}
	}()

	grpcServer := grpc.NewServer()
	addr := os.Getenv("EVENT_SERVICE_ADDR")
	go func() {
		server := NewEventGrpcServer(addr)
		l.Logger.Info("Event gRPC server is running on " + addr)

		server.Run(grpcServer, categoryService, eventService)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	grpcServer.GracefulStop()
}
