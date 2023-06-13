package keeper

import (
	"context"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
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

	if _, ok := q.keeper.GetKeygenSession(ctx, req.KeyID); ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrMultisig, fmt.Sprintf("keygen in progress for key id [%s]", req.KeyID)).Error())
	}

	key, ok := q.keeper.GetKey(ctx, req.KeyID)
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrMultisig, fmt.Sprintf("key not found for key id [%s]", req.KeyID)).Error())
	}

	participants := slices.Map(key.GetParticipants(), func(p sdk.ValAddress) types.KeygenParticipant {
		return types.KeygenParticipant{
			Address: p.String(),
			Weight:  key.GetWeight(p),
			PubKey:  funcs.MustOk(key.GetPubKey(p)).String(),
		}
	})
	sort.SliceStable(participants, func(i, j int) bool {
		return participants[i].Weight.GT(participants[j].Weight)
	})

	return &types.KeyResponse{
		KeyID:              req.KeyID,
		State:              key.GetState(),
		StartedAt:          key.GetHeight(),
		StartedAtTimestamp: key.GetTimestamp(),
		ThresholdWeight:    key.GetMinPassingWeight(),
		BondedWeight:       key.GetBondedWeight(),
		Participants:       participants,
	}, nil
}

func getKeygenParticipants(key exported.Key) []types.KeygenParticipant {
	snapshot := key.GetSnapshot()
	participants := slices.Map(snapshot.GetParticipantAddresses(), func(p sdk.ValAddress) types.KeygenParticipant {
		var pubKey string
		if pub, ok := key.GetPubKey(p); ok {
			pubKey = pub.String()
		}

		return types.KeygenParticipant{
			Address: p.String(),
			Weight:  snapshot.GetParticipantWeight(p),
			PubKey:  pubKey,
		}
	})
	sort.SliceStable(participants, func(i, j int) bool {
		return participants[i].Weight.GT(participants[j].Weight)
	})

	return participants
}

// KeygenSession returns the keygen session info for the given key ID
func (q Querier) KeygenSession(c context.Context, req *types.KeygenSessionRequest) (*types.KeygenSessionResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	session, ok := q.keeper.GetKeygenSession(ctx, req.KeyID)

	var key exported.Key
	var state exported.MultisigState
	if !ok {
		key, ok = q.keeper.GetKey(ctx, req.KeyID)
		if !ok {
			return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrMultisig, fmt.Sprintf("key not found for key id [%s]", req.KeyID)).Error())
		}

		session = types.KeygenSession{KeygenThreshold: utils.ZeroThreshold}
		state = exported.Completed
	} else {
		key = &session.Key
		state = session.GetState()
	}

	return &types.KeygenSessionResponse{
		StartedAt:              key.GetHeight(),
		StartedAtTimestamp:     key.GetTimestamp(),
		ExpiresAt:              session.GetExpiresAt(),
		CompletedAt:            session.GetCompletedAt(),
		GracePeriod:            session.GetGracePeriod(),
		State:                  state,
		KeygenThresholdWeight:  key.GetSnapshot().CalculateMinPassingWeight(session.KeygenThreshold),
		SigningThresholdWeight: key.GetMinPassingWeight(),
		BondedWeight:           key.GetBondedWeight(),
		Participants:           getKeygenParticipants(key),
	}, nil
}

// Params returns the reward module params
func (q Querier) Params(c context.Context, req *types.ParamsRequest) (*types.ParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	params := q.keeper.GetParams(ctx)

	return &types.ParamsResponse{
		Params: params,
	}, nil
}
