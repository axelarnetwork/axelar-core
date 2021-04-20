package types

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

//go:generate moq -out ./mock/broadcast.go -pkg mock -stub . Client Msg Broadcaster
// go:generate moq -out ./mock/interfaces.go -pkg mock -stub . Keybase Client KVStore Info Msg

// Msg is a wrapped interface for moq generation
type (
	Msg sdk.Msg
)

// Client represents a tendermint/Cosmos client
type Client interface {
	BroadcastTxSync(tx legacytx.StdTx) (*coretypes.ResultBroadcastTx, error)
	GetAccountNumberSequence(clientCtx client.Context, addr sdk.AccAddress) (accNum uint64, accSeq uint64, err error)
}

// Sign returns a signature for the given message from the account associated with the given address
type Sign func(from sdk.AccAddress, msg legacytx.StdSignMsg) (legacytx.StdSignature, error)

// Broadcaster interface allows the submission of messages to the axelar network
type Broadcaster interface {
	Broadcast(msgs ...sdk.Msg) error
}
