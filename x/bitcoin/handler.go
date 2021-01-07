package bitcoin

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewHandler creates an sdk.Handler for all bitcoin type messages
func NewHandler(k keeper.Keeper, v types.Voter, rpc types.RPCClient, s types.Signer, b types.Balancer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgTrack:
			return handleMsgTrack(ctx, k, s, rpc, msg)
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, v, rpc, msg)
		case *types.MsgVoteVerifiedTx:
			return handleMsgVoteVerifiedTx(ctx, k, v, msg)
		case types.MsgRawTx:
			return handleMsgRawTx(ctx, k, s, msg)
		case types.MsgSendTx:
			return handleMsgSendTx(ctx, k, rpc, s, msg)
		case types.MsgTransfer:
			return handleMsgTransfer(ctx, b, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgTransfer(ctx sdk.Context, b types.Balancer, msg types.MsgTransfer) (*sdk.Result, error) {
	btcAddr := balance.CrossChainAddress{Chain: balance.Bitcoin, Address: msg.BTCAddress.String()}
	b.LinkAddresses(ctx, btcAddr, msg.Destination)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, btcAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Destination.String()),
		),
	)

	return &sdk.Result{
		Log:    fmt.Sprintf("successfully linked {%s} and {%s}", btcAddr.String(), msg.Destination.String()),
		Events: ctx.EventManager().Events(),
	}, nil
}

// This can be used as a potential hook to immediately act on a poll being decided by the vote
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg *types.MsgVoteVerifiedTx) (*sdk.Result, error) {
	if err := v.TallyVote(ctx, msg); err != nil {
		return nil, err
	}

	if confirmed := v.Result(ctx, msg.Poll()); confirmed != nil {
		if err := k.ProcessUTXOPollResult(ctx, msg.PollMeta.ID, confirmed.(bool)); err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, fmt.Sprintf("utxo for poll %s was not stored", msg.PollMeta.String()))
		}
		v.DeletePoll(ctx, msg.Poll())
	}
	return &sdk.Result{}, nil
}

func handleMsgTrack(ctx sdk.Context, k keeper.Keeper, s types.Signer, rpc types.RPCClient, msg types.MsgTrack) (*sdk.Result, error) {
	var encodedAddr string
	if msg.Mode == types.ModeSpecificAddress {
		encodedAddr = msg.Address.EncodedString
	} else {
		var key ecdsa.PublicKey
		var ok bool

		switch msg.Mode {
		case types.ModeSpecificKey:
			key, ok = s.GetKey(ctx, msg.KeyID)
			if !ok {
				return nil, sdkerrors.Wrap(types.ErrBitcoin, fmt.Sprintf("key with ID %s not found", msg.KeyID))
			}
		case types.ModeCurrentMasterKey:
			key, ok = s.GetCurrentMasterKey(ctx, balance.Bitcoin)
			if !ok {
				return nil, sdkerrors.Wrap(types.ErrBitcoin, "master key not set")
			}
		}

		addr, err := types.PKHashFromKey(key, msg.Chain)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, sdkerrors.Wrap(err, "could not convert the given public key into a bitcoin address").Error())
		}

		encodedAddr = addr.EncodeAddress()
	}
	k.Logger(ctx).Debug(fmt.Sprintf("start tracking address %v", encodedAddr))
	trackAddress(ctx, k, rpc, encodedAddr, msg.Rescan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyId, msg.KeyID),
			sdk.NewAttribute(types.AttributeAddress, encodedAddr),
		),
	)

	return &sdk.Result{
		Data:   []byte(encodedAddr),
		Log:    fmt.Sprintf("successfully tracked address %s", encodedAddr),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, rpc types.RPCClient, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying bitcoin transaction")
	minConfHeight := k.GetConfirmationHeight(ctx)
	if msg.OutPointInfo.Confirmations < minConfHeight {
		return nil, sdkerrors.Wrapf(types.ErrBitcoin, "not enough confirmations")
	}

	txId := msg.OutPointInfo.OutPoint.Hash.String()
	poll := exported.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: txId}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxId, txId),
			sdk.NewAttribute(types.AttributeAmount, msg.OutPointInfo.Amount.String()),
		),
	)

	k.SetOutpointInfo(ctx, txId, msg.OutPointInfo)

	/*
	 Anyone not able to verify the transaction will automatically record a negative vote,
	 but only validators will later send out that vote.
	*/

	err := verifyTx(rpc, msg.OutPointInfo)
	switch err {
	// verification successful
	case nil:
		if err := v.RecordVote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "voting failed").Error())
			return &sdk.Result{Log: err.Error(), Events: ctx.EventManager().Events()}, nil
		}

		k.Logger(ctx).Debug(fmt.Sprintf("transaction (%s) was verified", txId))
		return &sdk.Result{Log: "successfully verified transaction", Events: ctx.EventManager().Events()}, nil
	// verification unsuccessful
	default:
		if err := v.RecordVote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "voting failed").Error())
			return &sdk.Result{Log: err.Error(), Events: ctx.EventManager().Events()}, nil
		}

		k.Logger(ctx).Debug(sdkerrors.Wrapf(err, "expected transaction (%s) could not be verified", txId).Error())
		return &sdk.Result{Log: err.Error(), Events: ctx.EventManager().Events()}, nil
	}
}

func handleMsgRawTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg types.MsgRawTx) (*sdk.Result, error) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)

	for _, in := range msg.RawTx.TxIn {
		txID := in.PreviousOutPoint.Hash.String()
		if !isTxVerified(ctx, v, txID) {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, fmt.Sprintf("transaction %s not verified", txID))
		}
	}
	k.SetRawTx(ctx, msg.TxID, msg.RawTx)

	script, err := getPkScript(ctx, k, msg.TxID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	// Print out the hash that becomes the input for the threshold signing
	hash, err := txscript.CalcSignatureHash(script, txscript.SigHashAll, msg.RawTx, 0)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	} else {
		k.Logger(ctx).Info(fmt.Sprintf("bitcoin tx to sign: %s", k.Codec().MustMarshalJSON(hash)))
		return &sdk.Result{
			Data:   hash,
			Log:    fmt.Sprintf("successfully created an unsigned transaction for Bitcoin. Hash to sign: %s", k.Codec().MustMarshalJSON(hash)),
			Events: ctx.EventManager().Events(),
		}, nil
	}
}

func handleMsgSendTx(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, s types.Signer, msg types.MsgSendTx) (*sdk.Result, error) {
	pk, ok := s.GetKeyForSigID(ctx, msg.SignatureID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, fmt.Sprintf("could not find a corresponding key for sig ID %s", msg.SignatureID))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxId, msg.TxID),
			sdk.NewAttribute(types.AttributeSigId, msg.SignatureID),
		),
	)

	rawTx, err := assembleBtcTx(ctx, k, s, msg.TxID, pk, msg.SignatureID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	// This is beyond axelar's control, so we can only log the error and move on regardless
	hash, err := rpc.SendRawTransaction(rawTx, false)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "sending transaction to Bitcoin failed").Error())
		return &sdk.Result{Log: fmt.Sprintf("failed to sent transaction to Bitcoin (other nodes might have succeeded): %s", err.Error())}, nil
	}

	return &sdk.Result{Data: hash[:], Log: fmt.Sprintf("successfully sent transaction %s to Bitcoin", hash), Events: ctx.EventManager().Events()}, nil
}

func isTxVerified(ctx sdk.Context, v types.Voter, txId string) bool {
	poll := exported.PollMeta{ID: txId, Module: types.ModuleName, Type: types.MsgVerifyTx{}.Type()}
	res := v.Result(ctx, poll)
	return res != nil && res.(bool)
}

func assembleBtcTx(ctx sdk.Context, k keeper.Keeper, s types.Signer, txID string, pk ecdsa.PublicKey, sigID string) (*wire.MsgTx, error) {
	rawTx := k.GetRawTx(ctx, txID)
	if rawTx == nil {
		return nil, fmt.Errorf("withdraw tx for ID %s has not been prepared yet", txID)
	}

	sigScript, err := createSigScript(ctx, s, sigID, pk)
	if err != nil {
		return nil, err
	}
	rawTx.TxIn[0].SignatureScript = sigScript

	pkScript, err := getPkScript(ctx, k, txID)
	if err != nil {
		return nil, err
	}
	if err := validateTxScripts(rawTx, pkScript); err != nil {
		return nil, err
	}
	return rawTx, nil
}

func createSigScript(ctx sdk.Context, s types.Signer, sigID string, pk ecdsa.PublicKey) ([]byte, error) {
	sig, ok := s.GetSig(ctx, sigID)
	if !ok {
		return nil, fmt.Errorf("signature not found")
	}

	btcSig := btcec.Signature{
		R: sig.R,
		S: sig.S,
	}
	sigBytes := append(btcSig.Serialize(), byte(txscript.SigHashAll))

	key := btcec.PublicKey(pk)
	keyBytes := key.SerializeCompressed()

	sigScript, err := txscript.NewScriptBuilder().AddData(sigBytes).AddData(keyBytes).Script()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create bitcoin signature script")
	}
	return sigScript, nil
}

func validateTxScripts(tx *wire.MsgTx, pkScript []byte) error {
	flags := txscript.StandardVerifyFlags

	// execute (dry-run) the public key and signature script to validate them
	scriptEngine, err := txscript.NewEngine(pkScript, tx, 0, flags, nil, nil, tx.TxOut[0].Value)
	if err != nil {
		return sdkerrors.Wrap(err, "could not create execution engine, aborting")
	}
	if err := scriptEngine.Execute(); err != nil {
		return sdkerrors.Wrap(err, "transaction failed to execute, aborting")
	}
	return nil
}

func getPkScript(ctx sdk.Context, k keeper.Keeper, txId string) ([]byte, error) {
	out, ok := k.GetOutPoint(ctx, txId)
	if !ok {
		return nil, fmt.Errorf("transaction ID is not known")
	}
	return out.Recipient.PkScript(), nil
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

func verifyTx(rpc types.RPCClient, expectedInfo types.OutPointInfo) error {
	actualInfo, err := rpc.GetOutPointInfo(expectedInfo.OutPoint)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Bitcoin transaction")
	}

	if actualInfo.Recipient != expectedInfo.Recipient {
		return fmt.Errorf("expected destination address does not match actual destination address")
	}

	if actualInfo.Amount != expectedInfo.Amount {
		return fmt.Errorf("expected amount does not match actual amount")
	}

	if actualInfo.Confirmations < expectedInfo.Confirmations {
		return fmt.Errorf("not enough confirmations yet")
	}

	return nil
}
