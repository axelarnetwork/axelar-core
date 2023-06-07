package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/axelarnetwork/utils/log"
)

// Connect connects to tofnd gRPC Server
func Connect(host string, port string, timeout time.Duration) (*grpc.ClientConn, error) {
	serverAddr := fmt.Sprintf("%s:%s", host, port)
	log.Infof("initiate connection to tofnd gRPC server: %s", serverAddr)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return grpc.DialContext(ctx, serverAddr, grpc.WithInsecure(), grpc.WithBlock())
}
