package bitcoin

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ keeper.Keeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, k keeper.Keeper, signer types.Signer) []abci.ValidatorUpdate {
	if req.Height%k.GetSigCheckInterval(ctx) != 0 {
		return nil
	}

	tx := k.GetRawTx(ctx)
	if tx == nil {
		return nil
	}

	hashes, err := k.GetHashesToSign(ctx, tx)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed to check for Bitcoin tx signatures").Error())
		return nil
	}

	k.Logger(ctx).Debug("checking for completed signatures")

	var sigs []btcec.Signature
	for _, hash := range hashes {
		sigID := hex.EncodeToString(hash)
		sig, ok := signer.GetSig(ctx, sigID)
		if !ok {
			k.Logger(ctx).Debug(fmt.Sprintf("signature for tx %s not yet found", sigID))
			return nil
		}
		sigs = append(sigs, btcec.Signature{R: sig.R, S: sig.S})
	}

	tx, err = k.AssembleBtcTx(ctx, tx, sigs)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed to assemble Bitcoin tx signatures").Error())
		return nil
	}

	k.SetSignedTx(ctx, tx)
	k.DeleteRawTx(ctx)

	return nil
}
