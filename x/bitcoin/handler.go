package bitcoin

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils/denom"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func NewHandler(k keeper.Keeper, v types.Voter, rpc types.RPCClient, s types.Signer, b types.Balancer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgTrack:
			return handleMsgTrack(ctx, k, s, rpc, msg)
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, v, rpc, s, msg)
		case *types.MsgVoteVerifiedTx:
			return handleMsgVoteVerifiedTx(ctx, k, v, b, msg)
		case types.MsgRawTx:
			return handleMsgRawTx(ctx, k, b, s, msg)
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
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, b types.Balancer, msg *types.MsgVoteVerifiedTx) (*sdk.Result, error) {
	if err := v.TallyVote(ctx, msg); err != nil {
		return nil, err
	}

	if confirmed := v.Result(ctx, msg.Poll()); confirmed != nil {
		if err := k.ProcessUTXOPollResult(ctx, msg.PollMeta.ID, confirmed.(bool)); err != nil {
			return nil, err
		}
		v.DeletePoll(ctx, msg.Poll())

		if confirmed.(bool) {
			utxo, ok := k.GetUTXO(ctx, msg.PollMeta.ID)
			if !ok {
				return nil, sdkerrors.Wrap(types.ErrBitcoin, fmt.Sprintf("utxo for poll %s was not stored", msg.PollMeta.String()))
			}
			sender := balance.CrossChainAddress{Chain: balance.Bitcoin, Address: utxo.Sender.String()}
			if recp, linked := b.GetRecipient(ctx, sender); linked {
				satoshi := sdk.NewCoin(denom.Satoshi, sdk.NewInt(int64(utxo.Amount)))
				err := b.PrepareForTransfer(ctx, sender, satoshi)
				if err != nil {
					return nil, sdkerrors.Wrap(types.ErrBitcoin, "could not prepare transaction for transfer")
				}
				k.Logger(ctx).Debug(fmt.Sprintf("%s prepared for tranfer to %s", satoshi.String(), recp.String()))
			}
		}
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

		addr, err := pkHashFromKey(key, msg.Chain)
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

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, rpc types.RPCClient, s types.Signer, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying bitcoin transaction")
	txId := msg.UTXO.Hash.String()
	if _, ok := k.GetUTXO(ctx, txId); ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "transaction already verified")
	}

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
			sdk.NewAttribute(types.AttributeAmount, msg.UTXO.Amount.String()),
		),
	)

	if msg.Mode != types.ModeSpecificAddress {
		var key ecdsa.PublicKey
		var ok bool
		switch msg.Mode {
		case types.ModeCurrentMasterKey:
			key, ok = s.GetCurrentMasterKey(ctx, balance.Bitcoin)
			if !ok {
				return nil, sdkerrors.Wrap(types.ErrBitcoin, "no current master key assigned")
			}
		case types.ModeNextMasterKey:
			key, ok = s.GetNextMasterKey(ctx, balance.Bitcoin)
			if !ok {
				return nil, sdkerrors.Wrap(types.ErrBitcoin, "no next master key assigned")
			}
		}
		addr, err := pkHashFromKey(key, msg.Network)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, "could not derive Bitcoin address next master key")
		}
		msg.UTXO.Recipient = types.BtcAddress{
			Network:       msg.Network,
			EncodedString: addr.EncodeAddress(),
		}
	}

	k.SetUTXOForPoll(ctx, poll.ID, msg.UTXO)

	/*
	 Anyone not able to verify the transaction will automatically record a negative vote,
	 but only validators will later send out that vote.
	*/
	err := verifyTx(rpc, msg.UTXO, k.GetConfirmationHeight(ctx))
	if err != nil {
		k.Logger(ctx).Debug(fmt.Sprintf("tx %s verification failed: %s", txId, err.Error()))
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

func handleMsgRawTx(ctx sdk.Context, k keeper.Keeper, b types.Balancer, s types.Signer, msg types.MsgRawTx) (*sdk.Result, error) {
	txId := msg.TxHash.String()
	var destination btcutil.Address
	var err error
	switch msg.Mode {
	case types.ModeSpecificAddress:
		// msg.ValidateBasic will be called before the handler, so there is no way we have a malformed address here
		destination, _ = msg.Destination.Convert()
	case types.ModeNextMasterKey:
		pk, ok := s.GetNextMasterKey(ctx, balance.Bitcoin)
		if !ok {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, "next master key not set")
		}
		destination, err = pkHashFromKey(pk, msg.Chain)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, "could not create bitcoin address from master key")
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxId, txId),
			sdk.NewAttribute(types.AttributeAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeDestination, destination.String()),
		),
	)

	utxo, ok := k.GetUTXO(ctx, txId)
	if !ok {
		return nil, fmt.Errorf("transaction not verified")
	}
	if hash, err := createRawTx(ctx, k, b, utxo, destination, int64(msg.Amount)); err != nil {
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
	utxo, ok := k.GetUTXO(ctx, txId)
	if !ok {
		return nil, fmt.Errorf("transaction ID is not known")
	}
	return utxo.Recipient.PkScript(), nil
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

/*
	Creating a Bitcoin transaction one step at a time:
		1. Create the transaction message
		2. Get the output of the deposit transaction and convert it into the transaction input
		3. Create a new output
	See https://blog.hlongvu.com/post/t0xx5dejn3-Understanding-btcd-Part-4-Create-and-Sign-a-Bitcoin-transaction-with-btcd
*/
func createRawTx(ctx sdk.Context, k keeper.Keeper, b types.Balancer, utxo types.UTXO, destination btcutil.Address, amount int64) (hash []byte, err error) {
	sender := balance.CrossChainAddress{Chain: balance.Bitcoin, Address: utxo.Sender.String()}
	if recipient, linked := b.GetRecipient(ctx, sender); linked {
		return nil, fmt.Errorf("transactions from %s are linked to recipient %s and cannot be withdrawn", sender, recipient)
	}

	tx := wire.NewMsgTx(wire.TxVersion)

	outPoint := wire.NewOutPoint(utxo.Hash, utxo.VoutIdx)
	// The signature script will be set later and we have no witness
	txIn := wire.NewTxIn(outPoint, nil, nil)
	tx.AddTxIn(txIn)

	addrScript, err := txscript.PayToAddrScript(destination)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create pay-to-address script for destination address")
	}

	txOut := wire.NewTxOut(amount, addrScript)
	tx.AddTxOut(txOut)

	k.SetRawTx(ctx, utxo.Hash.String(), tx)

	// Print out the hash that becomes the input for the threshold signing
	hash, err = txscript.CalcSignatureHash(utxo.Recipient.PkScript(), txscript.SigHashAll, tx, 0)
	k.Logger(ctx).Info(fmt.Sprintf("bitcoin tx to sign: %s", k.Codec().MustMarshalJSON(hash)))
	return hash, nil
}

// we use Pay2PKH for added security over Pay2PK as well as for the benefit of getting a parsed address back when calling
// getrawtransaction() on the Bitcoin rpc client
func pkHashFromKey(key ecdsa.PublicKey, chain types.Network) (*btcutil.AddressPubKeyHash, error) {
	btcPK := btcec.PublicKey(key)
	return btcutil.NewAddressPubKeyHash(btcutil.Hash160(btcPK.SerializeCompressed()), chain.Params())
}

func verifyTx(rpc types.RPCClient, utxo types.UTXO, expectedConfirmationHeight uint64) error {
	actualTx, err := rpc.GetRawTransactionVerbose(utxo.Hash)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Bitcoin transaction")
	}

	if uint32(len(actualTx.Vout)) <= utxo.VoutIdx {
		return fmt.Errorf("vout index out of range")
	}

	vout := actualTx.Vout[utxo.VoutIdx]

	if len(vout.ScriptPubKey.Addresses) != 1 {
		return fmt.Errorf("deposit must be only spendable by a single address")
	}

	if vout.ScriptPubKey.Addresses[0] != utxo.Recipient.String() {
		return fmt.Errorf("expected destination address does not match actual destination address")
	}

	actualAmount, err := btcutil.NewAmount(vout.Value)
	if err != nil {
		return sdkerrors.Wrap(err, "could not parse transaction amount of the Bitcoin response")
	}
	if actualAmount != utxo.Amount {
		return fmt.Errorf("expected amount does not match actual amount")
	}

	if actualTx.Confirmations < expectedConfirmationHeight {
		return fmt.Errorf("not enough confirmations yet")
	}

	if len(actualTx.Vin) < 1 {
		return fmt.Errorf("transaction does not have any input")
	}

	spentTxHash, err := chainhash.NewHashFromStr(actualTx.Vin[0].Txid)
	if err != nil {
		return fmt.Errorf("could not read the spent transaction's hash")
	}
	spentTx, err := rpc.GetRawTransactionVerbose(spentTxHash)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve spent Bitcoin transaction")
	}
	if uint32(len(spentTx.Vout)) <= actualTx.Vin[0].Vout {
		return fmt.Errorf("vout index of spent transaction out of range")
	}

	spentVout := spentTx.Vout[actualTx.Vin[0].Vout]

	if len(spentVout.ScriptPubKey.Addresses) != 1 {
		return fmt.Errorf("deposit must be sent by a single address")
	}

	if spentVout.ScriptPubKey.Addresses[0] != utxo.Sender.String() {
		return fmt.Errorf("expected recipient address does not match actual recipient address")
	}

	return nil
}
