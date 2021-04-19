package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Route implements sdk.Msg
func (m MsgLink) Route() string {
	return RouterKey
}

// Type  implements sdk.Msg
func (m MsgLink) Type() string {
	return "Link"
}

// ValidateBasic implements sdk.Msg
func (m MsgLink) ValidateBasic() error {
	if m.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if m.RecipientAddr == "" {
		return fmt.Errorf("missing recipient address")
	}
	if m.RecipientChain == "" {
		return fmt.Errorf("missing recipient chain")
	}

	if m.Symbol == "" {
		return fmt.Errorf("missing asset symbol")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m MsgLink) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m MsgLink) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
