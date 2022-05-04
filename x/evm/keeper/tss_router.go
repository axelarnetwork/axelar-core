package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewTssHandler returns the handler for processing signatures delivered by the tss module
func NewTssHandler(cdc codec.Codec, keeper types.BaseKeeper, signer types.Signer) tss.Handler {
	return func(ctx sdk.Context, sigInfo tss.SignInfo) error {
		var sigMetadata types.SigMetadata
		if err := cdc.Unmarshal(sigInfo.GetModuleMetadata().Value, &sigMetadata); err != nil {
			return err
		}

		if !keeper.HasChain(ctx, sigMetadata.Chain) {
			return fmt.Errorf("chain %s does not exist as an EVM chain", sigMetadata.Chain)
		}

		ck := keeper.ForChain(sigMetadata.Chain)
		commandBatch := ck.GetLatestCommandBatch(ctx)
		if !commandBatch.Is(types.BatchSigning) {
			return fmt.Errorf("the latest command batch of chain %s is not being signed", sigMetadata.Chain)
		}

		_, sigStatus := signer.GetSig(ctx, sigInfo.SigID)
		switch sigStatus {
		case tss.SigStatus_Signed:
			commandBatch.SetStatus(types.BatchSigned)
			ck.DeleteUnsignedCommandBatchID(ctx)
			ck.SetLatestSignedCommandBatchID(ctx, commandBatch.GetID())
		case tss.SigStatus_Aborted, tss.SigStatus_Invalid:
			commandBatch.SetStatus(types.BatchAborted)
		default:
			return fmt.Errorf("cannot handle signature %s with status %s for chain %s", sigInfo.SigID, sigStatus.String(), sigMetadata.Chain)
		}

		return nil
	}
}
