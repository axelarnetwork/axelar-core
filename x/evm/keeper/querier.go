package keeper

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Query labels
const (
	QTokenAddressBySymbol  = "token-address-symbol"
	QTokenAddressByAsset   = "token-address-asset"
	QDepositState          = "deposit-state"
	QAddressByKeyRole      = "address-by-key-role"
	QAddressByKeyID        = "address-by-key-id"
	QNextMasterAddress     = "next-master-address"
	QAxelarGatewayAddress  = "gateway-address"
	QBytecode              = "bytecode"
	QLatestBatchedCommands = "latest-batched-commands"
	QBatchedCommands       = "batched-commands"
	QPendingCommands       = "pending-commands"
	QCommand               = "command"
	QChains                = "chains"
)

// Bytecode labels
const (
	BCGateway           = "gateway"
	BCGatewayDeployment = "gateway-deployment"
	BCToken             = "token"
	BCBurner            = "burner"
)

// Token address labels
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
		case QAddressByKeyRole:
			return QueryAddressByKeyRole(ctx, s, n, path[1], path[2])
		case QAddressByKeyID:
			keyID := tss.KeyID(path[2])

			if err := keyID.Validate(); err != nil {
				return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
			}
			return QueryAddressByKeyID(ctx, s, n, path[1], keyID)
		case QNextMasterAddress:
			return queryNextMasterAddress(ctx, s, n, path[1])
		case QAxelarGatewayAddress:
			return queryAxelarGateway(ctx, chainKeeper, n)
		case QTokenAddressByAsset:
			return QueryTokenAddressByAsset(ctx, chainKeeper, n, path[2])
		case QTokenAddressBySymbol:
			return QueryTokenAddressBySymbol(ctx, chainKeeper, n, path[2])
		case QDepositState:
			return QueryDepositState(ctx, chainKeeper, n, req.Data)
		case QBatchedCommands:
			return QueryBatchedCommands(ctx, chainKeeper, s, n, path[2])
		case QLatestBatchedCommands:
			return QueryLatestBatchedCommands(ctx, chainKeeper, s)
		case QCommand:
			return queryCommand(ctx, chainKeeper, n, path[2])
		case QBytecode:
			return queryBytecode(ctx, chainKeeper, s, n, path[2])
		case QChains:
			return queryChains(ctx, n)
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

// QueryLatestBatchedCommands returns the latest batched commands
func QueryLatestBatchedCommands(ctx sdk.Context, keeper types.ChainKeeper, s types.Signer) ([]byte, error) {

	batchedCommands := keeper.GetLatestCommandBatch(ctx)
	if batchedCommands.Is(types.BatchNonExistent) {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("cannot find the latest signed batched commands for chain %s", keeper.GetName()))
	}

	resp, err := batchedCommandsToQueryResp(ctx, batchedCommands, s)
	if err != nil {
		return nil, err
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

func batchedCommandsToQueryResp(ctx sdk.Context, batchedCommands types.CommandBatch, s types.Signer) (types.QueryBatchedCommandsResponse, error) {
	batchedCommandsIDHex := hex.EncodeToString(batchedCommands.GetID())
	prevBatchedCommandsIDHex := ""
	if batchedCommands.GetPrevBatchedCommandsID() != nil {
		prevBatchedCommandsIDHex = hex.EncodeToString(batchedCommands.GetPrevBatchedCommandsID())
	}

	var commandIDs []string
	for _, id := range batchedCommands.GetCommandIDs() {
		commandIDs = append(commandIDs, id.Hex())
	}

	var resp types.QueryBatchedCommandsResponse

	switch {
	case batchedCommands.Is(types.BatchSigned):
		sig, sigStatus := s.GetSig(ctx, batchedCommandsIDHex)
		if sigStatus != tss.SigStatus_Signed {
			return resp, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding signature for sig ID %s", batchedCommandsIDHex))
		}

		var executeData []byte
		var signatures []string
		switch signature := sig.GetSig().(type) {
		case *tss.Signature_SingleSig_:
			batchedCommandsSig, err := getBatchedCommandsSig(signature.SingleSig.SigKeyPair, batchedCommands.GetSigHash())
			if err != nil {
				return resp, err
			}

			executeData, err = types.CreateExecuteDataSinglesig(batchedCommands.GetData(), batchedCommandsSig)
			if err != nil {
				return resp, sdkerrors.Wrapf(types.ErrEVM, "could not create transaction data: %s", err)
			}

			signatures = append(signatures, hex.EncodeToString(batchedCommandsSig[:]))
		case *tss.Signature_MultiSig_:
			var batchedCmdSigs []types.Signature
			var err error

			sigKeyPairs := types.SigKeyPairs(signature.MultiSig.SigKeyPairs)
			sort.Stable(sigKeyPairs)

			for _, pair := range sigKeyPairs {
				batchedCommandsSig, err := getBatchedCommandsSig(pair, batchedCommands.GetSigHash())
				if err != nil {
					return resp, err
				}

				batchedCmdSigs = append(batchedCmdSigs, batchedCommandsSig)
				signatures = append(signatures, hex.EncodeToString(batchedCommandsSig[:]))
			}

			executeData, err = types.CreateExecuteDataMultisig(batchedCommands.GetData(), batchedCmdSigs...)
			if err != nil {
				return resp, sdkerrors.Wrapf(types.ErrEVM, "could not create transaction data: %s", err)
			}
		}

		resp = types.QueryBatchedCommandsResponse{
			ID:                    batchedCommandsIDHex,
			Data:                  hex.EncodeToString(batchedCommands.GetData()),
			Status:                batchedCommands.GetStatus(),
			KeyID:                 batchedCommands.GetKeyID(),
			Signature:             signatures,
			ExecuteData:           hex.EncodeToString(executeData),
			PrevBatchedCommandsID: prevBatchedCommandsIDHex,
			CommandIDs:            commandIDs,
		}
	default:
		resp = types.QueryBatchedCommandsResponse{
			ID:                    batchedCommandsIDHex,
			Data:                  hex.EncodeToString(batchedCommands.GetData()),
			Status:                batchedCommands.GetStatus(),
			KeyID:                 batchedCommands.GetKeyID(),
			Signature:             nil,
			ExecuteData:           "",
			PrevBatchedCommandsID: prevBatchedCommandsIDHex,
			CommandIDs:            commandIDs,
		}
	}

	return resp, nil
}

// QueryAddressByKeyRole returns the current address of the given key role
func QueryAddressByKeyRole(ctx sdk.Context, s types.Signer, n types.Nexus, chainName string, keyRoleStr string) ([]byte, error) {
	keyRole, err := tss.KeyRoleFromSimpleStr(keyRoleStr)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	chain, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEVM, "%s is not a registered chain", chainName)
	}

	keyID, ok := s.GetCurrentKeyID(ctx, chain, keyRole)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEVM, "%s key not found", keyRole.SimpleString())
	}

	return QueryAddressByKeyID(ctx, s, n, chainName, keyID)
}

// QueryAddressByKeyID returns the address of the given key ID
func QueryAddressByKeyID(ctx sdk.Context, s types.Signer, n types.Nexus, chainName string, keyID tss.KeyID) ([]byte, error) {
	chain, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEVM, "%s is not a registered chain", chainName)
	}

	key, ok := s.GetKey(ctx, keyID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEVM, "threshold key %s not found", keyID)
	}

	switch chain.KeyType {
	case tss.Multisig:
		multisigPubKey, err := key.GetMultisigPubKey()
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
		}

		addressStrs := make([]string, len(multisigPubKey))
		for i, address := range types.KeysToAddresses(multisigPubKey...) {
			addressStrs[i] = address.Hex()
		}

		threshold := uint32(key.GetMultisigKey().Threshold)

		resp := types.QueryAddressResponse{
			Address: &types.QueryAddressResponse_MultisigAddresses_{MultisigAddresses: &types.QueryAddressResponse_MultisigAddresses{Addresses: addressStrs, Threshold: threshold}},
			KeyID:   keyID,
		}

		return resp.Marshal()
	case tss.Threshold:
		pk, err := key.GetECDSAPubKey()
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
		}

		address := crypto.PubkeyToAddress(pk)
		resp := types.QueryAddressResponse{
			Address: &types.QueryAddressResponse_ThresholdAddress_{ThresholdAddress: &types.QueryAddressResponse_ThresholdAddress{Address: address.Hex()}},
			KeyID:   key.ID,
		}

		return resp.Marshal()
	default:
		return nil, sdkerrors.Wrapf(types.ErrEVM, "unknown key type %s of chain %s", chain.KeyType, chain.Name)
	}
}

func queryNextMasterAddress(ctx sdk.Context, s types.Signer, n types.Nexus, chainName string) ([]byte, error) {
	chain, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", chainName))
	}

	keyID, ok := s.GetNextKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("next key ID for chain %s not set", chainName))
	}

	return QueryAddressByKeyID(ctx, s, n, chain.Name, keyID)
}

func queryAxelarGateway(ctx sdk.Context, k types.ChainKeeper, n types.Nexus) ([]byte, error) {

	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	addr, ok := k.GetGatewayAddress(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, "axelar gateway not set")
	}

	return addr.Bytes(), nil
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

// QueryDepositState returns the state of an ERC20 deposit confirmation
func QueryDepositState(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, data []byte) ([]byte, error) {
	var params types.QueryDepositStateParams
	if err := types.ModuleCdc.UnmarshalJSON(data, &params); err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, "could not unmarshal parameters")
	}

	status, log, code := queryDepositState(ctx, k, n, &params)
	if code != codes.OK {
		return nil, sdkerrors.Wrap(types.ErrEVM, log)
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&types.QueryDepositStateResponse{Status: status, Log: log})
}

func queryDepositState(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, params *types.QueryDepositStateParams) (types.DepositStatus, string, codes.Code) {
	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return -1, fmt.Sprintf("%s is not a registered chain", k.GetName()), codes.NotFound
	}

	pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s_%s", params.TxID.Hex(), params.BurnerAddress.Hex(), params.Amount))
	_, isPending := k.GetPendingDeposit(ctx, pollKey)
	_, state, ok := k.GetDeposit(ctx, common.Hash(params.TxID), common.Address(params.BurnerAddress))

	switch {
	case isPending:
		return types.DepositStatus_Pending, "deposit transaction is waiting for confirmation", codes.OK
	case !isPending && !ok:
		return types.DepositStatus_None, "deposit transaction is not confirmed", codes.OK
	case state == types.DepositStatus_Confirmed:
		return types.DepositStatus_Confirmed, "deposit transaction is confirmed", codes.OK
	case state == types.DepositStatus_Burned:
		return types.DepositStatus_Burned, "deposit has been transferred to the destination chain", codes.OK
	default:
		return -1, "deposit is in an unexpected state", codes.Internal
	}
}

func queryBytecode(ctx sdk.Context, k types.ChainKeeper, s types.Signer, n types.Nexus, contract string) ([]byte, error) {
	chain, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	var bz []byte
	switch strings.ToLower(contract) {
	case BCGateway:
		bz, _ = k.GetGatewayByteCode(ctx)
	case BCGatewayDeployment:
		deploymentBytecode, err := getGatewayDeploymentBytecode(ctx, k, s, chain)
		if err != nil {
			return nil, err
		}

		return deploymentBytecode, nil
	case BCToken:
		bz, _ = k.GetTokenByteCode(ctx)
	case BCBurner:
		bz, _ = k.GetBurnerByteCode(ctx)
	}

	if bz == nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not retrieve bytecodes for chain %s", k.GetName()))
	}

	return bz, nil
}

// QueryBatchedCommands returns the batched commands for the given ID
func QueryBatchedCommands(ctx sdk.Context, k types.ChainKeeper, s types.Signer, n types.Nexus, batchedCommandsIDHex string) ([]byte, error) {
	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	batchedCommandsID, err := hex.DecodeString(batchedCommandsIDHex)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("invalid batched commands ID: %v", err))
	}

	batchedCommands, ok := getBatchedCommands(ctx, k, batchedCommandsID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("batched commands with ID %s not found", batchedCommandsIDHex))
	}

	resp, err := batchedCommandsToQueryResp(ctx, batchedCommands, s)
	if err != nil {
		return nil, err
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

func getBatchedCommands(ctx sdk.Context, k types.ChainKeeper, id []byte) (types.CommandBatch, bool) {
	if batchedCommands := k.GetBatchByID(ctx, id); !batchedCommands.Is(types.BatchNonExistent) {
		return batchedCommands, true
	}

	return types.CommandBatch{}, false
}

func getBatchedCommandsSig(pair tss.SigKeyPair, batchedCommands types.Hash) (types.Signature, error) {
	pk, err := pair.GetKey()
	if err != nil {
		return types.Signature{}, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not parse pub key: %v", err))
	}

	sig, err := pair.GetSig()
	if err != nil {
		return types.Signature{}, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not parse signature: %v", err))
	}

	batchedCommandsSig, err := types.ToSignature(sig, common.Hash(batchedCommands), pk)
	if err != nil {
		return types.Signature{}, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not create recoverable signature: %v", err))
	}
	return batchedCommandsSig, nil
}

func queryChains(ctx sdk.Context, n types.Nexus) ([]byte, error) {
	evmChains := getEVMChains(ctx, n)

	response := types.QueryChainsResponse{Chains: evmChains}
	return response.Marshal()
}

func getEVMChains(ctx sdk.Context, n types.Nexus) []string {
	chains := n.GetChains(ctx)

	var evmChains []string
	for _, c := range chains {
		if c.Module == types.ModuleName {
			evmChains = append(evmChains, c.Name)
		}
	}
	return evmChains
}
