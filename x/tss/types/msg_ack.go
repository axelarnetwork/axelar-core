package types

import (
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewAckRequest constructor for AckRequest
func NewAckRequest(sender sdk.AccAddress, ID string, ackType exported.AckType, height int64) *AckRequest {
	return &AckRequest{
		Sender:  sender,
		ID:      ID,
		AckType: ackType,
		Height:  height,
	}
}

// Route implements the sdk.Msg interface.
func (m AckRequest) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m AckRequest) Type() string { return "Ack" }

// ValidateBasic implements the sdk.Msg interface.
func (m AckRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.ID == "" {
		return sdkerrors.Wrap(ErrTss, "id must be set")
	}

	if m.AckType != exported.AckType_Keygen && m.AckType != exported.AckType_Sign {
		return sdkerrors.Wrapf(ErrTss, "ack type must be either '%s' or '%s'",
			exported.AckType_Keygen.String(), exported.AckType_Sign.String())
	}

	if m.Height < 0 {
		return sdkerrors.Wrap(ErrTss, "invalid height")
	}

	return nil
}

// GetSignBytes implements the sdk.Msg interface
func (m AckRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements the sdk.Msg interface
func (m AckRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
