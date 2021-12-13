package keeper

import (
	"encoding/hex"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewTssHandler returns the handler for processing signatures delivered by the tss module
func NewTssHandler(keeper types.BaseKeeper, nexus types.Nexus, signer types.Signer) tss.Handler {
	return func(ctx sdk.Context, _ tss.SignInfo) error {
		chains := nexus.GetChains(ctx)

		for _, chain := range chains {
			handleUnsignedBatchedCommands(ctx, keeper.ForChain(chain.Name), signer)
		}

		return nil
	}
}

func handleUnsignedBatchedCommands(ctx sdk.Context, keeper types.ChainKeeper, signer types.Signer) (handled bool) {
	if _, ok := keeper.GetNetwork(ctx); !ok {
		return false
	}

	commandBatch := keeper.GetLatestCommandBatch(ctx)
	if !commandBatch.Is(types.BatchSigning) {
		return false
	}

	_, sigStatus := signer.GetSig(ctx, hex.EncodeToString(commandBatch.GetID()))
	switch sigStatus {
	case tss.SigStatus_Signed:
		commandBatch.SetStatus(types.BatchSigned)
		keeper.DeleteUnsignedCommandBatchID(ctx)
		keeper.SetLatestSignedCommandBatchID(ctx, commandBatch.GetID())

		return true
	case tss.SigStatus_Aborted:
		fallthrough
	case tss.SigStatus_Invalid:
		commandBatch.SetStatus(types.BatchAborted)

		return true
	}

	return false
}
