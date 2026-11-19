package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/axelarnetwork/utils/log"
)

// Connect connects to tofnd gRPC Server
func Connect(host string, port string, timeout time.Duration) (*grpc.ClientConn, error) {
	serverAddr := fmt.Sprintf("%s:%s", host, port)
	log.Infof("initiate connection to tofnd gRPC server: %s", serverAddr)

	// The channel is plaintext and unauthenticated: any caller that can
	// reach the port can request signatures for any key the tofnd daemon
	// holds. That is safe for the default loopback deployment, but if the
	// daemon is split onto another host or container network, anything on
	// that network gains the same power. Warn loudly so the exposure is a
	// deliberate operator choice rather than a silent default.
	if !isLoopbackHost(host) {
		log.Errorf("tofnd connection to %s is plaintext and unauthenticated; "+
			"any party that can reach this address can request signatures with the validator's keys. "+
			"Restrict network access to this port.", serverAddr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return grpc.DialContext(ctx, serverAddr, grpc.WithInsecure(), grpc.WithBlock())
}

func isLoopbackHost(host string) bool {
	h := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if h == "localhost" {
		return true
	}
	if ip := net.ParseIP(strings.Trim(h, "[]")); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
