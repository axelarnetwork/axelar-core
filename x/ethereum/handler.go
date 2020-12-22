package eth_bridge

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"

	ethereumRoot "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	ethereum = "ethereum"
	gasLimit = uint64(3000000)
)

func NewHandler(k keeper.Keeper, rpc types.RPCClient, v types.Voter, s types.Signer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {

		case *types.MsgVoteVerifiedTx:
			return handleMsgVoteVerifiedTx(ctx, v, *msg)

		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, rpc, v, msg)

		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgRawTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, rpc types.RPCClient, s types.Signer, msg types.MsgRawTx) (*sdk.Result, error) {
	txId := msg.TxHash.String()

	poll := exported.PollMeta{ID: txId, Module: types.ModuleName, Type: types.MsgVerifyTx{}.Type()}

	if isVerified(ctx, v, poll) {
		return nil, fmt.Errorf("transaction not verified")
	}

	tx, err := createTransaction(ctx, rpc, s, msg)

	if err != nil {
		return nil, fmt.Errorf("Could not create ethereum transaction: %s", err)
	}

	k.SetRawTx(ctx, txId, tx)

	// Print out the hash that becomes the input for the threshold signing
	chainID, err := rpc.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Could not retrieve ethereum network: %s", err)
	}
	signer := ethTypes.NewEIP155Signer(chainID)
	hash := signer.Hash(tx).Bytes()

	k.Logger(ctx).Info(fmt.Sprintf("ethereum tx to sign: %s", k.Codec().MustMarshalJSON(hash)))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxId, txId),
			sdk.NewAttribute(types.AttributeAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeDestination, msg.Destination.String()),
		),
	)

	return &sdk.Result{
		Data:   hash,
		Log:    fmt.Sprintf("successfully created withdraw transaction for Ethereum. Hash to sign: %s", k.Codec().MustMarshalJSON(hash)),
		Events: ctx.EventManager().Events(),
	}, nil
}

// This can be used as a potential hook to immediately act on a poll being decided by the vote
func handleMsgVoteVerifiedTx(ctx sdk.Context, v types.Voter, msg types.MsgVoteVerifiedTx) (*sdk.Result, error) {
	if err := v.TallyVote(ctx, &msg); err != nil {
		return nil, err
	}
	return nil, nil
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, v types.Voter, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying ethereum transaction")
	txId := msg.TX.Hash.String()

	poll := exported.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: txId}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthBridge, "could not initialize new poll")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxId, txId),
			sdk.NewAttribute(types.AttributeAmount, msg.TX.Amount.String()),
		),
	)

	k.SetTX(ctx, txId, msg.TX)

	/*
	 Anyone not able to verify the transaction will automatically record a negative vote,
	 but only validators will later send out that vote.
	*/

	if err := verifyTx(rpc, msg.TX, k.GetConfirmationHeight(ctx)); err != nil {

		if err := v.Vote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "voting failed").Error())
			return &sdk.Result{
				Log:    err.Error(),
				Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
				Events: ctx.EventManager().Events(),
			}, nil
		}

		k.Logger(ctx).Debug(sdkerrors.Wrapf(err,
			"expected transaction (%s) could not be verified", txId).Error())
		return &sdk.Result{
			Log:    err.Error(),
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}, nil

	} else {

		if err := v.Vote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "voting failed").Error())
			return &sdk.Result{
				Log:    err.Error(),
				Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
				Events: ctx.EventManager().Events(),
			}, nil
		}

		return &sdk.Result{
			Log:    "successfully verified transaction",
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(true),
			Events: ctx.EventManager().Events(),
		}, nil
	}

}

func isVerified(ctx sdk.Context, v types.Voter, poll exported.PollMeta) bool {
	res := v.Result(ctx, poll)
	return res == nil || !res.(bool)
}

func verifyTx(rpc types.RPCClient, tx types.TX, expectedConfirmationHeight uint64) error {

	//TODO: parallelise all 3 RPC calls
	actualTx, pending, err := rpc.TransactionByHash(context.Background(), *tx.Hash)

	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum transaction")
	}

	if pending {
		return fmt.Errorf("Transaction is pending")
	}

	if actualTx.To().String() != tx.Address.String() {
		return fmt.Errorf("expected destination address does not match actual destination address")
	}

	if !bytes.Equal(actualTx.Value().Bytes(), tx.Amount.Bytes()) {
		return fmt.Errorf("expected amount does not match actual amount")
	}

	receipt, err := rpc.TransactionReceipt(context.Background(), *tx.Hash)

	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum receipt")
	}

	number, err := rpc.BlockNumber(context.Background())

	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum block number")
	}

	if (number - receipt.BlockNumber.Uint64()) < expectedConfirmationHeight {
		return fmt.Errorf("not enough confirmations yet")
	}

	return nil
}

/*
	Creating an Ethereum transaction with eth client. See:

	https://medium.com/coinmonks/web3-go-part-1-31c68c68e20e
	https://goethereumbook.org/en/transaction-raw-create/
*/
func createTransaction(ctx sdk.Context, rpc types.RPCClient, s types.Signer, msg types.MsgRawTx) (*ethTypes.Transaction, error) {

	//TODO: Add support to specify a key other than the master key
	pk, ok := s.GetCurrentMasterKey(ctx, ethereum)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthBridge, "key not found")
	}

	fromAddress := crypto.PubkeyToAddress(pk)
	nonce, err := rpc.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("Could not create nonce: %s", err)
	}

	toAddress := msg.Destination.Convert()

	gasPrice, err := rpc.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Could not calculate gas price: %s", err)
	}

	switch msg.TXType {

	case types.TypeETH:

		return ethTypes.NewTransaction(nonce, toAddress, &msg.Amount, gasLimit, gasPrice, make([]byte, 0)), nil

	case types.TypeERC20mint:

		// TODO: correctly look up the contract address once the deploy contract functionality is done
		return createMintTransaction(rpc, fromAddress, toAddress, toAddress, gasLimit, &msg.Amount)

	default:

		return nil, fmt.Errorf("Unsuported transaction type: %s", err)

	}

}

/*
  Create a transaction for smart contract deployment. See:

  https://goethereumbook.org/en/smart-contract-deploy/
  https://gist.github.com/tomconte/6ce22128b15ba36bb3d7585d5180fba0

  If gasLimit is set to 0, the function will attempt to estimate the amount of gas needed
*/
func createDeploySCTransaction(rpc types.RPCClient, fromAddr common.Address, gasLimit uint64, byteCode []byte) (*ethTypes.Transaction, error) {

	nonce, err := rpc.PendingNonceAt(context.Background(), fromAddr)
	if err != nil {
		return nil, fmt.Errorf("Could not create nonce: %s", err)
	}

	gasPrice, err := rpc.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Could not calculate gas price: %s", err)
	}

	if gasLimit == 0 {

		gasLimit, err = rpc.EstimateGas(context.Background(), ethereumRoot.CallMsg{
			To:   nil,
			Data: byteCode,
		})

		if err != nil {
			return nil, fmt.Errorf("Could not estimate gas limit: %s", err)
		}
	}

	value := big.NewInt(0)

	return ethTypes.NewContractCreation(nonce, value, gasLimit, gasPrice, byteCode), nil

}

/*
  Create a transaction to mint tokens for a ERC20 smart contract. See:

  https://medium.com/swlh/understanding-data-payloads-in-ethereum-transactions-354dbe995371
  https://medium.com/mycrypto/why-do-we-need-transaction-data-39c922930e92
  https://goethereumbook.org/en/transfer-tokens/

  If gasLimit is set to 0, the function will attempt to estimate the amount of gas needed
*/
func createMintTransaction(rpc types.RPCClient, fromAddr, contractAddr, toAddr common.Address, gasLimit uint64, amount *big.Int) (*ethTypes.Transaction, error) {

	paddedAddr := hexutil.Encode(common.LeftPadBytes(toAddr.Bytes(), 32))
	paddedVal := hexutil.Encode(common.LeftPadBytes(amount.Bytes(), 32))

	var data []byte

	data = append(data, common.FromHex(types.ERC20MintSel)...)
	data = append(data, common.FromHex(paddedAddr)...)
	data = append(data, common.FromHex(paddedVal)...)

	nonce, err := rpc.PendingNonceAt(context.Background(), fromAddr)
	if err != nil {
		return nil, fmt.Errorf("Could not create nonce: %s", err)
	}

	value := big.NewInt(0)

	gasPrice, err := rpc.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Could not calculate gas price: %s", err)
	}

	if gasLimit == 0 {

		gasLimit, err = rpc.EstimateGas(context.Background(), ethereumRoot.CallMsg{
			To:   &contractAddr,
			Data: data,
		})

		if err != nil {
			return nil, fmt.Errorf("Could not estimate gas limit: %s", err)
		}
	}

	return ethTypes.NewTransaction(nonce, contractAddr, value, gasLimit, gasPrice, data), nil

}
