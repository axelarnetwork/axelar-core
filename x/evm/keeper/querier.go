package keeper

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Query labels
const (
	QTokenAddress          = "token-address"
	QDepositState          = "deposit-state"
	QAddressByKeyRole      = "address-by-key-role"
	QAddressByKeyID        = "address-by-key-id"
	QNextMasterAddress     = "next-master-address"
	QAxelarGatewayAddress  = "gateway-address"
	QDepositAddress        = "deposit-address"
	QBytecode              = "bytecode"
	QSignedTx              = "signed-tx"
	QLatestBatchedCommands = "latest-batched-commands"
	QBatchedCommands       = "batched-commands"
)

//Bytecode labels
const (
	BCGateway           = "gateway"
	BCGatewayDeployment = "gateway-deployment"
	BCToken             = "token"
	BCBurner            = "burner"
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
		case QTokenAddress:
			return QueryTokenAddress(ctx, chainKeeper, n, path[2])
		case QDepositState:
			return QueryDepositState(ctx, chainKeeper, n, req.Data)
		case QBatchedCommands:
			return QueryBatchedCommands(ctx, chainKeeper, s, n, path[2])
		case QLatestBatchedCommands:
			return QueryLatestBatchedCommands(ctx, chainKeeper, s)
		case QDepositAddress:
			return QueryDepositAddress(ctx, chainKeeper, n, req.Data)
		case QBytecode:
			return queryBytecode(ctx, chainKeeper, s, n, path[2])
		case QSignedTx:
			return querySignedTx(ctx, chainKeeper, s, n, path[2])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown evm-bridge query endpoint: %s", path[0]))
		}
	}
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

	var resp types.QueryBatchedCommandsResponse

	switch {
	case batchedCommands.Is(types.BatchSigned):
		sig, sigStatus := s.GetSig(ctx, batchedCommandsIDHex)
		if sigStatus != tss.SigStatus_Signed {
			return resp, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding signature for sig ID %s", batchedCommandsIDHex))
		}

		key, ok := s.GetKey(ctx, batchedCommands.GetKeyID())
		if !ok {
			return resp, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding key for batched commands with ID %s", batchedCommandsIDHex))
		}

		batchedCommandsSig, err := types.ToSignature(sig, common.Hash(batchedCommands.GetSigHash()), key.Value)
		if err != nil {
			return resp, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not create recoverable signature: %v", err))
		}

		executeData, err := types.CreateExecuteData(batchedCommands.GetData(), batchedCommandsSig)
		if err != nil {
			return resp, sdkerrors.Wrapf(types.ErrEVM, "could not create transaction data: %s", err)
		}

		resp = types.QueryBatchedCommandsResponse{
			ID:                    batchedCommandsIDHex,
			Data:                  hex.EncodeToString(batchedCommands.GetData()),
			Status:                batchedCommands.GetStatus(),
			KeyID:                 batchedCommands.GetKeyID(),
			Signature:             hex.EncodeToString(batchedCommandsSig[:]),
			ExecuteData:           hex.EncodeToString(executeData),
			PrevBatchedCommandsID: prevBatchedCommandsIDHex,
		}
	default:
		resp = types.QueryBatchedCommandsResponse{
			ID:                    batchedCommandsIDHex,
			Data:                  hex.EncodeToString(batchedCommands.GetData()),
			Status:                batchedCommands.GetStatus(),
			KeyID:                 batchedCommands.GetKeyID(),
			Signature:             "",
			ExecuteData:           "",
			PrevBatchedCommandsID: prevBatchedCommandsIDHex,
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

	address, key, err := getAddressAndKeyForRole(ctx, s, n, chainName, keyRole)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	resp := types.QueryAddressResponse{Address: address.Hex(), KeyID: key.ID}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

// QueryAddressByKeyID returns the address of the given key ID
func QueryAddressByKeyID(ctx sdk.Context, s types.Signer, n types.Nexus, chainName string, keyID tss.KeyID) ([]byte, error) {
	_, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", chainName))
	}

	key, ok := s.GetKey(ctx, keyID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("key %s not found", keyID))
	}

	address := crypto.PubkeyToAddress(key.Value)
	resp := types.QueryAddressResponse{Address: address.Hex(), KeyID: key.ID}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

// QueryDepositAddress returns the deposit address linked to the given recipient address
func QueryDepositAddress(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, data []byte) ([]byte, error) {
	depositChain, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}
	var params types.DepositQueryParams
	if err := types.ModuleCdc.UnmarshalJSON(data, &params); err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not parse the recipient"))
	}

	gatewayAddr, ok := k.GetGatewayAddress(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("axelar gateway address not set"))
	}

	token := k.GetERC20Token(ctx, params.Asset)
	if !token.Is(types.Confirmed) {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("token for asset '%s' not confirmed", params.Asset))
	}

	depositAddr, _, err := k.GetBurnerAddressAndSalt(ctx, token.GetAddress(), params.Address, gatewayAddr)
	if err != nil {
		return nil, err
	}

	_, ok = n.GetRecipient(ctx, nexus.CrossChainAddress{Chain: depositChain, Address: depositAddr.String()})
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("deposit address is not linked with recipient address"))
	}

	return depositAddr.Bytes(), nil
}

func queryNextMasterAddress(ctx sdk.Context, s types.Signer, n types.Nexus, chainName string) ([]byte, error) {

	chain, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", chainName))
	}

	pk, ok := s.GetNextKey(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, "next key not found")
	}

	fromAddress := crypto.PubkeyToAddress(pk.Value)

	bz := fromAddress.Bytes()

	return bz, nil
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

// QueryTokenAddress returns the address of the token contract with the given parameters
func QueryTokenAddress(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, asset string) ([]byte, error) {
	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	token := k.GetERC20Token(ctx, asset)
	if token.Is(types.NonExistent) {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("token for asset '%s' non-existent", asset))
	}

	return token.GetAddress().Bytes(), nil
}

// QueryDepositState returns the state of an ERC20 deposit confirmation
func QueryDepositState(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, data []byte) ([]byte, error) {
	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	var params types.QueryDepositStateParams
	if err := types.ModuleCdc.UnmarshalJSON(data, &params); err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not unmarshal parameters"))
	}

	pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s_%d", params.TxID.Hex(), params.BurnerAddress.Hex(), params.Amount))
	_, isPending := k.GetPendingDeposit(ctx, pollKey)
	_, state, ok := k.GetDeposit(ctx, common.Hash(params.TxID), common.Address(params.BurnerAddress))

	var depositState types.QueryDepositStateResponse
	switch {
	case isPending:
		depositState = types.QueryDepositStateResponse{Status: types.DepositStatus_Pending, Log: "deposit transaction is waiting for confirmation"}
	case !isPending && !ok:
		depositState = types.QueryDepositStateResponse{Status: types.DepositStatus_None, Log: "deposit transaction is not confirmed"}
	case state == types.CONFIRMED:
		depositState = types.QueryDepositStateResponse{Status: types.DepositStatus_Confirmed, Log: "deposit transaction is confirmed"}
	case state == types.BURNED:
		depositState = types.QueryDepositStateResponse{Status: types.DepositStatus_Burned, Log: "deposit has been transferred to the destination chain"}
	default:
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("deposit is in an unexpected state"))
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&depositState)
}

func queryBytecode(ctx sdk.Context, k types.ChainKeeper, s types.Signer, n types.Nexus, contract string) ([]byte, error) {
	chain, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	var bz []byte
	switch strings.ToLower(contract) {
	case BCGateway:
		bz, _ = k.GetGatewayByteCodes(ctx)
	case BCGatewayDeployment:
		deploymentBytecode, err := getGatewayDeploymentBytecode(ctx, k, s, chain)
		if err != nil {
			return nil, err
		}

		return deploymentBytecode, nil
	case BCToken:
		bz, _ = k.GetTokenByteCodes(ctx)
	case BCBurner:
		bz, _ = k.GetBurnerByteCodes(ctx)
	}

	if bz == nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not retrieve bytecodes for chain %s", k.GetName()))
	}

	return bz, nil
}

func querySignedTx(ctx sdk.Context, k types.ChainKeeper, s types.Signer, n types.Nexus, txID string) ([]byte, error) {

	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	sig, status := s.GetSig(ctx, txID)
	if status != tss.SigStatus_Signed {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find signature for tx '%s' (current status: %s)", txID, status.String()))
	}

	signedTx, err := k.AssembleTx(ctx, txID, sig)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not insert generated signature: %v", err))
	}

	return signedTx.MarshalBinary()
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

func getAddressAndKeyForRole(ctx sdk.Context, s types.Signer, n types.Nexus, chainName string, keyRole tss.KeyRole) (common.Address, tss.Key, error) {
	chain, ok := n.GetChain(ctx, chainName)
	if !ok {
		return common.Address{}, tss.Key{}, fmt.Errorf("%s is not a registered chain", chainName)
	}

	key, ok := s.GetCurrentKey(ctx, chain, keyRole)
	if !ok {
		return common.Address{}, tss.Key{}, fmt.Errorf("%s key not found", keyRole.SimpleString())
	}

	return crypto.PubkeyToAddress(key.Value), key, nil
}
