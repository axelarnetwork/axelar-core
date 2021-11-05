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
	return func(ctx sdk.Context, info tss.SignInfo) error {
		chains := nexus.GetChains(ctx)

		for _, chain := range chains {
			handleUnsignedBatchedCommands(ctx, keeper.ForChain(chain.Name), signer, info)
		}

		return nil
	}
}

func handleUnsignedBatchedCommands(ctx sdk.Context, keeper types.ChainKeeper, signer types.Signer, info tss.SignInfo) {
	if _, ok := keeper.GetNetwork(ctx); !ok {
		return
	}

	batchedCommands := keeper.GetLatestCommandBatch(ctx)
	if !batchedCommands.Is(types.BatchSigning) {
		return
	}

	batchedCommandsIDHex := hex.EncodeToString(batchedCommands.GetID())

	//TODO: merge tss key & multisigKey, tss sig & multisig with protobuf
	var status tss.SigStatus
	switch signer.GetKeyType(ctx, info.KeyID) {
	case tss.Threshold:
		_, status = signer.GetSig(ctx, batchedCommandsIDHex)
	case tss.Multisig:
		_, status = signer.GetMultisig(ctx, batchedCommandsIDHex)
	default:
		panic(fmt.Sprintf("unknown key type set for keyID %s", info.KeyID))
	}

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
