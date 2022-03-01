package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewTssHandler returns the handler for processing signatures delivered by the tss module
func NewTssHandler(keeper types.BaseKeeper, nexus types.Nexus, signer types.Signer) tss.Handler {
	return func(ctx sdk.Context, sigInfo tss.SignInfo) error {
		chains := nexus.GetChains(ctx)

		ok := false
		for _, chain := range chains {
			if ok = handleUnsignedBatchedCommands(ctx, keeper.ForChain(chain.Name), signer); ok {
				break
			}
		}

		if !ok {
			return fmt.Errorf("no command batch found to handle for signature %s", sigInfo.GetSigID())
		}

		return nil
	}
}

func handleUnsignedBatchedCommands(ctx sdk.Context, keeper types.ChainKeeper, signer types.Signer) bool {
	if _, ok := keeper.GetNetwork(ctx); !ok {
		return false
	}

	commandBatch := keeper.GetLatestCommandBatch(ctx)
	if ctx.BlockHeight() >= 690489 {
		if !(commandBatch.Is(types.BatchSigning) || commandBatch.Is(types.BatchAborted)) {
			return false
		}
	} else {
		if !commandBatch.Is(types.BatchSigning) {
			return false
		}
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
