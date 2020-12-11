package types

import (
	"github.com/axelarnetwork/tssd/pb"
)

type Stream interface {
	Send(in *pb.MessageIn) error
	Recv() (*pb.MessageOut, error)
	CloseSend() error
}
