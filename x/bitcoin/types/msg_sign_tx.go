package types

import (
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewSignTxRequest is the constructor for SignTxRequest
func NewSignTxRequest(sender sdk.AccAddress, keyRole tss.KeyRole) *SignTxRequest {
	return &SignTxRequest{
		Sender:  sender,
		KeyRole: keyRole,
	}
}

// Route returns the route for this message
func (m SignTxRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m SignTxRequest) Type() string {
	return "SignTx"
}

// ValidateBasic executes a stateless message validation
func (m SignTxRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.KeyRole.Validate(); err != nil {
		return sdkerrors.Wrap(ErrBitcoin, err.Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m SignTxRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m SignTxRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
