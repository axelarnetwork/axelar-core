package grpc

import "testing"

func TestIsLoopbackHost(t *testing.T) {
	loopback := []string{"localhost", "LOCALHOST", "localhost.", "127.0.0.1", "127.1.2.3", "::1", "[::1]", " 127.0.0.1 "}
	for _, h := range loopback {
		if !isLoopbackHost(h) {
			t.Errorf("expected %q to be recognized as loopback", h)
		}
	}

	remote := []string{"tofnd", "10.0.0.5", "192.168.1.10", "tofnd.svc.cluster.local", "0.0.0.0", ""}
	for _, h := range remote {
		if isLoopbackHost(h) {
			t.Errorf("expected %q to NOT be recognized as loopback", h)
		}
	}
}
