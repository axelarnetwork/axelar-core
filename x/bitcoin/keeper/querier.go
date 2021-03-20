package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// Query paths
const (
	QueryDepositAddress = "depositAddr"
	GetTx               = "getTransferTx"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k types.BTCKeeper, s types.Signer, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QueryDepositAddress:
			res, err = queryDepositAddress(ctx, k, s, n, req.Data)
		case GetTx:
			res, err = getConsolidationTx(ctx, k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

func queryDepositAddress(ctx sdk.Context, k types.BTCKeeper, s types.Signer, n types.Nexus, data []byte) ([]byte, error) {
	var params types.DepositQueryParams
	if err := types.ModuleCdc.UnmarshalJSON(data, &params); err != nil {
		return nil, fmt.Errorf("could not parse the recipient")
	}

	chain, ok := n.GetChain(ctx, params.Chain)
	if !ok {
		return nil, fmt.Errorf("recipient chain not found")
	}

	recipient := nexus.CrossChainAddress{Chain: chain, Address: params.Address}

	pk, ok := s.GetCurrentMasterKey(ctx, exported.Bitcoin)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	addr := types.NewLinkedAddress(pk, k.GetNetwork(ctx), recipient)

	return []byte(addr.EncodeAddress()), nil
}

func getConsolidationTx(ctx sdk.Context, k types.BTCKeeper) ([]byte, error) {
	tx, ok := k.GetSignedTx(ctx)
	if !ok {
		return nil, fmt.Errorf("no signed consolidation transaction ready")
	}
	return k.Codec().MustMarshalJSON(tx), nil
}
