package keeper

import (
	"encoding/hex"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewTssHandler returns the handler for processing signatures delivered by the tss module
func NewTssHandler(keeper types.BaseKeeper, nexus types.Nexus, signer types.Signer) tss.Handler {
	return func(ctx sdk.Context, info tss.SignInfo) error {
		chains := nexus.GetChains(ctx)

		for _, chain := range chains {
			handleUnsignedBatchedCommands(ctx, keeper.ForChain(chain.Name), signer)
		}

		return nil
	}
}

func handleUnsignedBatchedCommands(ctx sdk.Context, keeper types.ChainKeeper, signer types.Signer) {
	if _, ok := keeper.GetNetwork(ctx); !ok {
		return
	}

	batchedCommands := keeper.GetLatestCommandBatch(ctx)
	if !batchedCommands.Is(types.BatchSigning) {
		return
	}

	batchedCommandsIDHex := hex.EncodeToString(batchedCommands.GetID())
	_, status := signer.GetSig(ctx, batchedCommandsIDHex)
	switch status {
	case tss.SigStatus_Signed:
		batchedCommands.SetStatus(types.BatchSigned)
	case tss.SigStatus_Signing:
		return
	default:
		batchedCommands.SetStatus(types.BatchAborted)
		return
	}
}
