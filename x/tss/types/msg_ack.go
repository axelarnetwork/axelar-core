package types

import (
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Route implements the sdk.Msg interface.
func (m AckRequest) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m AckRequest) Type() string { return "KeygenTraffic" }

// ValidateBasic implements the sdk.Msg interface.
func (m AckRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.ID == "" {
		return sdkerrors.Wrap(ErrTss, "id must be set")
	}

	if m.AckType != exported.AckKeygen && m.AckType != exported.AckSign {
		return sdkerrors.Wrapf(ErrTss, "ack type must be either '%s' or '%s'",
			exported.AckKeygen.String(), exported.AckSign.String())
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
