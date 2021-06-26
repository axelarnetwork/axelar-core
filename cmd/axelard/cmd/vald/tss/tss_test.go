package tss

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCTimeout(t *testing.T) {
	t.Run("connect to server", func(t *testing.T) {
		listener := bufconn.Listen(1)
		server := grpc.NewServer()
		go func() {
			if err := server.Serve(listener); err != nil {
				panic(err)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		_, err := grpc.DialContext(
			ctx,
			"",
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
			grpc.WithInsecure(),
			grpc.WithBlock(),
		)
		assert.NoError(t, err)
	})
	t.Run("timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithBlock())
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}
