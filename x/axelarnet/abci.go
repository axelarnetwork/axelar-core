package axelarnet

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/utils/funcs"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, bk types.BaseKeeper, ibcKeeper keeper.IBCKeeper) ([]abci.ValidatorUpdate, error) {
	queue := bk.GetIBCTransferQueue(ctx)

	var failed []types.IBCTransfer
	for !queue.IsEmpty() {
		var transfer types.IBCTransfer
		queue.Dequeue(&transfer)

		succeeded := false
		_ = utils.RunCached(ctx, bk, func(cachedCtx sdk.Context) ([]abci.ValidatorUpdate, error) {
			err := ibcKeeper.SendIBCTransfer(cachedCtx, transfer)
			if err != nil {
				bk.Logger(cachedCtx).Error(fmt.Sprintf("failed to send IBC transfer %s with id %s for %s:  %s", transfer.Token, transfer.ID.String(), transfer.Receiver, err))
				return nil, err
			}

			funcs.MustNoErr(cachedCtx.EventManager().EmitTypedEvent(
				&types.IBCTransferSent{
					ID:         transfer.ID,
					Receipient: transfer.Receiver,
					Asset:      transfer.Token,
					Sequence:   transfer.Sequence,
					PortID:     transfer.PortID,
					ChannelID:  transfer.ChannelID,
				}))

			bk.Logger(cachedCtx).Debug(fmt.Sprintf("successfully sent IBC transfer %s with id %s from %s to %s", transfer.Token, transfer.ID.String(), transfer.Sender, transfer.Receiver))
			succeeded = true
			return nil, nil
		})
		// mark the transfer as failed in the event of an error or a panic
		if !succeeded {
			failed = append(failed, transfer)
		}
	}

	// set transfer as failed
	for _, f := range failed {
		funcs.MustNoErr(bk.SetTransferFailed(ctx, f.ID))

		funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(
			&types.IBCTransferFailed{
				ID:        f.ID,
				Sequence:  f.Sequence,
				PortID:    f.PortID,
				ChannelID: f.ChannelID,
			}))
	}

	return nil, nil
}
