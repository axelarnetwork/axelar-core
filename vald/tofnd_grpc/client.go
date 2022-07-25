package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
)

// Connect connects to tofnd gRPC Server
func Connect(host string, port string, timeout time.Duration, logger log.Logger) (*grpc.ClientConn, error) {
	serverAddr := fmt.Sprintf("%s:%s", host, port)
	logger.Info(fmt.Sprintf("initiate connection to tofnd gRPC server: %s", serverAddr))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return grpc.DialContext(ctx, serverAddr, grpc.WithInsecure(), grpc.WithBlock())
}
