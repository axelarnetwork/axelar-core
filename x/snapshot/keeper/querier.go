package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

//Query labels
const (
	QProxy       = "proxy"
	QOperator    = "operator"
	QInfo        = "info"
	QDeactivated = "deactivated"
)

// NewQuerier returns a new querier for the evm module
func NewQuerier(k Keeper, t types.TSS) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QProxy:
			return queryProxy(ctx, k, path[1])
		case QOperator:
			return queryOperator(ctx, k, path[1])
		case QInfo:
			return querySnapshot(ctx, k, path[1])
		case QDeactivated:
			return queryDeactivatedPrinciple(ctx, k, t, path[1])
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

func queryDeactivatedPrinciple(ctx sdk.Context, k Keeper, t types.TSS, keyID string) ([]byte, error) {
	var found bool
	var snapshot exported.Snapshot

	counter, found := t.GetSnapshotCounterForKeyID(ctx, keyID)
	if !found {
		return nil, fmt.Errorf("could not obtain snapshot counter for key ID %s", keyID)
	}

	snapshot, found = k.GetSnapshot(ctx, counter)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, "no snapshot found")
	}

	var res []string
	for _, validator := range snapshot.Validators {
		_, active := k.GetProxy(ctx, validator.GetSDKValidator().GetOperator())
		if !active {
			res = append(res, validator.GetSDKValidator().GetOperator().String())
		}
	}

	resp := types.QueryDeactivatedPrincipleResponse{
		PrincipalAddresses: res,
	}

	return types.ModuleCdc.MarshalBinaryLengthPrefixed(&resp)

}
