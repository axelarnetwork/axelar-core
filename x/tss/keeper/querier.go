package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	ethereum "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	QueryMasterKey = "get-masterkey"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryMasterKey:
			return queryMasterAddress(ctx, path[1:], k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown tss query endpoint: %s", path[0]))
		}
	}
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
