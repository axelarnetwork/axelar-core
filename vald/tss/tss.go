// Package tss provides the tofnd gRPC connection functionality for vald.
package tss

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/axelarnetwork/utils/log"
)

// Connect connects to tofnd gRPC Server
func Connect(host string, port string, timeout time.Duration) (*grpc.ClientConn, error) {
	tofndServerAddress := host + ":" + port
	log.Infof("initiate connection to tofnd gRPC server: %s", tofndServerAddress)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return grpc.DialContext(ctx, tofndServerAddress, grpc.WithInsecure(), grpc.WithBlock())
}
