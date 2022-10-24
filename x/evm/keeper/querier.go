package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// Query labels
const (
	QTokenAddressBySymbol = "token-address-symbol"
	QTokenAddressByAsset  = "token-address-asset"
	QPendingCommands      = "pending-commands"
	QCommand              = "command"
)

// Bytecode labels
const (
	BCToken  = "token"
	BCBurner = "burner"
)

// Token address labels
const (
	BySymbol  = "symbol"
	ByAsset   = "asset"
	ByAddress = "address"
)

// NewQuerier returns a new querier for the evm module
func NewQuerier(k types.BaseKeeper, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		if len(path) <= 1 {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "missing evm chain name")
		}
		chainKeeper, err := k.ForChain(ctx, exported.ChainName(path[1]))
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}

		switch path[0] {
		case QTokenAddressByAsset:
			return QueryTokenAddressByAsset(ctx, chainKeeper, n, path[2])
		case QTokenAddressBySymbol:
			return QueryTokenAddressBySymbol(ctx, chainKeeper, n, path[2])
		case QCommand:
			return queryCommand(ctx, chainKeeper, n, path[2])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown evm-bridge query endpoint: %s", path[0]))
		}
	}
}

func queryCommand(ctx sdk.Context, keeper types.ChainKeeper, n types.Nexus, id string) ([]byte, error) {
	cmdID, err := types.HexToCommandID(id)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	cmd, ok := keeper.GetCommand(ctx, cmdID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find command '%s'", cmd.ID.Hex()))
	}

	resp, err := GetCommandResponse(cmd)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	return resp.Marshal()
}

// GetCommandResponse converts a Command into a CommandResponse type
func GetCommandResponse(cmd types.Command) (types.QueryCommandResponse, error) {
	params, err := cmd.DecodeParams()
	if err != nil {
		return types.QueryCommandResponse{}, err
	}

	return types.QueryCommandResponse{
		ID:         cmd.ID.Hex(),
		Type:       cmd.Type.String(),
		KeyID:      string(cmd.KeyID),
		MaxGasCost: cmd.MaxGasCost,
		Params:     params,
	}, nil
}

// QueryTokenAddressByAsset returns the address of the token contract by asset
// Deprecated: Use token-info query instead
func QueryTokenAddressByAsset(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, asset string) ([]byte, error) {
	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	token := k.GetERC20TokenByAsset(ctx, asset)
	if token.Is(types.NonExistent) {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("token for asset '%s' non-existent", asset))
	}

	resp := types.QueryTokenAddressResponse{
		Address:   token.GetAddress().Hex(),
		Confirmed: token.Is(types.Confirmed),
	}
	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

// QueryTokenAddressBySymbol returns the address of the token contract by symbol
// Deprecated: Use token-info query instead
func QueryTokenAddressBySymbol(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, symbol string) ([]byte, error) {
	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	token := k.GetERC20TokenBySymbol(ctx, symbol)
	if token.Is(types.NonExistent) {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("token for symbol '%s' non-existent", symbol))
	}

	resp := types.QueryTokenAddressResponse{
		Address:   token.GetAddress().Hex(),
		Confirmed: token.Is(types.Confirmed),
	}
	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}
