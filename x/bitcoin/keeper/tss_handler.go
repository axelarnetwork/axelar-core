package keeper

import (
	"encoding/hex"
	"fmt"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

type signingAbortError struct {
	err          error
	abortedKeyID tss.KeyID
}

func (e *signingAbortError) Error() string {
	return e.err.Error()
}

// NewTssHandler returns the handler for processing signatures delivered by the tss module
func NewTssHandler(keeper types.BTCKeeper, signer types.Signer) tss.Handler {
	return func(ctx sdk.Context, info tss.SignInfo) error {
		for _, txType := range types.GetTxTypes() {
			handleUnsignedTxForTxType(ctx, keeper, signer, txType)
		}
		return nil
	}
}

func handleUnsignedTxForTxType(ctx sdk.Context, keeper types.BTCKeeper, signer types.Signer, txType types.TxType) {
	unsignedTx, ok := keeper.GetUnsignedTx(ctx, txType)
	if !ok || !unsignedTx.Is(types.Signing) {
		keeper.Logger(ctx).Debug(fmt.Sprintf("no unsigned %s transaction ready", txType.SimpleString()))
		return
	}

	signedTx, err := assembleTx(ctx, keeper, signer, &unsignedTx)
	if err != nil {
		switch e := err.(type) {
		case *signingAbortError:
			ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueSigningAborted),
				sdk.NewAttribute(types.AttributeTxType, txType.SimpleString()),
			))

			unsignedTx.ConfirmationRequired = true
			unsignedTx.Status = types.Aborted
			unsignedTx.PrevAbortedKeyId = e.abortedKeyID
			keeper.SetUnsignedTx(ctx, unsignedTx)
		default:
		}

		keeper.Logger(ctx).Debug(sdkerrors.Wrapf(err, "failed to assemble tx %s with signatures", unsignedTx.GetTx().TxHash().String()).Error())
		return
	}

	txHash := signedTx.TxHash()
	knownOutPoints, err := getKnownOutPoints(ctx, keeper, signedTx)
	if err != nil {
		keeper.Logger(ctx).Debug(sdkerrors.Wrapf(err, "failed to get known out points in tx %s", txHash.String()).Error())
		return
	}

	for _, outPoint := range knownOutPoints {
		// Ignore error here because out point here must be known
		addressInfo, _ := keeper.GetAddress(ctx, outPoint.Address)

		if unsignedTx.ConfirmationRequired {
			unconfirmedAmount := keeper.GetUnconfirmedAmount(ctx, addressInfo.KeyID)
			keeper.SetUnconfirmedAmount(ctx, addressInfo.KeyID, unconfirmedAmount+outPoint.Amount)
		} else {
			keeper.SetConfirmedOutpointInfo(ctx, addressInfo.KeyID, outPoint)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(types.EventTypeOutpointConfirmation,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
					sdk.NewAttribute(types.AttributeKeyOutPointInfo, string(types.ModuleCdc.MustMarshalJSON(&outPoint))),
				),
			)
		}
	}

	// Rotate key if necessary
	if unsignedTx.Info.RotateKey {
		var keyRole tss.KeyRole

		switch txType {
		case types.MasterConsolidation:
			keyRole = tss.MasterKey
		case types.SecondaryConsolidation:
			keyRole = tss.SecondaryKey
		default:
			keeper.Logger(ctx).Error(fmt.Sprintf("%s transaction should not involve key rotation", txType.SimpleString()))
			return
		}

		if err := signer.RotateKey(ctx, exported.Bitcoin, keyRole); err != nil {
			keeper.Logger(ctx).Error(sdkerrors.Wrap(err, fmt.Sprintf("failed to rotate to the next %s key", keyRole.SimpleString())).Error())
			return
		}
	}

	keeper.DeleteUnsignedTx(ctx, txType)
	keeper.SetSignedTx(ctx, types.NewSignedTx(txType, signedTx, unsignedTx.ConfirmationRequired, unsignedTx.AnyoneCanSpendVout))
	keeper.SetLatestSignedTxHash(ctx, txType, txHash)

	// Notify that consolidation tx can be queried
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueSigned),
		sdk.NewAttribute(types.AttributeTxType, txType.SimpleString()),
	))
	keeper.Logger(ctx).Info(fmt.Sprintf("transaction %s is fully signed", txHash.String()))
}

func assembleTx(ctx sdk.Context, k types.BTCKeeper, signer types.Signer, unsignedTx *types.UnsignedTx) (*wire.MsgTx, error) {
	tx := unsignedTx.GetTx()
	outPointsToSign, err := getOutPointsToSign(ctx, tx, k)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to collect outpoints waiting to be signed for unsigned tx %s", tx.TxHash().String())
	}

	// Assemble transaction with signatures
	var sigs [][]btcec.Signature
	for _, inputInfo := range unsignedTx.Info.InputInfos {
		var sigsForOutPoint []btcec.Signature

		for _, sigRequirement := range inputInfo.SigRequirements {
			sigHashHex := hex.EncodeToString(sigRequirement.SigHash)
			sigID := fmt.Sprintf("%s-%s", sigHashHex, sigRequirement.KeyID)
			sig, status := signer.GetSig(ctx, sigID)
			if status != tss.SigStatus_Signed {
				err := fmt.Errorf("signature for tx %s not yet found", sigID)

				if status != tss.SigStatus_Queued && status != tss.SigStatus_Signing {
					return nil, &signingAbortError{err: err, abortedKeyID: sigRequirement.KeyID}
				}

				return nil, err
			}

			sigsForOutPoint = append(sigsForOutPoint, btcec.Signature{R: sig.R, S: sig.S})
		}

		sigs = append(sigs, sigsForOutPoint)
	}

	signedTx, err := types.AssembleBtcTx(tx, outPointsToSign, sigs)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func getKnownOutPoints(ctx sdk.Context, k types.BTCKeeper, signedTx *wire.MsgTx) ([]types.OutPointInfo, error) {
	var knownOutPoints []types.OutPointInfo

	networkName := k.GetNetwork(ctx).Name
	network, err := types.NetworkFromStr(networkName)
	if err != nil {
		return nil, sdkerrors.Wrap(err, fmt.Sprintf("failed to get network %s", networkName))
	}

	txHash := signedTx.TxHash()
	// Find all outpoints that axelar controls the keys of
	for i, output := range signedTx.TxOut {
		_, addresses, _, err := txscript.ExtractPkScriptAddrs(output.PkScript, network.Params())
		if err != nil {
			continue
		}

		if len(addresses) != 1 {
			continue
		}

		addressInfo, ok := k.GetAddress(ctx, addresses[0].EncodeAddress())
		if !ok {
			continue
		}

		outpointInfo := types.NewOutPointInfo(wire.NewOutPoint(&txHash, uint32(i)), btcutil.Amount(output.Value), addressInfo.Address)
		knownOutPoints = append(knownOutPoints, outpointInfo)
	}

	return knownOutPoints, nil
}

func getOutPointsToSign(ctx sdk.Context, tx *wire.MsgTx, k types.BTCKeeper) ([]types.OutPointToSign, error) {
	var toSign []types.OutPointToSign
	for _, in := range tx.TxIn {
		prevOutInfo, state, ok := k.GetOutPointInfo(ctx, in.PreviousOutPoint)
		if !ok {
			return nil, fmt.Errorf("cannot find %s", in.PreviousOutPoint.String())
		}

		if state != types.OutPointState_Spent {
			return nil, fmt.Errorf("outpoint %s is not set as spent", in.PreviousOutPoint.String())
		}

		addr, ok := k.GetAddress(ctx, prevOutInfo.Address)
		if !ok {
			return nil, fmt.Errorf("address %s not found", prevOutInfo.Address)
		}

		toSign = append(toSign, types.OutPointToSign{
			OutPointInfo: prevOutInfo,
			AddressInfo:  addr,
		})
	}
	return toSign, nil
}
