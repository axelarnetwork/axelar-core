package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHaltChainRequest is the constructor for HaltChainRequest
func NewHaltChainRequest(sender sdk.AccAddress, reason string, overridePendingUpgrade bool) *HaltChainRequest {
	return &HaltChainRequest{
		Sender:                 sender.String(),
		Reason:                 reason,
		OverridePendingUpgrade: overridePendingUpgrade,
	}
}

// Route returns the route for this message
func (m HaltChainRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m HaltChainRequest) Type() string {
	return "HaltChain"
}

// ValidateBasic executes a stateless message validation
func (m HaltChainRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	return validateReason(m.Reason)
}

// GetSignBytes returns the message bytes that need to be signed
func (m HaltChainRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
