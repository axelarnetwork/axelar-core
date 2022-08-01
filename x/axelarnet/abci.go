package axelarnet

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/utils"
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

		_ = utils.RunCached(ctx, bk, func(cachedCtx sdk.Context) ([]abci.ValidatorUpdate, error) {
			err := sendIBCTransfer(ctx, bk, t, c, transfer)
			if err != nil {
				bk.Logger(ctx).Error(fmt.Sprintf("failed to send IBC transfer %s with id %s for %s:  %s", transfer.Token, transfer.ID.String(), transfer.Receiver, err))
				failed = append(failed, transfer)
				return nil, err
			}

			bk.Logger(ctx).Debug(fmt.Sprintf("successfully sent IBC transfer %s with id %s from %s to %s", transfer.Token, transfer.ID.String(), transfer.Sender, transfer.Receiver))
			return nil, nil
		})
	}

	// park the failed transfer aside
	for _, f := range failed {
		bk.SetFailedTransfer(ctx, f)
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
