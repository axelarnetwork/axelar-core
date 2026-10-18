package axelarnet

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, bk types.BaseKeeper, ibcKeeper keeper.IBCKeeper, n types.Nexus, bank types.BankKeeper) ([]abci.ValidatorUpdate, error) {
	queue := bk.GetIBCTransferQueue(ctx)
	endBlockerLimit := bk.GetEndBlockerLimit(ctx)

	var failed []types.IBCTransfer
	count := uint64(0)
	for count < endBlockerLimit && !queue.IsEmpty() {
		var transfer types.IBCTransfer
		queue.Dequeue(&transfer)
		count++

		succeeded := false
		_ = utils.RunCached(ctx, bk, func(cachedCtx sdk.Context) ([]abci.ValidatorUpdate, error) {
			err := ibcKeeper.SendIBCTransfer(cachedCtx, transfer)
			if err != nil {
				bk.Logger(cachedCtx).Error(fmt.Sprintf("failed to send IBC transfer %s with id %s for %s:  %s", transfer.Token, transfer.ID.String(), transfer.Receiver, err))
				return nil, err
			}

			events.Emit(cachedCtx,
				&types.IBCTransferSent{
					ID:         transfer.ID,
					Receipient: transfer.Receiver,
					Asset:      transfer.Token,
					Sequence:   transfer.Sequence,
					PortID:     transfer.PortID,
					ChannelID:  transfer.ChannelID,
					Recipient:  transfer.Receiver,
				})

			bk.Logger(cachedCtx).Debug(fmt.Sprintf("successfully sent IBC transfer %s with id %s from %s to %s", transfer.Token, transfer.ID.String(), transfer.Sender, transfer.Receiver))
			succeeded = true
			return nil, nil
		})
		// mark the transfer as failed in the event of an error or a panic
		if !succeeded {
			failed = append(failed, transfer)
		}
	}

	// re-lock tokens to escrow and set transfer as failed
	for _, f := range failed {
		relocked := utils.RunCached(ctx, bk, func(cachedCtx sdk.Context) (bool, error) {
			lockableAsset, err := n.NewLockableAsset(cachedCtx, ibcKeeper, bank, f.Token)
			if err != nil {
				return false, err
			}

			if err := lockableAsset.LockFrom(cachedCtx, types.AxelarIBCAccount); err != nil {
				return false, err
			}

			if err := bk.SetTransferFailed(cachedCtx, f.ID); err != nil {
				return false, err
			}

			events.Emit(cachedCtx,
				&types.IBCTransferFailed{
					ID:        f.ID,
					Sequence:  f.Sequence,
					PortID:    f.PortID,
					ChannelID: f.ChannelID,
				})

			return true, nil
		})

		if !relocked {
			bk.Logger(ctx).Error(fmt.Sprintf("failed to re-lock tokens for failed IBC transfer %s with id %s to %s", f.Token, f.ID.String(), f.Receiver))
		}
	}

	return nil, nil
}
