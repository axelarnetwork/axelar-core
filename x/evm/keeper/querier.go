package keeper

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	ethereumRoot "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Query labels
const (
	QTokenAddress         = "token-address"
	QMasterAddress        = "master-address"
	QNextMasterAddress    = "next-master-address"
	QKeyAddress           = "query-key-address"
	QAxelarGatewayAddress = "gateway-address"
	QCommandData          = "command-data"
	QDepositAddress       = "deposit-address"
	CreateDeployTx        = "deploy-gateway"
	SendTx                = "send-tx"
	SendCommand           = "send-command"
)

// NewQuerier returns a new querier for the evm module
func NewQuerier(rpcs map[string]types.RPCClient, k Keeper, s types.Signer, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QMasterAddress:
			return queryMasterAddress(ctx, s, n, path[1])
		case QNextMasterAddress:
			return queryNextMasterAddress(ctx, s, n, path[1])
		case QKeyAddress:
			return queryKeyAddress(ctx, s, req.Data)
		case QAxelarGatewayAddress:
			return queryAxelarGateway(ctx, k, n, path[1])
		case QTokenAddress:
			return QueryTokenAddress(ctx, k, n, path[1], path[2])
		case QCommandData:
			return queryCommandData(ctx, k, s, n, path[1], path[2])
		case QDepositAddress:
			return QueryDepositAddress(ctx, k, n, path[1], req.Data)
		case CreateDeployTx:
			return createDeployGateway(ctx, k, rpcs, s, n, req.Data)
		case SendTx:
			return sendSignedTx(ctx, k, rpcs, s, n, path[1], path[2])
		case SendCommand:
			return createTxAndSend(ctx, k, rpcs, s, n, req.Data)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown evm-bridge query endpoint: %s", path[0]))
		}
	}
}

// QueryDepositAddress returns the deposit address linked to the given recipient address
func QueryDepositAddress(ctx sdk.Context, k types.EVMKeeper, n types.Nexus, chainName string, data []byte) ([]byte, error) {
	depositChain, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", chainName))
	}
	var params types.DepositQueryParams
	if err := types.ModuleCdc.UnmarshalJSON(data, &params); err != nil {
		return nil, fmt.Errorf("could not parse the recipient")
	}

	gatewayAddr, ok := k.GetGatewayAddress(ctx, chainName)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	tokenAddr, err := k.GetTokenAddress(ctx, chainName, params.Symbol, gatewayAddr)
	if err != nil {
		return nil, err
	}

	depositAddr, _, err := k.GetBurnerAddressAndSalt(ctx, chainName, tokenAddr, params.Address, gatewayAddr)
	if err != nil {
		return nil, err
	}

	_, ok = n.GetRecipient(ctx, nexus.CrossChainAddress{Chain: depositChain, Address: depositAddr.String()})
	if !ok {
		return nil, fmt.Errorf("deposit address is not linked with recipient address")
	}

	return depositAddr.Bytes(), nil
}

func queryMasterAddress(ctx sdk.Context, s types.Signer, n types.Nexus, chainName string) ([]byte, error) {

	chain, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", chainName))
	}

	pk, ok := s.GetCurrentKey(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, "key not found")
	}

	fromAddress := crypto.PubkeyToAddress(pk.Value)

	resp := types.QueryMasterAddressResponse{
		MasterAddress: fromAddress.Bytes(),
		MasterKeyId:   pk.ID,
	}

	return resp.Marshal()
}

func queryKeyAddress(ctx sdk.Context, s types.Signer, keyIDBytes []byte) ([]byte, error) {
	keyID := string(keyIDBytes)
	pk, ok := s.GetKey(ctx, keyID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, "no key found for key ID "+keyID)
	}

	fromAddress := crypto.PubkeyToAddress(pk.Value)

	bz := fromAddress.Bytes()

	return bz, nil
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

func queryAxelarGateway(ctx sdk.Context, k Keeper, n types.Nexus, chainName string) ([]byte, error) {

	_, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", chainName))
	}

	addr, ok := k.GetGatewayAddress(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, "axelar gateway not set")
	}

	return addr.Bytes(), nil
}

// QueryTokenAddress returns the address of the token contract with the given parameters
func QueryTokenAddress(ctx sdk.Context, k types.EVMKeeper, n types.Nexus, chainName, symbol string) ([]byte, error) {

	_, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", chainName))
	}

	gateway, ok := k.GetGatewayAddress(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, "axelar gateway not set")
	}

	addr, err := k.GetTokenAddress(ctx, chainName, symbol, gateway)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	return addr.Bytes(), nil
}

/*
  Create a transaction for smart contract deployment. See:

  https://goethereumbook.org/en/smart-contract-deploy/
  https://gist.github.com/tomconte/6ce22128b15ba36bb3d7585d5180fba0

  If gasLimit is set to 0, the function will attempt to estimate the amount of gas needed
*/
func createDeployGateway(ctx sdk.Context, k Keeper, rpcs map[string]types.RPCClient, s types.Signer, n types.Nexus, data []byte) ([]byte, error) {
	var params types.DeployParams
	err := types.ModuleCdc.LegacyAmino.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	rpc, found := rpcs[strings.ToLower(params.Chain)]
	if !found {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find RPC for chain '%s'", params.Chain))
	}

	contractOwner, err := getContractOwner(ctx, s, n, params.Chain)
	if err != nil {
		return nil, err
	}

	nonce, err := rpc.PendingNonceAt(context.Background(), contractOwner)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not create nonce: %s", err))
	}

	gasPrice := params.GasPrice.BigInt()
	if params.GasPrice.IsZero() {
		gasPrice, err = rpc.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not calculate gas price: %s", err))
		}
	}

	byteCodes, ok := k.GetGatewayByteCodes(ctx, params.Chain)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("Could not retrieve gateway bytecodes for chain %s", params.Chain))
	}

	gasLimit := params.GasLimit
	if gasLimit == 0 {
		gasLimit, err = rpc.EstimateGas(context.Background(), ethereumRoot.CallMsg{
			To:   nil,
			Data: byteCodes,
		})

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not estimate gas limit: %s", err))
		}
	}

	tx := ethTypes.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, byteCodes)
	result := types.DeployResult{
		Tx:              tx,
		ContractAddress: crypto.CreateAddress(contractOwner, nonce).String(),
	}
	k.Logger(ctx).Debug(fmt.Sprintf("Contract address: %s", result.ContractAddress))
	return types.ModuleCdc.LegacyAmino.MustMarshalJSON(result), nil
}

func sendSignedTx(ctx sdk.Context, k Keeper, rpcs map[string]types.RPCClient, s types.Signer, n types.Nexus, chainName, txID string) ([]byte, error) {

	_, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", chainName))
	}

	rpc, found := rpcs[strings.ToLower(chainName)]
	if !found {
		return nil, fmt.Errorf("could not find RPC for chain '%s'", chainName)
	}

	pk, ok := s.GetKeyForSigID(ctx, txID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding key for sig ID %s", txID))
	}

	sig, ok := s.GetSig(ctx, txID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding signature for sig ID %s", txID))
	}

	signedTx, err := k.AssembleEthTx(ctx, chainName, txID, pk.Value, sig)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not insert generated signature: %v", err))
	}

	err = rpc.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	return signedTx.Hash().Bytes(), nil
}

func createTxAndSend(ctx sdk.Context, k Keeper, rpcs map[string]types.RPCClient, s types.Signer, n types.Nexus, data []byte) ([]byte, error) {
	var params types.CommandParams
	err := types.ModuleCdc.LegacyAmino.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	rpc, found := rpcs[strings.ToLower(params.Chain)]
	if !found {
		return nil, fmt.Errorf("could not find RPC for chain '%s'", params.Chain)
	}

	_, ok := n.GetChain(ctx, params.Chain)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", params.Chain))
	}

	commandIDHex := common.Bytes2Hex(params.CommandID[:])
	sig, ok := s.GetSig(ctx, commandIDHex)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding signature for sig ID %s", commandIDHex))
	}

	pk, ok := s.GetKeyForSigID(ctx, commandIDHex)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding key for sig ID %s", commandIDHex))
	}

	commandData := k.GetCommandData(ctx, params.Chain, params.CommandID)
	commandSig, err := types.ToEthSignature(sig, types.GetEthereumSignHash(commandData), pk.Value)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not create recoverable signature: %v", err))
	}

	executeData, err := types.CreateExecuteData(commandData, commandSig)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVM, "could not create transaction data: %s", err)
	}

	k.Logger(ctx).Debug(common.Bytes2Hex(executeData))

	contractAddr, ok := k.GetGatewayAddress(ctx, params.Chain)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEVM, "axelar gateway not deployed yet")
	}

	msg := ethereumRoot.CallMsg{
		From: common.HexToAddress(params.Sender),
		To:   &contractAddr,
		Data: executeData,
		Gas:  uint64(5000000),
	}

	txHash, err := rpc.SendAndSignTransaction(context.Background(), msg)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVM, "could not send transaction: %s", err)
	}

	return common.FromHex(txHash), nil
}

func queryCommandData(ctx sdk.Context, k Keeper, s types.Signer, n types.Nexus, chainName, commandIDHex string) ([]byte, error) {

	_, ok := n.GetChain(ctx, chainName)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", chainName))
	}

	sig, ok := s.GetSig(ctx, commandIDHex)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding signature for sig ID %s", commandIDHex))
	}

	pk, ok := s.GetKeyForSigID(ctx, commandIDHex)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding key for sig ID %s", commandIDHex))
	}

	var commandID types.CommandID
	copy(commandID[:], common.Hex2Bytes(commandIDHex))

	commandData := k.GetCommandData(ctx, chainName, commandID)
	commandSig, err := types.ToEthSignature(sig, types.GetEthereumSignHash(commandData), pk.Value)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not create recoverable signature: %v", err))
	}

	executeData, err := types.CreateExecuteData(commandData, commandSig)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEVM, "could not create transaction data: %s", err)
	}

	return executeData, nil
}

func getContractOwner(ctx sdk.Context, s types.Signer, n types.Nexus, chainName string) (common.Address, error) {
	chain, ok := n.GetChain(ctx, chainName)
	if !ok {
		return common.Address{}, sdkerrors.Wrap(types.ErrEVM, fmt.Errorf("%s is not a registered chain", chainName).Error())
	}

	pk, ok := s.GetCurrentKey(ctx, chain, tss.MasterKey)
	if !ok {
		return common.Address{}, fmt.Errorf("key not found")
	}

	fromAddress := crypto.PubkeyToAddress(pk.Value)
	return fromAddress, nil
}
