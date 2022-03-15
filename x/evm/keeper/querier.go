package keeper

import (
	"fmt"
	"strconv"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Query labels
const (
	QTokenAddressBySymbol = "token-address-symbol"
	QTokenAddressByAsset  = "token-address-asset"
	QAddressByKeyRole     = "address-by-key-role"
	QAddressByKeyID       = "address-by-key-id"
	QPendingCommands      = "pending-commands"
	QCommand              = "command"
)

//Bytecode labels
const (
	BCToken  = "token"
	BCBurner = "burner"
)

//Token address labels
const (
	BySymbol = "symbol"
	ByAsset  = "asset"
)

// NewQuerier returns a new querier for the evm module
func NewQuerier(k types.BaseKeeper, s types.Signer, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var chainKeeper types.ChainKeeper
		if len(path) > 1 {
			chainKeeper = k.ForChain(path[1])
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

	resp, err := GetCommandResponse(ctx, keeper.GetName(), n, cmd)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	return resp.Marshal()
}

// GetCommandResponse converts a Command into a CommandResponse type
func GetCommandResponse(ctx sdk.Context, chainName string, n types.Nexus, cmd types.Command) (types.QueryCommandResponse, error) {
	params := make(map[string]string)

	switch cmd.Command {
	case types.AxelarGatewayCommandDeployToken:
		name, symbol, decs, cap, err := types.DecodeDeployTokenParams(cmd.Params)
		if err != nil {
			return types.QueryCommandResponse{}, err
		}

		params["name"] = name
		params["symbol"] = symbol
		params["decimals"] = strconv.FormatUint(uint64(decs), 10)
		params["cap"] = cap.String()

	case types.AxelarGatewayCommandMintToken:
		symbol, addr, amount, err := types.DecodeMintTokenParams(cmd.Params)
		if err != nil {
			return types.QueryCommandResponse{}, err
		}

		params["symbol"] = symbol
		params["account"] = addr.Hex()
		params["amount"] = amount.String()

	case types.AxelarGatewayCommandBurnToken:
		symbol, salt, err := types.DecodeBurnTokenParams(cmd.Params)
		if err != nil {
			return types.QueryCommandResponse{}, err
		}

		params["symbol"] = symbol
		params["salt"] = salt.Hex()

	case types.AxelarGatewayCommandTransferOwnership, types.AxelarGatewayCommandTransferOperatorship:
		chain, ok := n.GetChain(ctx, chainName)
		if !ok {
			return types.QueryCommandResponse{}, fmt.Errorf("unknown chain '%s'", chainName)
		}

		switch chain.KeyType {
		case tss.Threshold:
			address, err := types.DecodeTransferSinglesigParams(cmd.Params)
			if err != nil {
				return types.QueryCommandResponse{}, err
			}

			param := "newOwner"
			if cmd.Command == types.AxelarGatewayCommandTransferOperatorship {
				param = "newOperator"
			}
			params[param] = address.Hex()

		case tss.Multisig:
			addresses, threshold, err := types.DecodeTransferMultisigParams(cmd.Params)
			if err != nil {
				return types.QueryCommandResponse{}, err
			}

			var hexs []string
			for _, address := range addresses {
				hexs = append(hexs, address.Hex())
			}

			param := "newOwners"
			if cmd.Command == types.AxelarGatewayCommandTransferOperatorship {
				param = "newOperators"
			}
			params[param] = strings.Join(hexs, ";")
			params["newThreshold"] = strconv.FormatUint(uint64(threshold), 10)

		default:
			return types.QueryCommandResponse{}, fmt.Errorf("unsupported key type '%s'", chain.KeyType.SimpleString())
		}

	default:
		return types.QueryCommandResponse{}, fmt.Errorf("unknown command type '%s'", cmd.Command)
	}

	return types.QueryCommandResponse{
		ID:         cmd.ID.Hex(),
		Type:       cmd.Command,
		KeyID:      string(cmd.KeyID),
		MaxGasCost: cmd.MaxGasCost,
		Params:     params,
	}, nil
}

// QueryTokenAddressByAsset returns the address of the token contract by asset
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
