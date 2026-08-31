package ante

import (
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
)

// LimitFeeDenomDecorator restricts which denominations may be used to pay
// transaction fees. It enforces two rules on every transaction:
//
//  1. a transaction may carry at most one fee denomination and
//  2. its denomination must be on the allowlist
//
// Transactions with no fee are intentionally allowed through, needed for gas simulation and zero-fee genesis gentxs.
//
// The allowlist is read from a governance-controlled parameter (the x/feepolicy
// module's AllowedFeeDenoms param), so it can be updated by governance without a
// binary upgrade.
type LimitFeeDenomDecorator struct {
	feePolicy types.FeePolicy
}

// NewLimitFeeDenomDecorator returns a LimitFeeDenomDecorator that reads its
// allowlist from the given fee policy source.
func NewLimitFeeDenomDecorator(feePolicy types.FeePolicy) LimitFeeDenomDecorator {
	return LimitFeeDenomDecorator{feePolicy: feePolicy}
}

// AnteHandle enforces the single-denomination, allowlisted fee policy.
func (d LimitFeeDenomDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "tx must be a FeeTx")
	}

	fees := feeTx.GetFee()

	if len(fees) > 1 {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins,
			"a transaction may pay fees in at most one denomination, got %d: %s", len(fees), fees)
	}

	if len(fees) == 1 {
		allowed := d.feePolicy.GetAllowedFeeDenoms(ctx)
		if !slices.Contains(allowed, fees[0].Denom) {
			return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins,
				"fee denomination %q is not allowed", fees[0].Denom)
		}
	}

	return next(ctx, tx, simulate)
}
