package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgRegisterProxy - MsgRegisterProxy constructor
func NewMsgRegisterProxy(principal sdk.ValAddress, proxy sdk.AccAddress) *MsgRegisterProxy {
	return &MsgRegisterProxy{
		PrincipalAddr: principal,
		ProxyAddr:     proxy,
	}
}

// Route returns the route for this message
func (m MsgRegisterProxy) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m MsgRegisterProxy) Type() string {
	return "RegisterProxy"
}

// ValidateBasic executes a stateless message validation
func (m MsgRegisterProxy) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.PrincipalAddr); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "principal").Error())
	}
	if err := sdk.VerifyAddressFormat(m.ProxyAddr); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "proxy").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m MsgRegisterProxy) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m MsgRegisterProxy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(m.PrincipalAddr)}
}
