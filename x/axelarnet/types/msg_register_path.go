package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/ibc-go/modules/core/24-host"
)

// NewRegisterIBCPathRequest creates a message of type RegisterIBCPathRequest
func NewRegisterIBCPathRequest(sender sdk.AccAddress, chain, path string) *RegisterIBCPathRequest {
	return &RegisterIBCPathRequest{
		Sender: sender,
		Chain:  chain,
		Path:   path,
	}
}

// Route returns the route for this message
func (m RegisterIBCPathRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RegisterIBCPathRequest) Type() string {
	return "RegisterIBCPath"
}

// ValidateBasic executes a stateless message validation
func (m RegisterIBCPathRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.Chain == "" {
		return fmt.Errorf("missing chain")
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
func (m RegisterIBCPathRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m RegisterIBCPathRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
