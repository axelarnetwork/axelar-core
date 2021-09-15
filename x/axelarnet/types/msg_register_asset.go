package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRegisterAssetRequest is the constructor for RegisterAssetRequest
func NewRegisterAssetRequest(sender sdk.AccAddress, chain, denom string) *RegisterAssetRequest {
	return &RegisterAssetRequest{
		Sender: sender,
		Chain:  chain,
		Denom:  denom,
	}
}

// Route returns the route for this message
func (m RegisterAssetRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RegisterAssetRequest) Type() string {
	return "RegisterAsset"
}

// ValidateBasic executes a stateless message validation
func (m RegisterAssetRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.Chain == "" {
		return fmt.Errorf("missing chain name")
	}

	if m.Denom == "" {
		return fmt.Errorf("missing asset name")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RegisterAssetRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m RegisterAssetRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
