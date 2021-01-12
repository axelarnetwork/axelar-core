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

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	QueryMasterKey = "masterkey"
	CreateDeployTx = "deploy"
	CreateMintTx   = "mint"
	SendTx         = "send"
)

func NewQuerier(rpc types.RPCClient, k Keeper, s types.Signer) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryMasterKey:
			return queryMasterAddress(ctx, s)
		case CreateDeployTx:
			return createDeployTx(ctx, rpc, s, req.Data)
		case CreateMintTx:
			return createMintTx(ctx, k, s, rpc, req.Data)
		case SendTx:
			return sendTx(ctx, k, rpc, s, path[1])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown eth-bridge query endpoint: %s", path[0]))
		}
	}
}

func queryMasterAddress(ctx sdk.Context, s ethTypes.Signer) ([]byte, error) {

	pk, ok := s.GetCurrentMasterKey(ctx, balance.Ethereum)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	fromAddress := crypto.PubkeyToAddress(pk)

	bz := fromAddress.Bytes()

	return bz, nil
}

/*
  Create a transaction for smart contract deployment. See:

  https://goethereumbook.org/en/smart-contract-deploy/
  https://gist.github.com/tomconte/6ce22128b15ba36bb3d7585d5180fba0

  If gasLimit is set to 0, the function will attempt to estimate the amount of gas needed
*/
func createDeployTx(ctx sdk.Context, rpc types.RPCClient, s types.Signer, data []byte) ([]byte, error) {
	var params types.DeployParams
	err := types.ModuleCdc.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, err
	}

	contractOwner, err := getContractOwner(ctx, s)
	if err != nil {
		return nil, err
	}
	nonce, err := rpc.PendingNonceAt(context.Background(), contractOwner)
	if err != nil {
		return nil, fmt.Errorf("could not create nonce: %s", err)
	}

	gasPrice, err := rpc.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not calculate gas price: %s", err)
	}

	if params.GasLimit == 0 {
		params.GasLimit, err = rpc.EstimateGas(context.Background(), ethereumRoot.CallMsg{
			To:   nil,
			Data: params.ByteCode,
		})

		if err != nil {
			return nil, fmt.Errorf("could not estimate gas limit: %s", err)
		}
	}

	value := big.NewInt(0)

	tx := ethTypes.NewContractCreation(nonce, value, params.GasLimit, gasPrice, params.ByteCode)

	return types.ModuleCdc.MustMarshalJSON(tx), nil
}

/*
  Create a transaction to mint tokens for a ERC20 smart contract. See:

  https://medium.com/swlh/understanding-data-payloads-in-ethereum-transactions-354dbe995371
  https://medium.com/mycrypto/why-do-we-need-transaction-data-39c922930e92
  https://goethereumbook.org/en/transfer-tokens/

  If gasLimit is set to 0, the function will attempt to estimate the amount of gas needed
*/
func createMintTx(ctx sdk.Context, k Keeper, s types.Signer, rpc types.RPCClient, data []byte) ([]byte, error) {
	var params types.MintParams
	err := types.ModuleCdc.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, err
	}

	contractOwner, err := getContractOwner(ctx, s)
	if err != nil {
		return nil, err
	}

	hash, ok := k.GetTxIDForContractID(ctx, params.ContractID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "selected contract's deployment is not verified")
	}

	receipt, err := rpc.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "could not obtain receipt: %v", err)
	}

	nonce, err := rpc.PendingNonceAt(context.Background(), contractOwner)
	if err != nil {
		return nil, fmt.Errorf("could not create nonce: %s", err)
	}

	callData := types.CreateMintCallData(params.Recipient, params.Amount.BigInt())

	if params.GasLimit == 0 {

		params.GasLimit, err = rpc.EstimateGas(context.Background(), ethereumRoot.CallMsg{
			To:   &receipt.ContractAddress,
			Data: callData,
		})
		if err != nil {
			return nil, fmt.Errorf("could not estimate gas limit: %s", err)
		}
	}

	gasPrice, err := rpc.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not calculate gas price: %s", err)
	}

	tx := ethTypes.NewTransaction(nonce, receipt.ContractAddress, big.NewInt(0), params.GasLimit, gasPrice, callData)
	return types.ModuleCdc.MustMarshalJSON(tx), nil
}

func getContractOwner(ctx sdk.Context, s types.Signer) (common.Address, error) {
	pk, ok := s.GetCurrentMasterKey(ctx, balance.Ethereum)
	if !ok {
		return common.Address{}, fmt.Errorf("key not found")
	}

	fromAddress := crypto.PubkeyToAddress(pk)
	return fromAddress, nil
}

func sendTx(ctx sdk.Context, k Keeper, rpc types.RPCClient, s types.Signer, txID string) ([]byte, error) {
	h, err := k.GetHashToSign(ctx, txID)
	if err != nil {
		return nil, err
	}
	sigID := h.String()
	pk, ok := s.GetKeyForSigID(ctx, sigID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding key for sig ID %s", sigID))
	}

	sig, ok := s.GetSig(ctx, sigID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding signature for sig ID %s", sigID))
	}

	signedTx, err := k.SignRawTransaction(ctx, txID, sig, pk)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not insert generated signature: %v", err))
	}

	err = rpc.SendTransaction(context.Background(), signedTx)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
	}

	return k.Codec().MustMarshalJSON(fmt.Sprintf("successfully sent transaction %s to Ethereum", signedTx.Hash().String())), nil
}
