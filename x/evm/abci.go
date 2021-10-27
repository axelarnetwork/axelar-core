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
		handleUnsignedBatchedCommands(ctx, k.ForChain(chain.Name), signer)
	}

	return nil
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
