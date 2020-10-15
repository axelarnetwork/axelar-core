package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Ensure MsgTrackAddress implements sdk.Msg interface
var _ sdk.Msg = &MsgRegisterProxy{}

type MsgRegisterProxy struct {
	Principal sdk.ValAddress
	Proxy     sdk.AccAddress
}

func NewMsgRegisterProxy(principal sdk.ValAddress, proxy sdk.AccAddress) MsgRegisterProxy {
	return MsgRegisterProxy{
		Principal: principal,
		Proxy:     proxy,
	}
}

func (msg MsgRegisterProxy) Route() string {
	return RouterKey
}

func (msg MsgRegisterProxy) Type() string {
	return "RegisterProxy"
}

func (msg MsgRegisterProxy) ValidateBasic() error {
	if msg.Principal.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing validator principal")
	}
	if msg.Proxy.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing broadcast proxy")
	}

	return nil
}

func (msg MsgRegisterProxy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgRegisterProxy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.Principal)}
}
