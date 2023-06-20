package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/bank/types"
)

// BankKeeper wraps the bankkeeper.BaseKeeper.
type BankKeeper struct {
	types.BankKeeper
}

// NewBankKeeper returns a new BankKeeper.
func NewBankKeeper(bk types.BankKeeper) BankKeeper {
	return BankKeeper{bk}
}

// SendCoins transfers amt coins from a sending account to a receiving account.
// An error is returned upon failure, or when the from/to address is blocked.
func (k BankKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	if k.BlockedAddr(fromAddr) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to send funds", fromAddr)
	}

	if k.BlockedAddr(toAddr) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", toAddr)
	}

	return k.BankKeeper.SendCoins(ctx, fromAddr, toAddr, amt)
}
