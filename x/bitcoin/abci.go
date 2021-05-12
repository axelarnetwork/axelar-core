package bitcoin

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

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

	tx, ok := k.GetUnsignedTx(ctx)
	if !ok {
		k.Logger(ctx).Debug("no unsigned transaction ready")
		return nil
	}

	outpointsToSign, err := getOutPointsToSign(ctx, tx, k)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrapf(err, "failed to collect outpoints waiting to be signed for unsigned tx %s", tx.TxHash().String()).Error())
	}

	k.Logger(ctx).Debug("checking for completed signatures")

	var sigs []btcec.Signature
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
		sigs = append(sigs, btcec.Signature{R: sig.R, S: sig.S})
	}

	tx, err = types.AssembleBtcTx(tx, outpointsToSign, sigs)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed to assemble Bitcoin tx signatures").Error())
		return nil
	}

	k.DeleteUnsignedTx(ctx)
	k.SetSignedTx(ctx, tx)

	// Notify that consolidation tx can be queried
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeTransactionSigned,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyTxHash, tx.TxHash().String()),
	))

	k.Logger(ctx).Info(fmt.Sprintf("transaction %s is fully signed", tx.TxHash().String()))

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
