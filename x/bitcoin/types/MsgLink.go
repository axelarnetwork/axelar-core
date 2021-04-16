package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgLink - MsgLink constructor
func NewMsgLink(sender sdk.AccAddress, recipientAddr string, recipientChain string) *MsgLink {
	return &MsgLink{
		Sender:         sender.String(),
		RecipientAddr:  recipientAddr,
		RecipientChain: recipientChain,
	}
}

// Route returns the route for this message
func (m MsgLink) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m MsgLink) Type() string {
	return "Link"
}

// ValidateBasic executes a stateless message validation
func (m MsgLink) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if m.RecipientAddr == "" {
		return fmt.Errorf("missing recipient address")
	}
	if m.RecipientChain == "" {
		return fmt.Errorf("missing recipient chain")
	}
	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m MsgLink) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m MsgLink) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
