package types

import (
	"github.com/axelarnetwork/tssd/pb"
)

type Stream interface {
	Send(in *pb.MessageIn) error
	Recv() (*pb.MessageOut, error)
	CloseSend() error
}

type MasterKey struct {
	BlockHeight int64
	PK          []byte
}
