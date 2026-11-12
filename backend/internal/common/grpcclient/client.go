package grpcclient

import (
	log "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGrpcClient(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	log.Logger.Info("gRPC client is connected to " + addr)
	return conn, nil
}
