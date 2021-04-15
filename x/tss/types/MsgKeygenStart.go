package types

import (
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

<<<<<<< HEAD
// NewMsgKeygenStart constructor for MsgKeygenStart
func NewMsgKeygenStart(sender sdk.AccAddress, newKeyID string, subsetSize int64) *MsgKeygenStart {
	return &MsgKeygenStart{
		Sender:     sender,
		NewKeyID:   newKeyID,
		SubsetSize: subsetSize,
=======
// MsgKeygenStart indicate the start of keygen
type MsgKeygenStart struct {
	Sender                     sdk.AccAddress
	NewKeyID                   string
	SubsetSize                 int64
	KeyShareDistributionPolicy exported.KeyShareDistributionPolicy
}

// NewMsgKeygenStart constructor for MsgKeygenStart
func NewMsgKeygenStart(sender sdk.AccAddress, newKeyID string, subsetSize int64, keyShareDistributionPolicy exported.KeyShareDistributionPolicy) sdk.Msg {
	return MsgKeygenStart{
		Sender:                     sender,
		NewKeyID:                   newKeyID,
		SubsetSize:                 subsetSize,
		KeyShareDistributionPolicy: keyShareDistributionPolicy,
>>>>>>> 3760186... add KeyShareDistributionPolicy to keygen
	}
}

// Route implements the sdk.Msg interface.
func (m MsgKeygenStart) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m MsgKeygenStart) Type() string { return "KeyGenStart" }

// ValidateBasic implements the sdk.Msg interface.
func (m MsgKeygenStart) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.NewKeyID == "" {
		return sdkerrors.Wrap(ErrTss, "new key id must be set")
	}

	if m.SubsetSize < 0 {
		return sdkerrors.Wrap(ErrTss, "subset size has to be greater than or equal to 0")
	}

<<<<<<< HEAD
	// TODO enforce a maximum length for m.NewKeyID?
=======
	if err := msg.KeyShareDistributionPolicy.Validate(); err != nil {
		return nil
	}

	// TODO enforce a maximum length for msg.NewKeyID?
>>>>>>> 3760186... add KeyShareDistributionPolicy to keygen
	return nil
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgKeygenStart) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements sdk.Msg
func (m MsgKeygenStart) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
