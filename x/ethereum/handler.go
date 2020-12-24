package ethereum

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
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
	gasLimit = uint64(3000000)
)

func NewHandler(k keeper.Keeper, rpc types.RPCClient, v types.Voter, s types.Signer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {

		case types.MsgInstallSC:
			return handleMsgInstallSC(ctx, k, msg)

		case types.MsgVoteVerifiedTx:
			return handleMsgVoteVerifiedTx(ctx, v, msg)

		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, s, rpc, v, msg)

		case types.MsgRawTx:
			return handleMsgRawTx(ctx, k, v, rpc, s, msg)

		case types.MsgSendTx:

			return handleMsgSendTx(ctx, k, rpc, s, msg)

		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgSendTx(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, s types.Signer, msg types.MsgSendTx) (*sdk.Result, error) {

	pk, ok := s.GetKeyForSigID(ctx, msg.SignatureID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding key for sig ID %s", msg.SignatureID))
	}

	rawTx := k.GetRawTx(ctx, msg.TxID)
	if rawTx == nil {
		return nil, fmt.Errorf("raw tx for ID %s has not been prepared yet", msg.TxID)
	}

	networkID, err := rpc.NetworkID(context.Background())
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not obtain network ID %v", err))
	}

	sig, ok := s.GetSig(ctx, msg.SignatureID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding signature for sig ID %s", msg.SignatureID))
	}

	signer := ethTypes.NewEIP155Signer(networkID)
	hash := signer.Hash(rawTx).Bytes()

	recoverableSig, err := encodeSig(hash, pk, sig.R, sig.S)

	signedTx, err := rawTx.WithSignature(signer, recoverableSig)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not insert generated signature: %v", err))
	}

	err = rpc.SendTransaction(context.Background(), signedTx)

	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not send transaction to network : %v", err))
	}

	hash = signedTx.Hash().Bytes()

	return &sdk.Result{Data: hash, Log: fmt.Sprintf("successfully sent transaction %s to Ethereum", hash), Events: ctx.EventManager().Events()}, nil

}

func handleMsgInstallSC(ctx sdk.Context, k keeper.Keeper, msg types.MsgInstallSC) (*sdk.Result, error) {

	k.SetSmartContract(ctx, msg.SmartContractID, msg.Bytecode)

	str := fmt.Sprintf("successfully installed smart contract for Ethereum with ID '%s'", msg.SmartContractID)

	k.Logger(ctx).Info(str)

	return &sdk.Result{
		Data:   k.Codec().MustMarshalBinaryLengthPrefixed(true),
		Log:    str,
		Events: ctx.EventManager().Events(),
	}, nil

}

func handleMsgRawTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, rpc types.RPCClient, s types.Signer, msg types.MsgRawTx) (*sdk.Result, error) {

	tx, err := createTransaction(ctx, rpc, s, k, v, msg)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not create ethereum transaction: %s", err))
	}

	txId := tx.Hash().String()
	k.SetRawTx(ctx, txId, tx)
	k.Logger(ctx).Info(fmt.Sprintf("storing tx %s", txId))

	// Print out the hash that becomes the input for the threshold signing
	networkID, err := rpc.NetworkID(context.Background())
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not retrieve ethereum network: %s", err))
	}
	signer := ethTypes.NewEIP155Signer(networkID)
	hash := signer.Hash(tx).Bytes()

	k.Logger(ctx).Info(fmt.Sprintf("ethereum tx [%s] to sign: %s", txId, k.Codec().MustMarshalJSON(hash)))
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

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, s types.Signer, rpc types.RPCClient, v types.Voter, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying ethereum transaction")
	txId := msg.Tx.Hash.String()

	mk, ok := s.GetCurrentMasterKey(ctx, balance.Ethereum)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "master key not found")
	}

	poll := exported.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: txId}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "could not initialize new poll")
	}

	if msg.TxType == types.TypeSCDeploy {
		k.SetTxIDForContractID(ctx, msg.Tx.ContractID, msg.Tx.Network, msg.Tx.Hash)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxId, txId),
			sdk.NewAttribute(types.AttributeAmount, msg.Tx.Amount.String()),
		),
	)

	k.SetTX(ctx, txId, msg.Tx)

	/*
	 Anyone not able to verify the transaction will automatically record a negative vote,
	 but only validators will later send out that vote.
	*/

	if err := verifyTx(ctx, rpc, k, msg, crypto.PubkeyToAddress(mk)); err != nil {
		k.Logger(ctx).Debug(sdkerrors.Wrapf(err, "expected transaction (%s) could not be verified", txId).Error())
		if err := v.RecordVote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "voting failed").Error())
			return &sdk.Result{
				Log:    err.Error(),
				Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
				Events: ctx.EventManager().Events(),
			}, nil
		}

		return &sdk.Result{
			Log:    err.Error(),
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}, nil

	} else {
		if err := v.RecordVote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true}); err != nil {
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

func verifyTx(ctx sdk.Context, rpc types.RPCClient, k keeper.Keeper, msg types.MsgVerifyTx, expectedSender common.Address) error {
	actualTx, pending, err := rpc.TransactionByHash(context.Background(), msg.Tx.Hash)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum transaction")
	}

	if pending {
		return fmt.Errorf("transaction is pending")
	}

	sender, err := ethTypes.Sender(ethTypes.NewEIP155Signer(msg.Tx.Network.Params().ChainID), actualTx)
	if err != nil {
		return fmt.Errorf("could not derive sender")
	}

	if sender != expectedSender {
		return fmt.Errorf("sender does not match")
	}

	receipt, err := rpc.TransactionReceipt(context.Background(), msg.Tx.Hash)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum receipt")
	}

	blockNumber, err := rpc.BlockNumber(context.Background())
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum block number")
	}

	if (blockNumber - receipt.BlockNumber.Uint64()) < k.GetConfirmationHeight(ctx) {
		return fmt.Errorf("not enough confirmations yet")
	}
	switch msg.TxType {
	case types.TypeERC20mint:

		if !bytes.Equal(actualTx.Data(), createMintCallData(msg.Tx.Destination, msg.Tx.Amount.BigInt())) {
			return fmt.Errorf("mint call mismatch")
		}

		return nil
	case types.TypeSCDeploy:
		if !bytes.Equal(actualTx.Data(), k.GetSmartContract(ctx, msg.Tx.ContractID)) {
			return fmt.Errorf("smart contract byte code mismatch")
		}
		return nil
	default:
		return fmt.Errorf("unknown tx type")
	}
}

/*
	Creating an Ethereum transaction with eth client. See:

	https://medium.com/coinmonks/web3-go-part-1-31c68c68e20e
	https://goethereumbook.org/en/transaction-raw-create/
*/
func createTransaction(ctx sdk.Context, rpc types.RPCClient, s types.Signer, k keeper.Keeper, v types.Voter, msg types.MsgRawTx) (*ethTypes.Transaction, error) {

	pk, ok := s.GetCurrentMasterKey(ctx, balance.Ethereum)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	fromAddress := crypto.PubkeyToAddress(pk)

	switch msg.TXType {

	case types.TypeERC20mint:

		hash, err := verifyContract(ctx, k, v, msg.ContractID, msg.Network)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrEthereum, "could not verify contract transaction: %v", err)
		}

		receipt, err := rpc.TransactionReceipt(context.Background(), *hash)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrEthereum, "could not obtain receipt: %v", err)
		}

		contractAddress := receipt.ContractAddress
		return createMintTransaction(rpc, fromAddress, contractAddress, msg.Destination, gasLimit, msg.Amount.BigInt())

	case types.TypeSCDeploy:

		byteCodes := k.GetSmartContract(ctx, msg.ContractID)

		return createDeploySCTransaction(rpc, fromAddress, gasLimit, byteCodes)

	default:

		return nil, fmt.Errorf("unknown tx type")
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

	data := createMintCallData(toAddr, amount)

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

func createMintCallData(toAddr common.Address, amount *big.Int) []byte {
	paddedAddr := hexutil.Encode(common.LeftPadBytes(toAddr.Bytes(), 32))
	paddedVal := hexutil.Encode(common.LeftPadBytes(amount.Bytes(), 32))

	var data []byte

	data = append(data, common.FromHex(types.ERC20MintSel)...)
	data = append(data, common.FromHex(paddedAddr)...)
	data = append(data, common.FromHex(paddedVal)...)
	return data
}

func verifyContract(ctx sdk.Context, k keeper.Keeper, v types.Voter, contractID string, networkID types.Network) (*common.Hash, error) {
	verifiedTxID, ok := k.GetTxIDForContractID(ctx, contractID, networkID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "contract ID unknown")
	}

	if isVerified(ctx, v, exported.PollMeta{Module: types.ModuleName, Type: types.MsgVerifyTx{}.Type(), ID: verifiedTxID.String()}) {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("contract not deployed yet"))
	}

	return &verifiedTxID, nil
}

func encodeSig(hash []byte, expectedPubKey ecdsa.PublicKey, R, S *big.Int) ([]byte, error) {

	var sig []byte
	var err error
	var pubkey *ecdsa.PublicKey

	sig = append(sig, common.LeftPadBytes(R.Bytes(), 32)...)
	sig = append(sig, common.LeftPadBytes(S.Bytes(), 32)...)
	sig = append(sig, make([]byte, 1)...)
	sig[64] = 0

	if pubkey, err = crypto.SigToPub(hash, sig); err != nil {

		return nil, err
	}

	if bytes.Equal(expectedPubKey.Y.Bytes(), pubkey.Y.Bytes()) {

		return sig, nil
	}

	sig[64] = 1

	return sig, nil

}
