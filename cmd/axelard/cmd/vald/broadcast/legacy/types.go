package legacy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// SignFn returns a signature for the given message from the account associated with the given address
type SignFn func(from sdk.AccAddress, msg legacytx.StdSignMsg) (legacytx.StdSignature, error)

// Broadcaster submits legacy transactions to a tendermint node
type Broadcaster interface {
	BroadcastSync(tx legacytx.StdTx) (*sdk.TxResponse, error)
}

// Client provides an interface for broadcasting signed legacy transactions
type Client interface {
	GetAccountNumberSequence(addr sdk.AccAddress) (uint64, uint64, error)
	BroadcastTxSync(stdTx legacytx.StdTx) (*sdk.TxResponse, error)
	BroadcastTx(stdTx legacytx.StdTx, mode string) (sdk.TxResponse, error)
	GetAccount(address sdk.AccAddress) (authtypes.BaseAccount, error)
}
