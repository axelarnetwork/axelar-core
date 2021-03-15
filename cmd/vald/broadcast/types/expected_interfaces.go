package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

//go:generate moq -out ./mock/broadcast.go -pkg mock -stub . Client Msg
// go:generate moq -out ./mock/expected_interfaces.go -pkg mock -stub . Keybase Client KVStore Info Msg

// Msg is a wrapped interface for moq generation
type (
	Msg sdk.Msg
)

// Client represents a tendermint/Cosmos client
type Client interface {
	BroadcastTxSync(tx auth.StdTx) (*coretypes.ResultBroadcastTx, error)
	GetAccountNumberSequence(addr sdk.AccAddress) (uint64, uint64, error)
}

// Sign returns a signature for the given message from the account associated with the given address
type Sign func(from sdk.AccAddress, msg auth.StdSignMsg) (auth.StdSignature, error)
