package keeper

import (
	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

// PrepareProposalHandler puts the pause skip marker first in the block if this validator skips the pause
func (k Keeper) PrepareProposalHandler(next sdk.PrepareProposalHandler) sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		if !k.SkipsPause(ctx) {
			return next(ctx, req)
		}

		marker := types.PauseSkipMarker()

		withoutMarker := *req
		withoutMarker.MaxTxBytes -= cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{marker})

		res, err := next(ctx, &withoutMarker)
		if err != nil {
			return nil, err
		}

		res.Txs = append([][]byte{marker}, res.Txs...)
		return res, nil
	}
}

// ProcessProposalHandler rejects blocks with the pause skip marker unless this validator skips the pause
func (k Keeper) ProcessProposalHandler(next sdk.ProcessProposalHandler) sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		if len(req.Txs) == 0 || !types.IsPauseSkipMarker(req.Txs[0]) {
			return next(ctx, req)
		}

		if !k.SkipsPause(ctx) {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		withoutMarker := *req
		withoutMarker.Txs = req.Txs[1:]

		return next(ctx, &withoutMarker)
	}
}
