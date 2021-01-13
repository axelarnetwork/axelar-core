package keeper

import (
	"fmt"

	"github.com/axelarnetwork/tssd/convert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/tss/types"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	ethereum "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	QueryGetKey    = "get-key"
	QueryMasterKey = "get-masterkey"
)

const ()

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryGetKey:
			return queryGetKey(ctx, path[1], k)
		case QueryMasterKey:
			return queryMasterAddress(ctx, path[1:], k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown tss query endpoint: %s", path[0]))
		}
	}
}

func queryGetKey(ctx sdk.Context, keyID string, k Keeper) ([]byte, error) {
	pk, ok := k.GetKey(ctx, keyID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrTss, "key [%s] not found", keyID)
	}

	// pk is of type ecdsa.PublicKey, which is inherently un-marshalable
	// convert pk to tss-libs crypto.ECPoint, which implements json.Marshaler
	pkMarshalable := convert.PubkeyToPoint(pk)
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, pkMarshalable)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryMasterAddress(ctx sdk.Context, args []string, k Keeper) ([]byte, error) {

	var address string
	var err error

	chain := balance.ChainFromString(args[0])

	err = chain.Validate()

	if err != nil {
		return []byte(address), err
	}

	pk, ok := k.GetCurrentMasterKey(ctx, chain)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	switch chain {

	case balance.Bitcoin:

		if len(args) < 2 {

			err = fmt.Errorf("no network specified")

		} else {

			network := args[1]

			address, err = bitcoin.PubkeyToAddress(pk, network)

		}
	case balance.Ethereum:

		address, err = ethereum.PubkeyToAddress(pk)

	default:

		err = fmt.Errorf("unknown chain")
	}

	return []byte(address), err
}
