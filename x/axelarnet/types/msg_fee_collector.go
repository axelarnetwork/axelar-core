package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRegisterFeeCollectorRequest is the constructor for RegisterFeeCollector
func NewRegisterFeeCollectorRequest(sender sdk.AccAddress, feeCollector sdk.AccAddress) *RegisterFeeCollectorRequest {
	return &RegisterFeeCollectorRequest{
		Sender:       sender.String(),
		FeeCollector: feeCollector,
	}
}

// Route returns the route for this message
func (m RegisterFeeCollectorRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RegisterFeeCollectorRequest) Type() string {
	return "RegisterFeeCollector"
}

// ValidateBasic executes a stateless message validation
func (m RegisterFeeCollectorRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := sdk.VerifyAddressFormat(m.FeeCollector); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "fee collector").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RegisterFeeCollectorRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
