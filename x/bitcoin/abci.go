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
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ types.BTCKeeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, k types.BTCKeeper, signer types.Signer) []abci.ValidatorUpdate {
	if req.Height%k.GetSigCheckInterval(ctx) != 0 {
		return nil
	}

	unsignedTx, ok := k.GetUnsignedTx(ctx)
	if !ok {
		k.Logger(ctx).Debug("no unsigned transaction ready")
		return nil
	}

	tx := unsignedTx.GetTx()
	outpointsToSign, err := getOutPointsToSign(ctx, tx, k)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrapf(err, "failed to collect outpoints waiting to be signed for unsigned tx %s", tx.TxHash().String()).Error())
		return nil
	}

	k.Logger(ctx).Debug("checking for completed signatures")

	// Assemble transaction with signatures
	var sigs [][]btcec.Signature
	for i, in := range outpointsToSign {
		hash, err := txscript.CalcWitnessSigHash(in.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, i, int64(in.Amount))
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("could not calculate sig hash for outpoint %s", in.OutPoint))
			return nil
		}

		sigID := hex.EncodeToString(hash)
		sig, ok := signer.GetSig(ctx, sigID)
		if !ok {
			k.Logger(ctx).Debug(fmt.Sprintf("signature for tx %s not yet found", sigID))
			return nil
		}
		// TODO: handle multiple signatures per input
		sigs = append(sigs, []btcec.Signature{{R: sig.R, S: sig.S}})
	}

	tx, err = types.AssembleBtcTx(tx, outpointsToSign, sigs)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed to assemble Bitcoin tx signatures").Error())
		return nil
	}

	networkName := k.GetNetwork(ctx).Name
	network, err := types.NetworkFromStr(networkName)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, fmt.Sprintf("failed to get network %s", networkName)).Error())
		return nil
	}

	txHash := tx.TxHash()

	// Confirm all outpoints that axelar controls the keys of
	for i, output := range tx.TxOut {
		_, addresses, _, err := txscript.ExtractPkScriptAddrs(output.PkScript, network.Params())
		if err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, fmt.Sprintf("failed to extract change output address from transaction %s", txHash.String())).Error())
			return nil
		}

		if len(addresses) != 1 {
			continue
		}

		address, ok := k.GetAddress(ctx, addresses[0].EncodeAddress())
		if !ok {
			continue
		}

		outpointInfo := types.NewOutPointInfo(wire.NewOutPoint(&txHash, uint32(i)), btcutil.Amount(output.Value), address.Address)
		k.SetConfirmedOutpointInfo(ctx, address.KeyID, outpointInfo)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(types.EventTypeOutpointConfirmation,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyOutPointInfo, string(types.ModuleCdc.MustMarshalJSON(&outpointInfo))),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
			),
		)
	}

	// Assign the next key if necessary
	if unsignedTx.AssignNextKey {
		nextKey, ok := signer.GetKey(ctx, unsignedTx.NextKeyID)
		if !ok {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, fmt.Sprintf("failed to get next key %s to assign", unsignedTx.NextKeyID)).Error())
			return nil
		}

		if err := signer.AssignNextKey(ctx, exported.Bitcoin, unsignedTx.NextKeyRole, unsignedTx.NextKeyID); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, fmt.Sprintf("failed to assign the next %s key to %s", nextKey.Role.SimpleString(), nextKey.ID)).Error())
			return nil
		}
	}

	k.DeleteUnsignedTx(ctx)
	k.SetSignedTx(ctx, tx)

	// Notify that consolidation tx can be queried
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeTransactionSigned,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyTxHash, txHash.String()),
	))

	k.Logger(ctx).Info(fmt.Sprintf("transaction %s is fully signed", txHash.String()))

	return nil
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
