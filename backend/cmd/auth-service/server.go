package main

import (
	"net"

	authhandlers "github.com/HappyNaCl/new-debyut/internal/auth/handlers/grpc"
	"github.com/HappyNaCl/new-debyut/internal/auth/services"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type AuthGrpcServer struct {
	addr string
}

func NewAuthGrpcServer(addr string) *AuthGrpcServer {
	return &AuthGrpcServer{
		addr: addr,
	}
}

func (s *AuthGrpcServer) Run(
	server *grpc.Server,
	authService *services.AuthService,
	tokenService *services.TokenService,
) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		l.Logger.Fatal("Failed to listen", zap.String("address", s.addr))
		return err
	}

	authhandlers.NewAuthGrpcHandler(server, authService, tokenService)

	l.Logger.Info("Auth gRPC server is running!")
	return server.Serve(lis)
}
