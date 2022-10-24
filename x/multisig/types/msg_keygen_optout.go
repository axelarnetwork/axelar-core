package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &KeygenOptOutRequest{}

// ValidateBasic implements the sdk.Msg interface.
func (m *KeygenOptOutRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	return nil
}

// GetSigners implements the sdk.Msg interface.
func (m *KeygenOptOutRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
