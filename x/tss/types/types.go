package types

import (
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
)

// Stream is the abstracted communication stream with tofnd
type Stream interface {
	Send(in *tofnd.MessageIn) error
	Recv() (*tofnd.MessageOut, error)
	CloseSend() error
}
