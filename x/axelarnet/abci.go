package axelarnet

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, bk types.BaseKeeper, t types.IBCTransferKeeper, c types.ChannelKeeper) ([]abci.ValidatorUpdate, error) {
	queue := bk.GetIBCTransferQueue(ctx)

	var failed []types.IBCTransfer
	for !queue.IsEmpty() {
		var transfer types.IBCTransfer
		queue.Dequeue(&transfer)

		cacheCtx, writeCache := ctx.CacheContext()
		err := sendIBCTransfer(cacheCtx, bk, t, c, transfer)
		if err != nil {
			bk.Logger(cacheCtx).Error(fmt.Sprintf("failed to send IBC transfer %s for %s:  %s", transfer.Token, transfer.Receiver, err))
			failed = append(failed, transfer)
			continue
		}
		writeCache()
		ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())

		bk.Logger(ctx).Debug(fmt.Sprintf("successfully sent IBC transfer %s from %s to %s", transfer.Token, transfer.Sender, transfer.Receiver))
	}

	// re-queue
	for _, f := range failed {
		if err := bk.EnqueueTransfer(ctx, f); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// IBCTransfer inits an IBC transfer
func sendIBCTransfer(ctx sdk.Context, k types.BaseKeeper, t types.IBCTransferKeeper, c types.ChannelKeeper, transfer types.IBCTransfer) error {

	_, state, err := c.GetChannelClientState(ctx, transfer.PortID, transfer.ChannelID)
	if err != nil {
		return err
	}

	height := clienttypes.NewHeight(state.GetLatestHeight().GetRevisionNumber(), state.GetLatestHeight().GetRevisionHeight()+k.GetRouteTimeoutWindow(ctx))
	err = t.SendTransfer(ctx, transfer.PortID, transfer.ChannelID, transfer.Token, transfer.Sender, transfer.Receiver, height, 0)
	if err != nil {
		return err
	}

	return nil
}
