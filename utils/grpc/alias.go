package grpc

import (
	protogrpc "github.com/cosmos/gogoproto/grpc"
)

//go:generate moq -pkg mock -out mock/grpc.go . Server

// Server alias for mocking
type Server protogrpc.Server
