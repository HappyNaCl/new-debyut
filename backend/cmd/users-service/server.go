package main

import (
	"net"

	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	usershandlers "github.com/HappyNaCl/new-debyut/internal/users/handlers/grpc"
	"github.com/HappyNaCl/new-debyut/internal/users/services"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type UserGrpcServer struct {
	addr string
}

func NewUserGrpcServer(addr string) *UserGrpcServer {
	return &UserGrpcServer{
		addr: addr,
	}
}

func (s *UserGrpcServer) Run(
	server *grpc.Server,
	userService *services.UserService,
	applicationService *services.OrganizerApplicationService,
	organizationService *services.OrganizationService,
) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		l.Logger.Fatal("Failed to listen", zap.String("address", s.addr))
		return err
	}

	usershandlers.NewUserGrpcHandler(server, userService)
	usershandlers.NewOrganizerApplicationGrpcHandler(server, applicationService)
	usershandlers.NewOrganizationGrpcHandler(server, organizationService)

	l.Logger.Info("User gRPC server is running!")
	return server.Serve(lis)
}
