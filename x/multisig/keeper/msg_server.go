package keeper

import (
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper types.Keeper) types.MsgServiceServer {
	return msgServer{
		Keeper: keeper,
	}
}
