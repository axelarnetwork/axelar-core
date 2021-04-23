package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/wire"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// Query paths
const (
	QueryDepositAddress = "depositAddr"
	QueryKeyAddress     = "keyAddr"
	GetTx               = "getTx"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k types.BTCKeeper, s types.Signer, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QueryDepositAddress:
			res, err = queryDepositAddress(ctx, k, s, n, req.Data)
		case QueryKeyAddress:
			res, err = queryKeyAddress(ctx, k, s, n, req.Data)
		case GetTx:
			res, err = getRawConsolidationTx(ctx, k)
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

	masterKey, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("master key not set")
	}

	secondaryKey, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("secondary key not set")
	}

	addr := types.NewLinkedAddress(masterKey, secondaryKey, k.GetNetwork(ctx), recipient)

	return []byte(addr.EncodeAddress()), nil
}

func queryKeyAddress(ctx sdk.Context, k types.BTCKeeper, s types.Signer, n types.Nexus, data []byte) ([]byte, error) {
	keyRole, err := tss.KeyRoleFromStr(string(data))
	if err != nil {
		return nil, err
	}

	key, ok := s.GetCurrentKey(ctx, exported.Bitcoin, keyRole)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	addr := types.NewConsolidationAddress(key, k.GetNetwork(ctx))

	if _, ok := k.GetAddress(ctx, addr.EncodeAddress()); !ok {
		return nil, fmt.Errorf("no address found for current %s key", keyRole.String())
	}

	return []byte(addr.EncodeAddress()), nil
}

func getRawConsolidationTx(ctx sdk.Context, k types.BTCKeeper) ([]byte, error) {
	tx, ok := k.GetSignedTx(ctx)
	if !ok {
		return nil, fmt.Errorf("no signed consolidation transaction ready")
	}

	var buf bytes.Buffer
	if err := tx.BtcEncode(&buf, wire.FeeFilterVersion, wire.WitnessEncoding); err != nil {
		return nil, err
	}
	return k.Codec().MustMarshalJSON(hex.EncodeToString(buf.Bytes())), nil
}
