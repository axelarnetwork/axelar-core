package bitcoin

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ types.BTCKeeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, k types.BTCKeeper, signer types.Signer) []abci.ValidatorUpdate {
	if req.Height%k.GetSigCheckInterval(ctx) != 0 {
		return nil
	}

	for _, keyRole := range tss.GetKeyRoles() {
		unsignedTx, ok := k.GetUnsignedTx(ctx, keyRole)
		if !ok || !unsignedTx.Is(types.Signing) {
			k.Logger(ctx).Debug(fmt.Sprintf("no unsigned %s key transaction ready", keyRole.SimpleString()))
			continue
		}

		signedTx, abort, err := assembleTx(ctx, k, signer, &unsignedTx)
		if err != nil {
			if abort {
				ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueSigningAborted),
					sdk.NewAttribute(types.AttributeKeyRole, keyRole.SimpleString()),
				))

				unsignedTx.ConfirmationRequired = true
				unsignedTx.Status = types.Aborted
				k.SetUnsignedTx(ctx, keyRole, unsignedTx)
			}

			k.Logger(ctx).Debug(sdkerrors.Wrapf(err, "failed to assemble tx %s with signatures", unsignedTx.GetTx().TxHash().String()).Error())
			continue
		}

		txHash := signedTx.TxHash()
		knownOutPoints, err := getKnownOutPoints(ctx, k, signedTx)
		if err != nil {
			k.Logger(ctx).Debug(sdkerrors.Wrapf(err, "failed to confirm known out points in tx %s", txHash.String()).Error())
			continue
		}

		for _, outPoint := range knownOutPoints {
			// Ignore error here because out point here must be known
			addressInfo, _ := k.GetAddress(ctx, outPoint.Address)

			if !unsignedTx.ConfirmationRequired {
				k.SetConfirmedOutpointInfo(ctx, addressInfo.KeyID, outPoint)

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(types.EventTypeOutpointConfirmation,
						sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
						sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
						sdk.NewAttribute(types.AttributeKeyOutPointInfo, string(types.ModuleCdc.MustMarshalJSON(&outPoint))),
					),
				)

				continue
			}

			unconfirmedAmount := k.GetUnconfirmedAmount(ctx, addressInfo.KeyID)
			k.SetUnconfirmedAmount(ctx, addressInfo.KeyID, unconfirmedAmount+outPoint.Amount)
		}

		// Assign the next key if necessary
		if unsignedTx.Info.AssignNextKey {
			nextKey, ok := signer.GetKey(ctx, unsignedTx.Info.NextKeyID)
			if !ok {
				k.Logger(ctx).Error(sdkerrors.Wrap(err, fmt.Sprintf("failed to get the next %s key %s to assign", keyRole, unsignedTx.Info.NextKeyID)).Error())
				continue
			}

			if err := signer.AssignNextKey(ctx, exported.Bitcoin, keyRole, unsignedTx.Info.NextKeyID); err != nil {
				k.Logger(ctx).Error(sdkerrors.Wrap(err, fmt.Sprintf("failed to assign the next %s key to %s", keyRole.SimpleString(), nextKey.ID)).Error())
				continue
			}

			ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeKey,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueAssigned),
				sdk.NewAttribute(types.AttributeKeyRole, keyRole.SimpleString()),
				sdk.NewAttribute(types.AttributeKeyKeyID, unsignedTx.Info.NextKeyID),
			))
		}

		k.DeleteUnsignedTx(ctx, keyRole)
		k.SetSignedTx(ctx, keyRole, types.NewSignedTx(signedTx, unsignedTx.ConfirmationRequired, unsignedTx.AnyoneCanSpendVout))
		k.SetLatestSignedTxHash(ctx, keyRole, txHash)

		// Notify that consolidation tx can be queried
		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueSigned),
			sdk.NewAttribute(types.AttributeKeyRole, keyRole.SimpleString()),
		))
		k.Logger(ctx).Info(fmt.Sprintf("transaction %s is fully signed", txHash.String()))
	}

	return nil
}

func assembleTx(ctx sdk.Context, k types.BTCKeeper, signer types.Signer, unsignedTx *types.UnsignedTx) (*wire.MsgTx, bool, error) {
	tx := unsignedTx.GetTx()
	outPointsToSign, err := getOutPointsToSign(ctx, tx, k)
	if err != nil {
		return nil, false, sdkerrors.Wrapf(err, "failed to collect outpoints waiting to be signed for unsigned tx %s", tx.TxHash().String())
	}

	// Assemble transaction with signatures
	var sigs [][]btcec.Signature
	for _, inputInfo := range unsignedTx.Info.InputInfos {
		var sigsForOutPoint []btcec.Signature

		for _, sigRequirement := range inputInfo.SigRequirements {
			sigHashHex := hex.EncodeToString(sigRequirement.SigHash)
			sigID := fmt.Sprintf("%s-%s", sigHashHex, sigRequirement.KeyID)
			sig, ok := signer.GetSig(ctx, sigID)
			if !ok {
				err := fmt.Errorf("signature for tx %s not yet found", sigID)

				// TODO: keyID for sigID is deleted on signing failure/timeout. Some more explicit state is needed.
				if _, ok := signer.GetKeyForSigID(ctx, sigID); !ok {
					return nil, true, err
				}

				return nil, false, err
			}

			sigsForOutPoint = append(sigsForOutPoint, btcec.Signature{R: sig.R, S: sig.S})
		}

		sigs = append(sigs, sigsForOutPoint)
	}

	signedTx, err := types.AssembleBtcTx(tx, outPointsToSign, sigs)
	if err != nil {
		return nil, false, err
	}

	return signedTx, false, nil
}

func getKnownOutPoints(ctx sdk.Context, k types.BTCKeeper, signedTx *wire.MsgTx) ([]types.OutPointInfo, error) {
	var knownOutPoints []types.OutPointInfo

	networkName := k.GetNetwork(ctx).Name
	network, err := types.NetworkFromStr(networkName)
	if err != nil {
		return nil, sdkerrors.Wrap(err, fmt.Sprintf("failed to get network %s", networkName))
	}

	txHash := signedTx.TxHash()
	// Confirm all outpoints that axelar controls the keys of
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

		if state != types.SPENT {
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
