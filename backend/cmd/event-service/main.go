package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/HappyNaCl/new-debyut/internal/common/cache"
	"github.com/HappyNaCl/new-debyut/internal/common/database"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/event/repositories"
	"github.com/HappyNaCl/new-debyut/internal/event/services"
	"google.golang.org/grpc"
)

func main() {
	if err := l.Init(); err != nil {
		panic("Failed to initialize logger.")
	}
	defer l.Sync()

	db, err := database.ConnectPostgres(os.Getenv("EVENT_SERVICE_DB"))
	if err != nil {
		l.Logger.Fatal("Failed to connect to event service DB: " + err.Error())
	}

	redis, err := cache.ConnectRedis()
	if err != nil {
		l.Logger.Fatal("Failed to connect to Redis: " + err.Error())
	}

	categoryRepository := repositories.NewCategoryRepository(db)
	categoryService := services.NewCategoryService(categoryRepository, redis)

	grpcServer := grpc.NewServer()
	addr := os.Getenv("EVENT_SERVICE_ADDR")
	go func() {
		server := NewEventGrpcServer(addr)
		l.Logger.Info("Event gRPC server is running on " + addr)

		server.Run(grpcServer, categoryService)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcServer.GracefulStop()
}
