package ethereum

import (
	"bytes"
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"

	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

func NewHandler(k keeper.Keeper, rpc types.RPCClient, v types.Voter, s types.Signer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, rpc, v, msg)
		case *types.MsgVoteVerifiedTx:
			return handleMsgVoteVerifiedTx(ctx, k, v, msg)
		case types.MsgRawTx:
			return handleMsgRawTx(ctx, k, msg)
		case types.MsgSendTx:
			return handleMsgSendTx(ctx, k, rpc, s, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, v types.Voter, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying ethereum transaction")
	txID := msg.Tx.Hash().String()

	poll := exported.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: txID}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxID, txID),
		),
	)

	k.SetUnverifiedTx(ctx, txID, msg.Tx)

	/*
	 Anyone not able to verify the transaction will automatically record a negative vote,
	 but only validators will later send out that vote.
	*/

	if err := verifyTx(ctx, k, rpc, msg.Tx); err != nil {
		k.Logger(ctx).Debug(sdkerrors.Wrapf(err, "expected transaction (%s) could not be verified", txID).Error())
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

// This can be used as a potential hook to immediately act on a poll being decided by the vote
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg *types.MsgVoteVerifiedTx) (*sdk.Result, error) {
	if err := v.TallyVote(ctx, msg); err != nil {
		return nil, err
	}

	if confirmed := v.Result(ctx, msg.Poll()); confirmed != nil {
		if err := k.ProcessVerificationResult(ctx, msg.PollMeta.ID, confirmed.(bool)); err != nil {

			return nil, err
		}

		v.DeletePoll(ctx, msg.Poll())
	}
	return &sdk.Result{}, nil
}

func handleMsgRawTx(ctx sdk.Context, k keeper.Keeper, msg types.MsgRawTx) (*sdk.Result, error) {
	txID := msg.Tx.Hash().String()
	k.SetRawTx(ctx, txID, msg.Tx)
	k.Logger(ctx).Info(fmt.Sprintf("storing raw tx %s", txID))
	hash, err := k.GetHashToSign(ctx, txID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	json := k.Codec().MustMarshalJSON(hash)
	k.Logger(ctx).Info(fmt.Sprintf("ethereum tx [%s] to sign: %s", txID, json))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxID, txID),
		),
	)

	return &sdk.Result{
		Data:   json,
		Log:    fmt.Sprintf("successfully created withdraw transaction for Ethereum. Hash to sign: %s", json),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgSendTx(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, s types.Signer, msg types.MsgSendTx) (*sdk.Result, error) {
	pk, ok := s.GetKeyForSigID(ctx, msg.SignatureID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding key for sig ID %s", msg.SignatureID))
	}

	sig, ok := s.GetSig(ctx, msg.SignatureID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not find a corresponding signature for sig ID %s", msg.SignatureID))
	}

	signedTx, err := k.SignRawTransaction(ctx, msg.TxID, sig, pk)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not insert generated signature: %v", err))
	}

	err = rpc.SendTransaction(context.Background(), signedTx)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
	}

	json := k.Codec().MustMarshalJSON(signedTx)
	return &sdk.Result{
		Data:   json,
		Log:    fmt.Sprintf("successfully sent transaction %s to Ethereum", json),
		Events: ctx.EventManager().Events(),
	}, nil

}

func verifyTx(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, expectedTx *ethTypes.Transaction) error {
	hash := expectedTx.Hash()
	actualTx, pending, err := rpc.TransactionByHash(context.Background(), hash)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum transaction")
	}

	if pending {
		return fmt.Errorf("transaction is pending")
	}

	if !bytes.Equal(actualTx.Data(), expectedTx.Data()) {
		return fmt.Errorf("tx smart contract data not as expected")
	}

	actualVal := actualTx.Value()
	if actualVal == nil || actualVal.Cmp(expectedTx.Value()) != 0 {
		return fmt.Errorf("expected tx value %d, got %d", expectedTx.Value(), actualVal)
	}

	receipt, err := rpc.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum receipt")
	}

	blockNumber, err := rpc.BlockNumber(context.Background())
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum block number")
	}

	if (blockNumber - receipt.BlockNumber.Uint64()) < k.GetRequiredConfirmationHeight(ctx) {
		return fmt.Errorf("not enough confirmations yet")
	}
	return nil
}
