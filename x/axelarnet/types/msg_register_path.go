package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

// NewRegisterIbcPathRequest creates a message of type RegisterIbcPathRequest
func NewRegisterIbcPathRequest(sender sdk.AccAddress, asset, path string) *RegisterIbcPathRequest {
	return &RegisterIbcPathRequest{
		Sender: sender,
		Asset:  asset,
		Path:   path,
	}
}

// Route returns the route for this message
func (m RegisterIbcPathRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RegisterIbcPathRequest) Type() string {
	return "RegisterIbcPath"
}

// ValidateBasic executes a stateless message validation
func (m RegisterIbcPathRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.Asset == "" {
		return fmt.Errorf("missing asset")
	}

	f := host.NewPathValidator(func(path string) error {
		return nil
	})
	if err := f(m.Path); err != nil {
		return sdkerrors.Wrap(err, "invalid path")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RegisterIbcPathRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m RegisterIbcPathRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
