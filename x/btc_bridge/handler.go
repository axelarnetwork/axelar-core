package btc_bridge

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

const bitcoin = "bitcoin"

func NewHandler(k keeper.Keeper, v types.Voter, rpc types.RPCClient, s types.Signer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgTrackAddress:
			return handleMsgTrackAddress(ctx, k, rpc, msg)
		case types.MsgTrackAddressFromPubKey:
			return handleMsgTrackAddressFromPubKey(ctx, k, rpc, s, msg)
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, v, rpc, msg)
		case types.MsgRawTx:
			return handleMsgRawTx(ctx, k, v, msg)
		case types.MsgWithdraw:
			return handleMsgWithdraw(ctx, k, rpc, s, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgTrackAddressFromPubKey(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, s types.Signer, msg types.MsgTrackAddressFromPubKey) (*sdk.Result, error) {
	key, err := s.GetKey(ctx, msg.KeyID)
	if err != nil {
		return nil, fmt.Errorf("keyId not recognized")
	}

	addr, err := addressFromKey(key, msg.Chain)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not convert the given public key into a bitcoin address")
	}

	trackAddress(ctx, k, rpc, addr.EncodeAddress())

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyId, msg.KeyID),
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

func handleMsgTrackAddress(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, msg types.MsgTrackAddress) (*sdk.Result, error) {
	k.Logger(ctx).Debug(fmt.Sprintf("start tracking address %v", msg.Address))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Address),
		),
	)

	trackAddress(ctx, k, rpc, msg.Address)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func trackAddress(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, address string) {
	// Importing an address takes a long time, therefore it cannot be done in the critical path.
	// ctx might not be valid anymore when err is returned, so closing over logger to be safe
	go func(logger log.Logger) {
		if err := rpc.ImportAddress(address); err != nil {
			logger.Error(fmt.Sprintf("Could not track address %v", address))
		} else {
			logger.Debug(fmt.Sprintf("successfully tracked all past transaction for address %v", address))
		}
	}(k.Logger(ctx))

	k.SetTrackedAddress(ctx, address)
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, rpc types.RPCClient, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying bitcoin transaction")
	txId := msg.UTXO.Hash.String()
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxId, txId),
			sdk.NewAttribute(types.AttributeAmount, msg.UTXO.Amount.String()),
		),
	)

	k.SetUTXO(ctx, txId, msg.UTXO)

	txForVote := exported.ExternalTx{Chain: bitcoin, TxID: txId}
	if err := verifyTx(rpc, msg.UTXO, k.GetConfirmationHeight(ctx)); err != nil {
		v.SetFutureVote(ctx, exported.FutureVote{Tx: txForVote, LocalAccept: false})

		k.Logger(ctx).Debug(sdkerrors.Wrapf(err,
			"expected transaction (%s) could not be verified", txId).Error())
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

func handleMsgRawTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg types.MsgRawTx) (*sdk.Result, error) {
	txId := msg.TxHash.String()
	if !v.IsVerified(ctx, exported.ExternalTx{
		Chain: bitcoin,
		TxID:  txId,
	}) {
		return nil, fmt.Errorf("transaction not verified")
	}
	utxo := k.GetUTXO(ctx, txId)
	if utxo == nil {
		return nil, fmt.Errorf("transaction ID is not known")
	}

	tx := wire.NewMsgTx(wire.TxVersion)

	outPoint := wire.NewOutPoint(utxo.Hash, utxo.VoutIdx)
	// The signature script will be set later
	txIn := wire.NewTxIn(outPoint, nil, nil)
	tx.AddTxIn(txIn)

	addrScript, err := txscript.PayToAddrScript(msg.Destination)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create pay-to-address script for destination address")
	}
	txOut := wire.NewTxOut(int64(msg.Amount), addrScript)
	tx.AddTxOut(txOut)

	k.SetRawTx(ctx, txId, tx)

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
		Log:    "successfully created withdraw transaction for Bitcoin",
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgWithdraw(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, signer types.Signer, msg types.MsgWithdraw) (*sdk.Result, error) {
	utxo := k.GetUTXO(ctx, msg.TxID)
	if utxo == nil {
		return nil, fmt.Errorf("transaction ID is not known")
	}

	rawTx := k.GetRawTx(ctx, msg.TxID)
	if rawTx == nil {
		return nil, fmt.Errorf("withdraw tx for ID %s has not been prepared yet", msg.TxID)
	}
	r, s := signer.GetSig(ctx, msg.SignatureID)
	if r == nil || s == nil {
		return nil, fmt.Errorf("signature not found")
	}
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

	vm, err := txscript.NewEngine(utxo.PkScript(), rawTx, 0, flags, nil, nil, rawTx.TxOut[0].Value)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create execution engine, aborting")
	}
	if err := vm.Execute(); err != nil {
		return nil, sdkerrors.Wrap(err, "transaction failed to execute, aborting")
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

	// This is beyond Axelar's control, so can only log the error but move on regardless
	if _, err := rpc.SendRawTransaction(rawTx, false); err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "sending transaction to Bitcoin failed").Error())
		return &sdk.Result{
			Log:    "failed to send transaction to Bitcoin (other nodes might have succeeded)",
			Events: ctx.EventManager().Events(),
		}, nil
	}

	return &sdk.Result{
		Log:    "successfully sent withdraw transaction to Bitcoin",
		Events: ctx.EventManager().Events(),
	}, nil
}

func verifyTx(rpc types.RPCClient, utxo types.UTXO, expectedConfirmationHeight uint64) error {
	actualTx, err := rpc.GetRawTransactionVerbose(utxo.Hash)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Bitcoin transaction")
	}

	if utxo.VoutIdx >= uint32(len(actualTx.Vout)) {
		return fmt.Errorf("vout index out of range")
	}

	vout := actualTx.Vout[utxo.VoutIdx]

	if len(vout.ScriptPubKey.Addresses) > 1 {
		return fmt.Errorf("deposit must be only spendable by a single address")
	}
	if vout.ScriptPubKey.Addresses[0] != utxo.Address.String() {
		return fmt.Errorf("expected destination address does not match actual destination address")
	}

	actualAmount, err := btcutil.NewAmount(vout.Value)
	if err != nil {
		return sdkerrors.Wrap(err, "could not parse transaction amount of the Bitcoin response")
	}
	if utxo.Amount != actualAmount {
		return fmt.Errorf("expected amount does not match actual amount")
	}

	if actualTx.Confirmations < expectedConfirmationHeight {
		return fmt.Errorf("not enough confirmations yet")
	}

	return nil
}
