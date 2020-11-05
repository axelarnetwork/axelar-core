package btc_bridge

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

func NewHandler(k keeper.Keeper, v types.Voter, b types.Bridge, s types.Signer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgTrackAddress:
			return handleMsgTrackAddress(ctx, k, b, msg)
		case types.MsgTrackAddressFromPubKey:
			return handleMsgTrackAddressFromPubKey(ctx, k, b, s, msg)
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, v, b, msg)
		case types.MsgWithdraw:
			return handleMsgWithdraw(ctx, k, b, s, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgTrackAddressFromPubKey(ctx sdk.Context, k keeper.Keeper, b types.Bridge, s types.Signer, msg types.MsgTrackAddressFromPubKey) (*sdk.Result, error) {
	key := s.GetKey(ctx, msg.Chain)
	if key == (ecdsa.PublicKey{}) {
		return nil, fmt.Errorf("keyId not recognized")
	}

	addr, err := addressFromKey(key, msg.Chain)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not convert the given public key into a bitcoin address")
	}

	trackAddress(ctx, k, b, addr.EncodeAddress())

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, addr.EncodeAddress()),
		),
	)

	return &sdk.Result{Log: "successfully created a tracked address", Events: ctx.EventManager().Events()}, nil
}

func addressFromKey(key ecdsa.PublicKey, chain string) (*btcutil.AddressPubKey, error) {
	btcPK := btcec.PublicKey(key)
	var params *chaincfg.Params
	switch chain {
	case chaincfg.MainNetParams.Name:
		params = &chaincfg.MainNetParams
	case chaincfg.TestNet3Params.Name:
		params = &chaincfg.TestNet3Params
	}

	// For compatibility we use the uncompressed key as the basis for address generation.
	// Could be changed in the future to decrease tx size
	return btcutil.NewAddressPubKey(btcPK.SerializeUncompressed(), params)
}

func handleMsgTrackAddress(ctx sdk.Context, k keeper.Keeper, b types.Bridge, msg types.MsgTrackAddress) (*sdk.Result, error) {
	k.Logger(ctx).Debug(fmt.Sprintf("start tracking address %v", msg.Address))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Address),
		),
	)

	trackAddress(ctx, k, b, msg.Address)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func trackAddress(ctx sdk.Context, k keeper.Keeper, b types.Bridge, address string) {
	// Importing an address takes a long time, therefore it cannot be done in the critical path.
	// ctx might not be valid anymore when err is returned, so closing over logger to be safe
	go func(logger log.Logger) {
		if err := b.TrackAddress(address); err != nil {
			logger.Error(fmt.Sprintf("Could not track address %v", address))
		} else {
			logger.Debug(fmt.Sprintf("successfully tracked all past transaction for address %v", address))
		}
	}(k.Logger(ctx))

	k.SetTrackedAddress(ctx, address)
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, b types.Bridge, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying bitcoin transaction")

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxId, msg.TxHash.String()),
			sdk.NewAttribute(types.AttributeAmount, msg.Amount.String()),
		),
	)

	txForVote := exported.ExternalTx{Chain: "bitcoin", TxID: msg.TxHash.String()}
	if err := b.VerifyTx(msg.TxHash, msg.Amount); err != nil {
		v.SetFutureVote(ctx, exported.FutureVote{Tx: txForVote, LocalAccept: false})
		k.Logger(ctx).Debug(sdkerrors.Wrapf(err,
			"expected transaction (%s) could not be verified", msg.TxHash.String()).Error())
		return &sdk.Result{
			Log:    err.Error(),
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}, nil
	} else {
		v.SetFutureVote(ctx, exported.FutureVote{Tx: txForVote, LocalAccept: true})
		return &sdk.Result{
			Log:    "successfully verified transaction",
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(true),
			Events: ctx.EventManager().Events(),
		}, nil
	}
}

func handleMsgWithdraw(ctx sdk.Context, k keeper.Keeper, b types.Bridge, signer types.Signer, msg types.MsgWithdraw) (*sdk.Result, error) {
	rawTx, pkscript, amount := k.GetRaw(msg.TxID)
	if rawTx == nil {
		return nil, fmt.Errorf("withdraw tx for ID %s has not been prepared yet", msg.TxID)
	}
	r, s := signer.GetSig(ctx, msg.SignatureID)
	sig := btcec.Signature{
		R: r,
		S: s,
	}

	sigScript, err := txscript.NewScriptBuilder().AddData(append(sig.Serialize(), byte(txscript.SigHashAll))).Script()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create bitcoin singature script")
	}
	rawTx.TxIn[0].SignatureScript = sigScript

	flags := txscript.StandardVerifyFlags

	vm, err := txscript.NewEngine(pkscript, rawTx, 0, flags, nil, nil, amount)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create execution engine, aborting")
	}
	if err := vm.Execute(); err != nil {
		return nil, sdkerrors.Wrap(err, "transaction failed to execute, aborting")
	}

	if err = b.Send(rawTx); err != nil {
		return nil, err
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

	return &sdk.Result{
		Log:    "successfully sent withdraw transaction to Bitcoin",
		Events: ctx.EventManager().Events(),
	}, nil
}
