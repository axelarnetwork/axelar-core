package keeper

import (
	"context"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc queries for the multisig module
type Querier struct {
	keeper types.Keeper
	staker types.Staker
}

// NewGRPCQuerier creates a new multisig Querier
func NewGRPCQuerier(k types.Keeper, s types.Staker) Querier {
	return Querier{
		keeper: k,
		staker: s,
	}
}

// KeyID returns the key ID assigned to a given chain
func (q Querier) KeyID(c context.Context, req *types.KeyIDRequest) (*types.KeyIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	keyID, ok := q.keeper.GetCurrentKeyID(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrMultisig, fmt.Sprintf("key id not found for chain [%s]", req.Chain)).Error())
	}

	return &types.KeyIDResponse{KeyID: keyID}, nil
}

// NextKeyID returns the key ID assigned for the next rotation on a given chain and empty if none is assigned
func (q Querier) NextKeyID(c context.Context, req *types.NextKeyIDRequest) (*types.NextKeyIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	keyID, ok := q.keeper.GetNextKeyID(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrMultisig, fmt.Sprintf("next key id not found for chain [%s]", req.Chain)).Error())
	}

	return &types.NextKeyIDResponse{KeyID: keyID}, nil
}

// Key returns the key corresponding to a given key ID
func (q Querier) Key(c context.Context, req *types.KeyRequest) (*types.KeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	key, ok := q.keeper.GetKey(ctx, req.KeyID)
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrMultisig, fmt.Sprintf("key not found for key id [%s]", req.KeyID)).Error())
	}

	participants := slices.Map(key.GetParticipants(), func(p sdk.ValAddress) types.KeyResponse_Participant {
		return types.KeyResponse_Participant{
			Address: p.String(),
			Weight:  key.GetWeight(p),
			PubKey:  fmt.Sprintf("0x%s", funcs.MustOk(key.GetPubKey(p)).String()),
		}
	})
	sort.SliceStable(participants, func(i, j int) bool {
		return participants[i].Weight.GT(participants[j].Weight)
	})

	return &types.KeyResponse{
		KeyID:           req.KeyID,
		State:           key.GetState(),
		Height:          key.GetHeight(),
		Timestamp:       key.GetTimestamp(),
		ThresholdWeight: key.GetMinPassingWeight(),
		BondedWeight:    key.GetBondedWeight(),
		Participants:    participants,
	}, nil
}

// Snapshot returns the snapshot corresponding to a given key ID
func (q Querier) Snapshot(c context.Context, req *types.SnapshotRequest) (*types.SnapshotResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	key, ok := q.keeper.GetKey(ctx, req.KeyID)
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrMultisig, fmt.Sprintf("key not found for key id [%s]", req.KeyID)).Error())
	}

	snapshot := key.GetSnapshot()

	participants := slices.Map(snapshot.GetParticipantAddresses(), func(p sdk.ValAddress) types.SnapshotResponse_Participant {
		return types.SnapshotResponse_Participant{
			Address: p.String(),
			Weight:  snapshot.GetParticipantWeight(p),
		}
	})
	sort.SliceStable(participants, func(i, j int) bool {
		return participants[i].Weight.GT(participants[j].Weight)
	})

	return &types.SnapshotResponse{
		Height:          key.GetHeight(),
		Timestamp:       key.GetTimestamp(),
		ThresholdWeight: key.GetMinPassingWeight(),
		BondedWeight:    key.GetBondedWeight(),
		Participants:    participants,
	}, nil
}
