package keeper

import (
	"bytes"
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
	BCGatewaySimple     = "simple-gateway"
	BCGatewayDeployment = "deployment-gateway"
	BCToken             = "token"
	BCBurner            = "burner"
)

// NewQuerier returns a new querier for the evm module
func NewQuerier(k types.BaseKeeper, s types.Signer, n types.Nexus) sdk.Querier {

	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var chainKeeper types.ChainKeeper
		if len(path) > 1 {
			chainKeeper = k.ForChain(ctx, path[1])
		}

		switch path[0] {
		case QAddressByKeyRole:
			return QueryAddressByKeyRole(ctx, s, n, path[1], path[2])
		case QAddressByKeyID:
			return QueryAddressByKeyID(ctx, s, n, path[1], path[2])
		case QNextMasterAddress:
			return queryNextMasterAddress(ctx, s, n, path[1])
		case QAxelarGatewayAddress:
			return queryAxelarGateway(ctx, chainKeeper, n)
		case QTokenAddress:
			return QueryTokenAddress(ctx, chainKeeper, n, path[2])
		case QDepositState:
			return QueryDepositState(ctx, chainKeeper, n, path[2], path[3])
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
	var batchedCommands types.BatchedCommands

	unsignedBatchedCommands, ok := keeper.GetUnsignedBatchedCommands(ctx)
	if ok {
		batchedCommands = unsignedBatchedCommands
	} else {
		latestSignedBatchedCommandsID, ok := keeper.GetLatestSignedBatchedCommandsID(ctx)
		if !ok {
			return nil, fmt.Errorf("no batched commands exist for chain %s", keeper.GetName())
		}

		latestSignedBatchedCommands, ok := keeper.GetSignedBatchedCommands(ctx, latestSignedBatchedCommandsID)
		if !ok {
			return nil, fmt.Errorf("cannot find the latest signed batched commands for chain %s", keeper.GetName())
		}

		batchedCommands = latestSignedBatchedCommands
	}

	resp, err := batchedCommandsToQueryResp(ctx, batchedCommands, s)
	if err != nil {
		return nil, err
	}

	return types.ModuleCdc.MarshalBinaryLengthPrefixed(&resp)
}

func batchedCommandsToQueryResp(ctx sdk.Context, batchedCommands types.BatchedCommands, s types.Signer) (types.QueryBatchedCommandsResponse, error) {
	batchedCommandsIDHex := hex.EncodeToString(batchedCommands.ID)
	prevBatchedCommandsIDHex := ""
	if batchedCommands.PrevBatchedCommandsID != nil {
		prevBatchedCommandsIDHex = hex.EncodeToString(batchedCommands.PrevBatchedCommandsID)
	}

	var resp types.QueryBatchedCommandsResponse

	switch batchedCommands.Status {
	case types.Signed:
		sig, sigStatus := s.GetSig(ctx, batchedCommandsIDHex)
		if sigStatus != tss.SigStatus_Signed {
			return resp, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding signature for sig ID %s", batchedCommandsIDHex))
		}

		key, ok := s.GetKey(ctx, batchedCommands.KeyID)
		if !ok {
			return resp, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding key for batched commands with ID %s", batchedCommandsIDHex))
		}

		batchedCommandsSig, err := types.ToSignature(sig, common.Hash(batchedCommands.SigHash), key.Value)
		if err != nil {
			return resp, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not create recoverable signature: %v", err))
		}

		executeData, err := types.CreateExecuteData(batchedCommands.Data, batchedCommandsSig)
		if err != nil {
			return resp, sdkerrors.Wrapf(types.ErrEVM, "could not create transaction data: %s", err)
		}

		resp = types.QueryBatchedCommandsResponse{
			ID:                    batchedCommandsIDHex,
			Data:                  hex.EncodeToString(batchedCommands.Data),
			Status:                batchedCommands.Status,
			KeyID:                 batchedCommands.KeyID,
			Signature:             hex.EncodeToString(batchedCommandsSig[:]),
			ExecuteData:           hex.EncodeToString(executeData),
			PrevBatchedCommandsID: prevBatchedCommandsIDHex,
		}
	default:
		resp = types.QueryBatchedCommandsResponse{
			ID:                    batchedCommandsIDHex,
			Data:                  hex.EncodeToString(batchedCommands.Data),
			Status:                batchedCommands.Status,
			KeyID:                 batchedCommands.KeyID,
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
		return nil, err
	}

	address, key, err := getAddressAndKeyForRole(ctx, s, n, chainName, keyRole)
	if err != nil {
		return nil, err
	}

	resp := types.QueryAddressResponse{Address: address.Hex(), KeyID: key.ID}

	return types.ModuleCdc.MarshalBinaryLengthPrefixed(&resp)
}

// QueryAddressByKeyID returns the address of the given key ID
func QueryAddressByKeyID(ctx sdk.Context, s types.Signer, n types.Nexus, chainName string, keyID string) ([]byte, error) {
	_, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", chainName)
	}

	key, ok := s.GetKey(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("key %s not found", keyID)
	}

	address := crypto.PubkeyToAddress(key.Value)
	resp := types.QueryAddressResponse{Address: address.Hex(), KeyID: key.ID}

	return types.ModuleCdc.MarshalBinaryLengthPrefixed(&resp)
}

// QueryDepositAddress returns the deposit address linked to the given recipient address
func QueryDepositAddress(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, data []byte) ([]byte, error) {
	depositChain, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}
	var params types.DepositQueryParams
	if err := types.ModuleCdc.UnmarshalJSON(data, &params); err != nil {
		return nil, fmt.Errorf("could not parse the recipient")
	}

	gatewayAddr, ok := k.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	tokenAddr, err := k.GetTokenAddress(ctx, params.Symbol, gatewayAddr)
	if err != nil {
		return nil, err
	}

	depositAddr, _, err := k.GetBurnerAddressAndSalt(ctx, tokenAddr, params.Address, gatewayAddr)
	if err != nil {
		return nil, err
	}

	_, ok = n.GetRecipient(ctx, nexus.CrossChainAddress{Chain: depositChain, Address: depositAddr.String()})
	if !ok {
		return nil, fmt.Errorf("deposit address is not linked with recipient address")
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
func QueryTokenAddress(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, symbol string) ([]byte, error) {

	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	gateway, ok := k.GetGatewayAddress(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, "axelar gateway not set")
	}

	addr, err := k.GetTokenAddress(ctx, symbol, gateway)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	return addr.Bytes(), nil
}

// QueryDepositState returns the state of an ERC20 deposit confirmation
func QueryDepositState(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, txID string, depositAddress string) ([]byte, error) {

	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	pollKey := vote.NewPollKey(types.ModuleName, txID+"_"+depositAddress)
	_, isPending := k.GetPendingDeposit(ctx, pollKey)
	_, state, ok := k.GetDeposit(ctx, common.HexToHash(txID), common.HexToAddress(depositAddress))

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
		return nil, fmt.Errorf("deposit is in an unexpected state")
	}

	return types.ModuleCdc.MarshalBinaryLengthPrefixed(&depositState)
}

func queryBytecode(ctx sdk.Context, k types.ChainKeeper, s types.Signer, n types.Nexus, contract string) ([]byte, error) {

	_, ok := n.GetChain(ctx, k.GetName())
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", k.GetName()))
	}

	var bz []byte
	switch strings.ToLower(contract) {
	case BCGatewaySimple:
		bz, _ = k.GetGatewayByteCodes(ctx)
	case BCGatewayDeployment:
		keyRole, err := tss.KeyRoleFromSimpleStr(tss.SecondaryKey.SimpleString())
		if err != nil {
			return nil, err
		}

		address, _, err := getAddressAndKeyForRole(ctx, s, n, k.GetName(), keyRole)
		if err != nil {
			return nil, err
		}

		bz, _ = k.GetGatewayByteCodes(ctx)
		deploymentBytecode, err := types.GetGatewayDeploymentBytecode(bz, address)
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

	pk, ok := s.GetKeyForSigID(ctx, txID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding key for sig ID %s", txID))
	}

	sig, status := s.GetSig(ctx, txID)
	if status != tss.SigStatus_Signed {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding signature for sig ID %s", txID))
	}

	signedTx, err := k.AssembleTx(ctx, txID, pk.Value, sig)
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

	return types.ModuleCdc.MarshalBinaryLengthPrefixed(&resp)
}

func getBatchedCommands(ctx sdk.Context, k types.ChainKeeper, id []byte) (types.BatchedCommands, bool) {
	if batchedCommands, ok := k.GetSignedBatchedCommands(ctx, id); ok {
		return batchedCommands, true
	}

	if batchedCommands, ok := k.GetUnsignedBatchedCommands(ctx); ok && bytes.Equal(batchedCommands.ID, id) {
		return batchedCommands, true
	}

	return types.BatchedCommands{}, false
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
