package btc

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types"
	rpc3 "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/btc/rpc"
)

// Mgr manages all communication with Bitcoin
type Mgr struct {
	cliCtx      sdkClient.Context
	logger      log.Logger
	broadcaster types.Broadcaster
	rpc         rpc3.Client
	cdc         *codec.LegacyAmino
}

// NewMgr returns a new Mgr instance
func NewMgr(rpc rpc3.Client, cliCtx sdkClient.Context, broadcaster types.Broadcaster, logger log.Logger, cdc *codec.LegacyAmino) *Mgr {
	return &Mgr{
		rpc:         rpc,
		cliCtx:      cliCtx,
		logger:      logger.With("listener", "btc"),
		broadcaster: broadcaster,
		cdc:         cdc,
	}
}
