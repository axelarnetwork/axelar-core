package keeper

import (
	"context"
	"fmt"
	"math/big"

	ethereumRoot "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Query labels
const (
	QueryTokenAddress         = "token-address"
	QueryMasterAddress        = "master-address"
	QueryAxelarGatewayAddress = "gateway-address"
	QueryCommandData          = "command-data"
	QueryDepositAddress		  = "deposit-address"
	CreateDeployTx            = "deploy-gateway"
	SendTx                    = "send-tx"
	SendCommand               = "send-command"
)

// NewQuerier returns a new querier for the ethereum module
func NewQuerier(rpc types.RPCClient, k Keeper, s types.Signer, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryMasterAddress:
			return queryMasterAddress(ctx, s)
		case QueryAxelarGatewayAddress:
			return queryAxelarGateway(ctx, k)
		case QueryTokenAddress:
			return queryTokenAddress(ctx, k, path[1])
		case QueryCommandData:
			return queryCommandData(ctx, k, s, path[1])
		case QueryDepositAddress:
			return queryDepositAddress(ctx, k, n, req.Data)
		case CreateDeployTx:
			return createDeployGateway(ctx, k, rpc, s, req.Data)
		case SendTx:
			return sendSignedTx(ctx, k, rpc, s, path[1])
		case SendCommand:
			return createTxAndSend(ctx, k, rpc, s, req.Data)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown eth-bridge query endpoint: %s", path[0]))
		}
	}
}

func queryDepositAddress(ctx sdk.Context, k types.EthKeeper, n types.Nexus, data []byte) ([]byte, error) {
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

	depositChain, ok := n.GetChain(ctx, exported.Ethereum.Name)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", exported.Ethereum.Name)
	}

	_, ok = n.GetRecipient(ctx, nexus.CrossChainAddress{Chain: depositChain, Address: depositAddr.String()})
	if !ok {
		return nil, fmt.Errorf("deposit address is not linked with recipient address")
	}

	return depositAddr.Bytes(), nil
}

func queryMasterAddress(ctx sdk.Context, s types.Signer) ([]byte, error) {

	pk, ok := s.GetCurrentKey(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "key not found")
	}

	fromAddress := crypto.PubkeyToAddress(pk.Value)

	bz := fromAddress.Bytes()

	return bz, nil
}

func queryAxelarGateway(ctx sdk.Context, k Keeper) ([]byte, error) {

	addr, ok := k.GetGatewayAddress(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "axelar gateway not set")
	}

	return addr.Bytes(), nil
}

func queryTokenAddress(ctx sdk.Context, k types.EthKeeper, symbol string) ([]byte, error) {

	gateway, ok := k.GetGatewayAddress(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "axelar gateway not set")
	}

	addr, err := k.GetTokenAddress(ctx, symbol, gateway)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	return addr.Bytes(), nil
}

/*
  Create a transaction for smart contract deployment. See:

  https://goethereumbook.org/en/smart-contract-deploy/
  https://gist.github.com/tomconte/6ce22128b15ba36bb3d7585d5180fba0

  If gasLimit is set to 0, the function will attempt to estimate the amount of gas needed
*/
func createDeployGateway(ctx sdk.Context, k Keeper, rpc types.RPCClient, s types.Signer, data []byte) ([]byte, error) {
	var params types.DeployParams
	err := types.ModuleCdc.LegacyAmino.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	contractOwner, err := getContractOwner(ctx, s)
	if err != nil {
		return nil, err
	}

	nonce, err := rpc.PendingNonceAt(context.Background(), contractOwner)
	if err != nil {
		return nil, fmt.Errorf("could not create nonce: %s", err)
	}

	gasPrice := params.GasPrice.BigInt()
	if params.GasPrice.IsZero() {
		gasPrice, err = rpc.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, fmt.Errorf("could not calculate gas price: %s", err)
		}
	}

	gasLimit := params.GasLimit
	if gasLimit == 0 {
		gasLimit, err = rpc.EstimateGas(context.Background(), ethereumRoot.CallMsg{
			To:   nil,
			Data: k.GetGatewayByteCodes(ctx),
		})

		if err != nil {
			return nil, fmt.Errorf("could not estimate gas limit: %s", err)
		}
	}

	tx := ethTypes.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, k.GetGatewayByteCodes(ctx))
	result := types.DeployResult{
		Tx:              tx,
		ContractAddress: crypto.CreateAddress(contractOwner, nonce).String(),
	}
	k.Logger(ctx).Debug(fmt.Sprintf("Contract address: %s", result.ContractAddress))
	return types.ModuleCdc.LegacyAmino.MustMarshalJSON(result), nil
}

func sendSignedTx(ctx sdk.Context, k Keeper, rpc types.RPCClient, s types.Signer, txID string) ([]byte, error) {
	pk, ok := s.GetKeyForSigID(ctx, txID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding key for sig ID %s", txID))
	}

	sig, ok := s.GetSig(ctx, txID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding signature for sig ID %s", txID))
	}

	signedTx, err := k.AssembleEthTx(ctx, txID, pk.Value, sig)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not insert generated signature: %v", err))
	}

	err = rpc.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	return signedTx.Hash().Bytes(), nil
}

func createTxAndSend(ctx sdk.Context, k Keeper, rpc types.RPCClient, s types.Signer, data []byte) ([]byte, error) {
	var params types.CommandParams
	err := types.ModuleCdc.LegacyAmino.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	commandIDHex := common.Bytes2Hex(params.CommandID[:])
	sig, ok := s.GetSig(ctx, commandIDHex)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding signature for sig ID %s", commandIDHex))
	}

	pk, ok := s.GetKeyForSigID(ctx, commandIDHex)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding key for sig ID %s", commandIDHex))
	}

	commandData := k.GetCommandData(ctx, params.CommandID)
	commandSig, err := types.ToEthSignature(sig, types.GetEthereumSignHash(commandData), pk.Value)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not create recoverable signature: %v", err))
	}

	executeData, err := types.CreateExecuteData(commandData, commandSig)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "could not create transaction data: %s", err)
	}

	k.Logger(ctx).Debug(common.Bytes2Hex(executeData))

	contractAddr, ok := k.GetGatewayAddress(ctx)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "axelar gateway not deployed yet")
	}

	msg := ethereumRoot.CallMsg{
		From: common.HexToAddress(params.Sender),
		To:   &contractAddr,
		Data: executeData,
		Gas:  uint64(5000000),
	}

	txHash, err := rpc.SendAndSignTransaction(context.Background(), msg)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "could not send transaction: %s", err)
	}

	return common.FromHex(txHash), nil
}

func queryCommandData(ctx sdk.Context, k Keeper, s types.Signer, commandIDHex string) ([]byte, error) {
	sig, ok := s.GetSig(ctx, commandIDHex)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding signature for sig ID %s", commandIDHex))
	}

	pk, ok := s.GetKeyForSigID(ctx, commandIDHex)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding key for sig ID %s", commandIDHex))
	}

	var commandID types.CommandID
	copy(commandID[:], common.Hex2Bytes(commandIDHex))

	commandData := k.GetCommandData(ctx, commandID)
	commandSig, err := types.ToEthSignature(sig, types.GetEthereumSignHash(commandData), pk.Value)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not create recoverable signature: %v", err))
	}

	executeData, err := types.CreateExecuteData(commandData, commandSig)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "could not create transaction data: %s", err)
	}

	return executeData, nil
}

func getContractOwner(ctx sdk.Context, s types.Signer) (common.Address, error) {
	pk, ok := s.GetCurrentKey(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return common.Address{}, fmt.Errorf("key not found")
	}

	fromAddress := crypto.PubkeyToAddress(pk.Value)
	return fromAddress, nil
}
