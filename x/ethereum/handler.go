package ethereum

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils/denom"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/ethereum/go-ethereum/crypto"
)

// NewHandler returns the handler of the ethereum module
func NewHandler(k keeper.Keeper, rpc types.RPCClient, v types.Voter, s types.Signer, snap snapshot.Snapshotter, b types.Balancer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgLink:
			return handleMsgLink(ctx, k, b, msg)
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, rpc, v, msg)
		case *types.MsgVoteVerifiedTx:
			return handleMsgVoteVerifiedTx(ctx, k, v, b, msg)
		case types.MsgSignDeployToken:
			return handleMsgSignDeployToken(ctx, k, s, snap, msg)
		case types.MsgSignTx:
			return handleMsgSignTx(ctx, k, s, snap, msg)
		case types.MsgSignPendingTransfers:
			return handleMsgSignPendingTransfersTx(ctx, k, s, snap, b, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgLink(ctx sdk.Context, k keeper.Keeper, b types.Balancer, msg types.MsgLink) (*sdk.Result, error) {
	burnerAddr, salt, err := k.GetBurnerAddressAndSalt(ctx, msg.Symbol, msg.Recipient.Address, common.HexToAddress(msg.GatewayAddr))
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	err = b.LinkAddresses(ctx, balance.CrossChainAddress{Chain: balance.Ethereum, Address: burnerAddr.String()}, msg.Recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	burnerInfo := types.BurnerInfo{
		Symbol: msg.Symbol,
		Salt:   salt,
	}
	k.SetBurnerInfo(ctx, burnerAddr, &burnerInfo)

	logMsg := fmt.Sprintf("successfully linked {%s} and {%s}", burnerAddr.String(), msg.Recipient.String())
	k.Logger(ctx).Info(logMsg)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, burnerAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Recipient.String()),
		),
	)

	return &sdk.Result{
		Data:   []byte(burnerAddr.String()),
		Log:    logMsg,
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgSignPendingTransfersTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snap snapshot.Snapshotter, balancer types.Balancer, msg types.MsgSignPendingTransfers) (*sdk.Result, error) {
	pendingTransfers := balancer.GetPendingTransfersForChain(ctx, balance.Ethereum)

	if len(pendingTransfers) == 0 {
		return &sdk.Result{
			Data:   nil,
			Log:    fmt.Sprintf("no pending transfer for chain %s found", balance.Ethereum),
			Events: ctx.EventManager().Events(),
		}, nil
	}

	chainID := k.GetNetwork(ctx).Params().ChainID

	data, err := types.CreateMintCommandData(chainID, pendingTransfers)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256(data)[:32])

	keyID, ok := signer.GetCurrentMasterKeyID(ctx, balance.Ethereum)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no master key for chain %s found", balance.Ethereum)
	}

	s, ok := snap.GetLatestSnapshot(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "no snapshot found")
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for mint command %s", commandIDHex))
	k.SetCommandData(ctx, commandID, data)

	k.Logger(ctx).Info(fmt.Sprintf("signing mint command [%s] for pending transfers to chain %s", commandIDHex, balance.Ethereum))
	signHash := types.GetEthereumSignHash(data)

	err = signer.StartSign(ctx, keyID, commandIDHex, signHash.Bytes(), s.Validators)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeCommandID, commandIDHex),
		),
	)

	return &sdk.Result{
		Data:   commandID[:],
		Log:    fmt.Sprintf("successfully started signing protocol for %s pending transfers, commandID: %s", balance.Ethereum, commandIDHex),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, v types.Voter, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying ethereum transaction")
	txHash := common.BytesToHash(msg.TxInfo.TxHash).String()

	poll := exported.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: txHash}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxID, txHash),
		),
	)

	k.SetUnverifiedTxInfo(ctx, &msg.TxInfo)

	/*
	 Anyone not able to verify the transaction will automatically record a negative vote,
	 but only validators will later send out that vote.
	*/

	if err := verifyTx(rpc, msg.TxInfo, k.GetRequiredConfirmationHeight(ctx)); err != nil {
		k.Logger(ctx).Debug(sdkerrors.Wrapf(err, "expected transaction (%s) could not be verified", txHash).Error())
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
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, b types.Balancer, msg *types.MsgVoteVerifiedTx) (*sdk.Result, error) {
	event := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		sdk.NewAttribute(types.AttributePoll, msg.PollMeta.String()),
		sdk.NewAttribute(types.AttributeVotingData, strconv.FormatBool(msg.VotingData)),
	)

	if err := v.TallyVote(ctx, msg); err != nil {
		return nil, err
	}

	if confirmed := v.Result(ctx, msg.Poll()); confirmed != nil {
		k.ProcessVerificationResult(ctx, msg.PollMeta.ID, confirmed.(bool))
		v.DeletePoll(ctx, msg.Poll())

		info, ok := k.GetVerifiedTxInfo(ctx, msg.PollMeta.ID)
		if !ok {
			return nil, fmt.Errorf("transaction not verified")
		}

		if types.IsERC20Transfer(info.Data) {

			addr, value, err := types.UnpackERC20Transfer(info.Data)
			if err != nil {
				return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
			}
			burnAddr := balance.CrossChainAddress{Address: addr, Chain: balance.Ethereum}
			amount := sdk.NewInt64Coin(denom.Satoshi, value.Int64())

			err = b.EnqueueForTransfer(ctx, burnAddr, amount)
			if err == nil {
				k.Logger(ctx).Debug(fmt.Sprintf("Transfer of %s from %s in %s successfully prepared",
					amount.Amount.String(), burnAddr.Address, burnAddr.Chain.String()))
			} else {
				k.Logger(ctx).Debug(fmt.Sprintf("prepared no transfer: %s", err))
			}
		}

		event = event.AppendAttributes(sdk.NewAttribute(types.AttributePollConfirmed, strconv.FormatBool(confirmed.(bool))))
	}

	ctx.EventManager().EmitEvent(event)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSignDeployToken(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snap snapshot.Snapshotter, msg types.MsgSignDeployToken) (*sdk.Result, error) {
	chainID := k.GetParams(ctx).Network.Params().ChainID

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256([]byte(msg.TokenName))[:32])

	data, err := types.CreateDeployTokenCommandData(chainID, commandID, msg.TokenName, msg.Symbol, msg.Decimals, msg.Capacity)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	keyID, ok := signer.GetCurrentMasterKeyID(ctx, balance.Ethereum)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no master key for chain %s found", balance.Ethereum)
	}

	s, ok := snap.GetLatestSnapshot(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "no snapshot found")
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for deploy-token command %s", commandIDHex))
	k.SetCommandData(ctx, commandID, data)

	signHash := types.GetEthereumSignHash(data)

	err = signer.StartSign(ctx, keyID, commandIDHex, signHash.Bytes(), s.Validators)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	k.SaveTokenInfo(ctx, msg)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeCommandID, commandIDHex),
		),
	)

	return &sdk.Result{
		Data:   commandID[:],
		Log:    fmt.Sprintf("successfully started signing protocol for deploy-token command %s", commandIDHex),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgSignTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snap snapshot.Snapshotter, msg types.MsgSignTx) (*sdk.Result, error) {
	tx := msg.UnmarshaledTx()
	txID := tx.Hash().String()
	k.SetRawTx(ctx, txID, tx)
	k.Logger(ctx).Info(fmt.Sprintf("storing raw tx %s", txID))
	hash, err := k.GetHashToSign(ctx, txID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	k.Logger(ctx).Info(fmt.Sprintf("ethereum tx [%s] to sign: %s", txID, k.Codec().MustMarshalJSON(hash)))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxID, txID),
		),
	)

	keyID, ok := signer.GetCurrentMasterKeyID(ctx, balance.Ethereum)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no master key for chain %s found", balance.Ethereum)
	}

	s, ok := snap.GetLatestSnapshot(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "no snapshot found")
	}
	err = signer.StartSign(ctx, keyID, txID, hash.Bytes(), s.Validators)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	return &sdk.Result{
		Data:   []byte(txID),
		Log:    fmt.Sprintf("successfully started signing protocol for transaction with ID %s.", txID),
		Events: ctx.EventManager().Events(),
	}, nil
}

func verifyTx(rpc types.RPCClient, expectedInfo types.TransactionInfo, requiredConfirmations uint64) error {

	actualInfo, err := rpc.GetTransactionInfo(context.Background(), common.BytesToHash(expectedInfo.TxHash))
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum transaction")
	}

	if actualInfo.To != expectedInfo.To {
		return fmt.Errorf("expected destination address does not match actual destination address")
	}

	if !actualInfo.Value.Equal(expectedInfo.Value) {
		return fmt.Errorf("expected value does not match actual value")
	}

	if !bytes.Equal(actualInfo.Data, expectedInfo.Data) {
		return fmt.Errorf("expected data does not match actual data")
	}

	if actualInfo.Confirmations.Uint64() < requiredConfirmations {
		return fmt.Errorf("not enough confirmations yet")
	}

	return nil
}
