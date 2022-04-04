package keeper

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc queries for the tss module
type Querier struct {
	keeper types.TSSKeeper
	nexus  types.Nexus
	staker types.StakingKeeper
}

// NewGRPCQuerier creates a new tss Querier
func NewGRPCQuerier(k types.TSSKeeper, n types.Nexus, s types.StakingKeeper) Querier {
	return Querier{
		keeper: k,
		nexus:  n,
		staker: s,
	}
}

// NextKeyID returns the key ID assigned for the next rotation on a given chain and for the given key role
func (q Querier) NextKeyID(c context.Context, req *types.NextKeyIDRequest) (*types.NextKeyIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrTss, fmt.Sprintf("chain [%s] not found", req.Chain)).Error())
	}

	keyID, ok := q.keeper.GetNextKeyID(ctx, chain, req.KeyRole)
	if !ok {
		return nil, status.Error(codes.OK, fmt.Errorf("no next key assigned for key role [%s] on chain [%s]", req.KeyRole.SimpleString(), chain.Name).Error())
	}

	return &types.NextKeyIDResponse{KeyID: keyID}, nil
}

// AssignableKey returns true if there is assign
func (q Querier) AssignableKey(c context.Context, req *types.AssignableKeyRequest) (*types.AssignableKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrTss, fmt.Sprintf("chain [%s] not found", req.Chain)).Error())
	}

	_, assigned := q.keeper.GetNextKeyID(ctx, chain, req.KeyRole)

	return &types.AssignableKeyResponse{Assignable: !assigned}, nil
}

// ValidatorKey returns a map of active multisig role key ids to pub keys of a validator
func (q Querier) ValidatorKey(c context.Context, req *types.ValidatorKeyRequest) (*types.ValidatorKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	valAddress, err := sdk.ValAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	validator := q.staker.Validator(ctx, valAddress)
	if validator == nil {
		return nil, fmt.Errorf("not a validator")
	}

	chains := q.nexus.GetChains(ctx)
	keys := make(map[string]*types.ValidatorKeyResponse_Keys, 10)

	keyRoles := []exported.KeyRole{exported.MasterKey, exported.SecondaryKey}

	for _, chain := range chains {
		for _, keyRole := range keyRoles {
			if currentKey, found := q.keeper.GetCurrentKey(ctx, chain, keyRole); found {
				if valKeys, ok := q.getSerializedMultisigKeys(ctx, currentKey, valAddress); ok {
					keys[string(currentKey.ID)] = valKeys
				}
			}

			if nextKey, found := q.keeper.GetNextKey(ctx, chain, keyRole); found {
				if valKeys, ok := q.getSerializedMultisigKeys(ctx, nextKey, valAddress); ok {
					keys[string(nextKey.ID)] = valKeys
				}
			}

			oldActiveKeys, err := q.keeper.GetOldActiveKeys(ctx, chain, keyRole)
			if err != nil {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
			}

			for _, oldActiveKey := range oldActiveKeys {
				if valKeys, ok := q.getSerializedMultisigKeys(ctx, oldActiveKey, valAddress); ok {
					keys[string(oldActiveKey.ID)] = valKeys
				}
			}
		}
	}

	return &types.ValidatorKeyResponse{Keys: keys}, nil
}

func (q Querier) getSerializedMultisigKeys(ctx sdk.Context, key exported.Key, valAddress sdk.ValAddress) (*types.ValidatorKeyResponse_Keys, bool) {
	if key.Type != exported.Multisig {
		return nil, false
	}

	if pubkeys, ok := q.keeper.GetMultisigPubKeysByValidator(ctx, key.ID, valAddress); ok {
		valKeys := make([][]byte, len(pubkeys))

		for i, pubkey := range pubkeys {
			wrappedKey := btcec.PublicKey(pubkey)
			valKeys[i] = (&wrappedKey).SerializeCompressed()
		}

		return &types.ValidatorKeyResponse_Keys{Keys: valKeys}, true
	}

	return nil, false
}
