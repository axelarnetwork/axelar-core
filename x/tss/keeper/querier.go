package keeper

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
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
	QueryRecovery  = "recovery"
)

// NewQuerier returns a new querier for the TSS module
func NewQuerier(k tssTypes.TSSKeeper, v tssTypes.Voter, s tssTypes.Snapshotter) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QuerySigStatus:
			res, err = querySigStatus(ctx, k, v, path[1])
		case QueryKeyStatus:
			res, err = queryKeygenStatus(ctx, k, v, path[1])
		case QueryRecovery:
			res, err = queryRecovery(ctx, k, s, path[1])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown tss query endpoint: %s", path[0]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

func queryRecovery(ctx sdk.Context, k tssTypes.TSSKeeper, s tssTypes.Snapshotter, keyID string) ([]byte, error) {
	counter, ok := k.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("could not obtain snapshot counter for key ID %s", keyID)
	}
	snapshot, ok := s.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("could not obtain snapshot for counter %d", counter)
	}

	threshold, found := k.GetCorruptionThreshold(ctx, keyID)
	if !found {
		return nil, fmt.Errorf("keyID %s has no corruption threshold defined", keyID)
	}

	participants := make([]string, 0, len(snapshot.Validators))
	participantShareCounts := make([]uint32, 0, len(snapshot.Validators))
	for _, validator := range snapshot.Validators {
		participants = append(participants, validator.GetSDKValidator().GetOperator().String())
		participantShareCounts = append(participantShareCounts, uint32(validator.ShareCount))
	}

	infos := k.GetAllRecoveryInfos(ctx, keyID)

	resp := tssTypes.QueryRecoveryResponse{
		Threshold:          int32(threshold),
		PartyUids:          participants,
		PartyShareCounts:   participantShareCounts,
		ShareRecoveryInfos: infos,
	}

	return resp.Marshal()
}

func querySigStatus(ctx sdk.Context, k tssTypes.TSSKeeper, v tssTypes.Voter, sigID string) ([]byte, error) {
	var resp tssTypes.QuerySigResponse
	if sig, status := k.GetSig(ctx, sigID); status == exported.SigStatus_Signed {
		// poll was successful
		resp := tssTypes.QuerySigResponse{
			VoteStatus: tssTypes.VoteStatus_Decided,
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
		resp.VoteStatus = tssTypes.VoteStatus_Unspecified
	} else {
		// poll still open, pending a decision
		resp.VoteStatus = tssTypes.VoteStatus_Pending
	}

	return resp.Marshal()
}

func queryKeygenStatus(ctx sdk.Context, k tssTypes.TSSKeeper, v tssTypes.Voter, keyID string) ([]byte, error) {
	var resp tssTypes.QueryKeyResponse

	if key, ok := k.GetKey(ctx, keyID); ok {
		// poll was successful
		resp = tssTypes.QueryKeyResponse{
			VoteStatus: tssTypes.VoteStatus_Decided,
			Role:       key.Role,
		}

		return resp.Marshal()
	}

	pollMeta := voting.NewPollKey(tssTypes.ModuleName, keyID)
	poll := v.GetPoll(ctx, pollMeta)
	if poll == nil {
		// poll either never existed or has been closed
		resp.VoteStatus = tssTypes.VoteStatus_Unspecified
	} else {
		// poll still open, pending a decision
		resp.VoteStatus = tssTypes.VoteStatus_Pending
	}

	return resp.Marshal()
}
