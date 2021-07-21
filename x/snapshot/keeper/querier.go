package keeper

import (
	"bytes"
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
	QProxy    = "proxy"
	QOperator = "operator"
	QInfo     = "info"
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

	bz, err := toJSON(reply)
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

	validators := make([]validator, len(snapshot.Validators))

	for i, val := range snapshot.Validators {
		validators[i].ShareCount = val.ShareCount
		validators[i].Validator = val.GetSDKValidator().GetOperator().String()
	}

	distPolicyStr := strings.ToLower(strings.TrimPrefix(
		snapshot.KeyShareDistributionPolicy.String(), "KEY_SHARE_DISTRIBUTION_POLICY_"))

	reply := struct {
		Validators []validator `json:"validators"`

		Timestamp                  string `json:"timestamp"`
		KeyShareDistributionPolicy string `json:"key_share_distribution_policy"`

		Height          int64 `json:"height"`
		TotalShareCount int64 `json:"total_share_count"`
		Counter         int64 `json:"counter"`
	}{
		Validators: validators,

		Timestamp:                  snapshot.Timestamp.String(),
		KeyShareDistributionPolicy: distPolicyStr,

		Height:          snapshot.Height,
		TotalShareCount: snapshot.TotalShareCount.Int64(),
		Counter:         snapshot.Counter,
	}

	bz, err := toJSON(reply)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, err.Error())
	}

	return bz, nil
}

func toJSON(v interface{}) ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(buff)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

type validator struct {
	Validator  string `json:"validator"`
	ShareCount int64  `json:"share_count"`
}
