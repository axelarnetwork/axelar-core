package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

//Query labels
const (
	QProxy      = "proxy"
	QOperator   = "operator"
	QInfo       = "info"
	QValidators = "validators"
)

// NewQuerier returns a new querier for the evm module
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QProxy:
			return queryProxy(ctx, k, path[1])
		case QOperator:
			return queryOperator(ctx, k, path[1])
		case QInfo:
			return querySnapshot(ctx, k, path[1])
		case QValidators:
			return QueryValidators(ctx, k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown snapshot query endpoint: %s", path[0]))
		}
	}
}

func queryProxy(ctx sdk.Context, k Keeper, address string) ([]byte, error) {
	addr, err := sdk.ValAddressFromBech32(address)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, "address invalid")
	}

	proxy, active := k.GetProxy(ctx, addr)
	if proxy == nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, "no proxy set for operator address")
	}

	statusStr := "inactive"
	if active {
		statusStr = "active"
	}

	reply := struct {
		Address string `json:"address"`
		Status  string `json:"status"`
	}{
		Address: proxy.String(),
		Status:  statusStr,
	}

	bz, err := json.Marshal(reply)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, err.Error())
	}

	return bz, nil
}

func queryOperator(ctx sdk.Context, k Keeper, proxy string) ([]byte, error) {
	proxyAddr, err := sdk.AccAddressFromBech32(proxy)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, "invalid proxy address")
	}

	operator := k.GetOperator(ctx, proxyAddr)
	if operator == nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, "no operator associated to the proxy address")
	}

	return []byte(operator.String()), nil
}

func querySnapshot(ctx sdk.Context, k Keeper, counter string) ([]byte, error) {
	var found bool
	var snapshot exported.Snapshot

	if strings.ToLower(counter) == "latest" {
		snapshot, found = k.GetLatestSnapshot(ctx)
	} else {
		c, err := strconv.ParseInt(counter, 10, 64)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrSnapshot, err.Error())
		}

		snapshot, found = k.GetSnapshot(ctx, c)
	}

	if !found {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, "no snapshot found")
	}

	bz, err := snapshot.GetSuccinctJSON()
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, err.Error())
	}

	return bz, nil
}

// QueryValidators returns validators' tss information
func QueryValidators(ctx sdk.Context, k Keeper) ([]byte, error) {
	var validators []*types.QueryValidatorsResponse_Validator

	validatorIter := func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {
		v, ok := validator.(stakingtypes.Validator)
		if !ok {
			return false
		}

		illegibility, err := k.GetValidatorIllegibility(ctx, &v)
		if err != nil {
			return false
		}

		validators = append(validators, &types.QueryValidatorsResponse_Validator{
			OperatorAddress: v.OperatorAddress,
			Moniker:         v.GetMoniker(),
			TssIllegibilityInfo: types.QueryValidatorsResponse_TssIllegibilityInfo{
				Tombstoned:            illegibility.Is(exported.Tombstoned),
				Jailed:                illegibility.Is(exported.Jailed),
				MissedTooManyBlocks:   illegibility.Is(exported.MissedTooManyBlocks),
				NoProxyRegistered:     illegibility.Is(exported.NoProxyRegistered),
				TssSuspended:          illegibility.Is(exported.TssSuspended),
				ProxyInsuficientFunds: illegibility.Is(exported.ProxyInsuficientFunds),
				StaleTssHeartbeat:     !k.tss.IsOperatorAvailable(ctx, v.GetOperator()),
			},
		})

		return false
	}

	k.staking.IterateBondedValidatorsByPower(ctx, validatorIter)
	resp := types.QueryValidatorsResponse{Validators: validators}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}
