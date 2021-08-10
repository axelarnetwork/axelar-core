package evm

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ types.BaseKeeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, k types.BaseKeeper, nexus types.Nexus, signer types.Signer) []abci.ValidatorUpdate {
	chains := nexus.GetChains(ctx)

	for _, chain := range chains {
		handleUnsignedBatchedCommands(ctx, k.ForChain(ctx, chain.Name), signer)
	}

	return nil
}

func handleUnsignedBatchedCommands(ctx sdk.Context, keeper types.ChainKeeper, signer types.Signer) {
	if _, ok := keeper.GetNetwork(ctx); !ok {
		return
	}

	batchedCommands, ok := keeper.GetUnsignedBatchedCommands(ctx)
	if !ok || !batchedCommands.Is(types.BatchedCommands_Signing) {
		return
	}

	batchedCommandsIDHex := hex.EncodeToString(batchedCommands.ID)
	_, status := signer.GetSig(ctx, batchedCommandsIDHex)
	switch status {
	case tss.SigStatus_Signed:
		keeper.DeleteUnsignedBatchedCommands(ctx)
		keeper.SetSignedBatchedCommands(ctx, batchedCommands)
		keeper.SetLatestSignedBatchedCommandsID(ctx, batchedCommands.ID)
	case tss.SigStatus_Scheduled, tss.SigStatus_Signing:
		return
	default:
		batchedCommands.Status = types.BatchedCommands_Aborted
		keeper.SetUnsignedBatchedCommands(ctx, batchedCommands)
		return
	}
}
