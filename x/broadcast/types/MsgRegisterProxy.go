package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewMsgRegisterProxy(principal sdk.ValAddress, proxy sdk.AccAddress) *MsgRegisterProxy {
	return &MsgRegisterProxy{
		PrincipalAddr: principal.String(),
		ProxyAddr:     proxy.String(),
	}
}

func (msg MsgRegisterProxy) Route() string {
	return RouterKey
}

func (msg MsgRegisterProxy) Type() string {
	return "RegisterProxy"
}

func (msg MsgRegisterProxy) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(msg.PrincipalAddr); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "malformed principal address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.ProxyAddr); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "malformed proxy address")
	}

	return nil
}

func (msg MsgRegisterProxy) GetPrincipal() sdk.ValAddress {
	addr, err := sdk.ValAddressFromBech32(msg.PrincipalAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgRegisterProxy) GetProxy() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.ProxyAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgRegisterProxy) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRegisterProxy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetProxy()}
}
