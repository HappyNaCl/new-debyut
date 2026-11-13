package main

import (
	"net"

	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	eventhandlers "github.com/HappyNaCl/new-debyut/internal/event/handlers/grpc"
	"github.com/HappyNaCl/new-debyut/internal/event/services"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type EventGrpcServer struct {
	addr string
}

func NewEventGrpcServer(addr string) *EventGrpcServer {
	return &EventGrpcServer{
		addr: addr,
	}
}

func (s *EventGrpcServer) Run(server *grpc.Server, categoryService *services.CategoryService, eventService *services.EventService) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		l.Logger.Fatal("Failed to listen", zap.String("address", s.addr))
		return err
	}

	eventhandlers.NewEventGrpcHandler(server, categoryService, eventService)

	l.Logger.Info("Event gRPC server is running!")
	return server.Serve(lis)
}
