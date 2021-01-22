package bitcoin

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils/denom"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewHandler creates an sdk.Handler for all bitcoin type messages
func NewHandler(k keeper.Keeper, v types.Voter, rpc types.RPCClient, signer types.Signer, snap types.Snapshotter, b types.Balancer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgTrack:
			return handleMsgTrack(ctx, k, rpc, msg)
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, v, rpc, msg)
		case *types.MsgVoteVerifiedTx:
			return handleMsgVoteVerifiedTx(ctx, k, v, b, msg)
		case types.MsgSignTx:
			return handleMsgSignTx(ctx, k, signer, snap, msg)
		case types.MsgLink:
			return handleMsgLink(ctx, k, signer, b, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgLink(ctx sdk.Context, k keeper.Keeper, s types.Signer, b types.Balancer, msg types.MsgLink) (*sdk.Result, error) {

	key, ok := s.GetCurrentMasterKey(ctx, balance.Bitcoin)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "master key not set")
	}

	btcAddr, script, err := k.GenerateDepositAddressAndRedeemScript(ctx, btcec.PublicKey(key), msg.Recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())

	}
	k.SetRedeemScript(ctx, btcAddr, script)

	b.LinkAddresses(ctx, balance.CrossChainAddress{Chain: balance.Bitcoin, Address: btcAddr.EncodeAddress()}, msg.Recipient)

	id, ok := s.GetCurrentMasterKeyID(ctx, balance.Bitcoin)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "master key not set")
	}
	k.SetKeyIDByAddress(ctx, btcAddr.EncodeAddress(), id)

	logMsg := fmt.Sprintf("successfully linked {%s} and {%s}", btcAddr.EncodeAddress(), msg.Recipient.String())
	k.Logger(ctx).Info(logMsg)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, btcAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Recipient.String()),
		),
	)

	return &sdk.Result{
		Data:   []byte(btcAddr.EncodeAddress()),
		Log:    logMsg,
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgTrack(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, msg types.MsgTrack) (*sdk.Result, error) {
	k.Logger(ctx).Debug(fmt.Sprintf("start tracking address %v", msg.Address))
	trackAddress(ctx, k, rpc, msg.Address, msg.Rescan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Address),
		),
	)

	return &sdk.Result{
		Data:   []byte(msg.Address),
		Log:    fmt.Sprintf("successfully tracked address %s", msg.Address),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, rpc types.RPCClient, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying bitcoin transaction")

	txID := msg.OutPointInfo.OutPoint.Hash.String()
	poll := exported.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: msg.OutPointInfo.OutPoint.String()}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxID, txID),
			sdk.NewAttribute(types.AttributeAmount, msg.OutPointInfo.Amount.String()),
		),
	)

	// store outpoint for later reference
	k.SetUnverifiedOutpointInfo(ctx, msg.OutPointInfo)

	/*
	 Anyone not able to verify the transaction will automatically record a negative vote,
	 but only validators will later send out that vote.
	*/

	err := verifyTx(rpc, msg.OutPointInfo, k.GetRequiredConfirmationHeight(ctx))
	switch err {
	// verification successful
	case nil:
		if err := v.RecordVote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "voting failed").Error())
			return &sdk.Result{Log: err.Error(), Events: ctx.EventManager().Events()}, nil
		}

		k.Logger(ctx).Debug(fmt.Sprintf("transaction (%s) was verified", txID))
		return &sdk.Result{Log: "successfully verified transaction", Events: ctx.EventManager().Events()}, nil
	// verification unsuccessful
	default:
		if err := v.RecordVote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "voting failed").Error())
			return &sdk.Result{Log: err.Error(), Events: ctx.EventManager().Events()}, nil
		}

		k.Logger(ctx).Debug(sdkerrors.Wrapf(err, "expected transaction (%s) could not be verified", txID).Error())
		return &sdk.Result{Log: err.Error(), Events: ctx.EventManager().Events()}, nil
	}
}

// This can be used as a potential hook to immediately act on a poll being decided by the vote
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, b types.Balancer, msg *types.MsgVoteVerifiedTx) (*sdk.Result, error) {
	if err := v.TallyVote(ctx, msg); err != nil {
		return nil, err
	}

	if confirmed := v.Result(ctx, msg.Poll()); confirmed != nil {
		outPoint, err := types.OutPointFromStr(msg.PollMeta.ID)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		if err := k.ProcessVerificationResult(ctx, msg.PollMeta.ID, confirmed.(bool)); err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, fmt.Sprintf("utxo for poll %s was not stored", msg.PollMeta.String()))
		}
		v.DeletePoll(ctx, msg.Poll())

		info, ok := k.GetVerifiedOutPointInfo(ctx, outPoint)
		if !ok {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, "outpoint not verified")
		}
		id, ok := k.GetKeyIDByAddress(ctx, info.DepositAddr)
		if !ok {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, "key id not found")
		}
		k.SetKeyIDByOutpoint(ctx, outPoint, id)

		err = enqueueForTransfer(ctx, k, b, info)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, fmt.Sprintf("error while preparing transfer: %v", err))
		}
	}
	return &sdk.Result{}, nil
}

func handleMsgSignTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snap types.Snapshotter, msg types.MsgSignTx) (*sdk.Result, error) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)

	k.SetRawTx(ctx, msg.Outpoint, msg.RawTx)

	// Print out the hash that becomes the input for the threshold signing
	hash, err := k.GetHashToSign(ctx, msg.RawTx)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}
	serializedHash := hex.EncodeToString(hash)
	k.Logger(ctx).Info(fmt.Sprintf("bitcoin tx to sign: %s", serializedHash))

	keyID, ok := k.GetKeyIDByOutpoint(ctx, msg.Outpoint)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrBitcoin, "no key ID for chain %s found", balance.Bitcoin)
	}

	round, ok := signer.GetSnapshotRoundForKeyID(ctx, keyID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrBitcoin, "no snapshot round for key ID %s registered", keyID)
	}
	s, ok := snap.GetSnapshot(ctx, round)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "no snapshot found")
	}
	err = signer.StartSign(ctx, keyID, serializedHash, hash, s.Validators)
	if err != nil {
		if !ok {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
	}

	return &sdk.Result{
		Data:   hash,
		Log:    fmt.Sprintf("successfully started signing protocol for transaction that spends %s.", msg.Outpoint),
		Events: ctx.EventManager().Events(),
	}, nil
}

func trackAddress(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, address string, rescan bool) {
	// Importing an address takes a long time, therefore it cannot be done in the critical path
	go func(logger log.Logger) {
		if rescan {
			logger.Debug("Rescanning entire Bitcoin blockchain for past transactions. This will take a while.")
		}
		if err := rpc.ImportAddressRescan(address, "", rescan); err != nil {
			logger.Error(fmt.Sprintf("Could not track address %v", address))
		} else {
			logger.Debug(fmt.Sprintf("successfully tracked address %v", address))
		}
		// ctx might not be valid anymore when err is returned, so closing over logger to be safe
	}(k.Logger(ctx))

	k.SetTrackedAddress(ctx, address)
}

func verifyTx(rpc types.RPCClient, expectedInfo types.OutPointInfo, requiredConfirmations uint64) error {
	actualInfo, err := rpc.GetOutPointInfo(expectedInfo.OutPoint)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Bitcoin transaction")
	}

	if actualInfo.DepositAddr != expectedInfo.DepositAddr {
		return fmt.Errorf("expected destination address does not match actual destination address")
	}

	if actualInfo.Amount != expectedInfo.Amount {
		return fmt.Errorf("expected amount does not match actual amount")
	}

	if actualInfo.Confirmations < requiredConfirmations {
		return fmt.Errorf("not enough confirmations yet")
	}

	return nil
}

func enqueueForTransfer(ctx sdk.Context, k keeper.Keeper, b types.Balancer, info types.OutPointInfo) error {
	depositAddr := balance.CrossChainAddress{Address: info.DepositAddr, Chain: balance.Bitcoin}
	recipient, ok := b.GetRecipient(ctx, depositAddr)
	// Do nothing if not linked
	if !ok {
		return nil
	}

	amount := sdk.NewInt64Coin(denom.Satoshi, int64(info.Amount))

	err := b.EnqueueForTransfer(ctx, depositAddr, amount)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info(fmt.Sprintf("Transfer of %s to cross chain address %s in %s successfully prepared",
		info.Amount.String(), recipient.Address, recipient.Chain.String()))

	return nil
}
