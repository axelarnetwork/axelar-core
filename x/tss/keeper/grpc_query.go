package keeper

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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

	chain, ok := q.nexus.GetChain(ctx, nexus.ChainName(req.Chain))
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

	chain, ok := q.nexus.GetChain(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrTss, fmt.Sprintf("chain [%s] not found", req.Chain)).Error())
	}

	_, assigned := q.keeper.GetNextKeyID(ctx, chain, req.KeyRole)

	return &types.AssignableKeyResponse{Assignable: !assigned}, nil
}

// ValidatorMultisigKeys returns a map of active multisig role key ids to pub keys of a validator
func (q Querier) ValidatorMultisigKeys(c context.Context, req *types.ValidatorMultisigKeysRequest) (*types.ValidatorMultisigKeysResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	valAddress, err := sdk.ValAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	chains := q.nexus.GetChains(ctx)
	keyRoles := []exported.KeyRole{exported.MasterKey, exported.SecondaryKey}

	var keys []exported.Key

	for _, chain := range chains {
		for _, keyRole := range keyRoles {
			if currentKey, found := q.keeper.GetCurrentKey(ctx, chain, keyRole); found {
				keys = append(keys, currentKey)
			}

			if nextKey, found := q.keeper.GetNextKey(ctx, chain, keyRole); found {
				keys = append(keys, nextKey)
			}

			oldActiveKeys, err := q.keeper.GetOldActiveKeys(ctx, chain, keyRole)
			if err != nil {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
			}

			keys = append(keys, oldActiveKeys...)
		}
	}

	res := types.ValidatorMultisigKeysResponse{Keys: make(map[string]*types.ValidatorMultisigKeysResponse_Keys, len(keys))}
	for _, key := range keys {
		valKeys, ok := q.getSerializedMultisigKeys(ctx, key, valAddress)
		if !ok {
			continue
		}

		res.Keys[string(key.ID)] = valKeys
	}

	return &res, nil
}

func (q Querier) getSerializedMultisigKeys(ctx sdk.Context, key exported.Key, valAddress sdk.ValAddress) (*types.ValidatorMultisigKeysResponse_Keys, bool) {
	if key.Type != exported.Multisig {
		return nil, false
	}

	pubkeys, ok := q.keeper.GetMultisigPubKeysByValidator(ctx, key.ID, valAddress)
	if !ok {
		return nil, false
	}

	valKeys := make([][]byte, len(pubkeys))

	for i, pubkey := range pubkeys {
		wrappedKey := btcec.PublicKey(pubkey)
		valKeys[i] = wrappedKey.SerializeCompressed()
	}

	return &types.ValidatorMultisigKeysResponse_Keys{Keys: valKeys}, true
}
