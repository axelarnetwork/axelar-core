package keeper

import (
	"fmt"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// Query paths
const (
	QuerySigStatus = "sig-status"
	QueryKeyStatus = "key-status"
)

// NewQuerier returns a new querier for the TSS module
func NewQuerier(k tssTypes.TSSKeeper, v tssTypes.Voter) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QuerySigStatus:
			res, err = querySigStatus(ctx, k, v, path[1])
		case QueryKeyStatus:
			res, err = queryKeygenStatus(ctx, k, v, path[1])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown tss query endpoint: %s", path[0]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

func querySigStatus(ctx sdk.Context, k tssTypes.TSSKeeper, v tssTypes.Voter, sigID string) ([]byte, error) {
	var resp tssTypes.QuerySigResponse
	if sig, ok := k.GetSig(ctx, sigID); ok {
		// poll was successful
		resp := tssTypes.QuerySigResponse{
			VoteStatus: tssTypes.Decided,
			Signature: &tssTypes.Signature{
				R: sig.R.Bytes(),
				S: sig.S.Bytes(),
			},
		}
		return resp.Marshal()
	}

	pollMeta := voting.NewPollKey(tssTypes.ModuleName, sigID)
	poll := v.GetPoll(ctx, pollMeta)

	if poll == nil {
		// poll either never existed or has been closed
		resp.VoteStatus = tssTypes.Unspecified
	} else {
		// poll still open, pending a decision
		resp.VoteStatus = tssTypes.Pending
	}

	return resp.Marshal()
}

func queryKeygenStatus(ctx sdk.Context, k tssTypes.TSSKeeper, v tssTypes.Voter, keyID string) ([]byte, error) {
	var resp tssTypes.QueryKeyResponse

	if key, ok := k.GetKey(ctx, keyID); ok {
		// poll was successful
		resp = tssTypes.QueryKeyResponse{
			VoteStatus: tssTypes.Decided,
			Role:       key.Role,
		}

		return resp.Marshal()
	}

	pollMeta := voting.NewPollKey(tssTypes.ModuleName, keyID)
	poll := v.GetPoll(ctx, pollMeta)
	if poll == nil {
		// poll either never existed or has been closed
		resp.VoteStatus = tssTypes.Unspecified
	} else {
		// poll still open, pending a decision
		resp.VoteStatus = tssTypes.Pending
	}

	return resp.Marshal()
}
