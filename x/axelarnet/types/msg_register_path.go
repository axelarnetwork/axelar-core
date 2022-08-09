package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/ibc-go/v2/modules/core/24-host"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewRegisterIBCPathRequest creates a message of type RegisterIBCPathRequest
func NewRegisterIBCPathRequest(sender sdk.AccAddress, chain, path string) *RegisterIBCPathRequest {
	return &RegisterIBCPathRequest{
		Sender: sender,
		Chain:  nexus.ChainName(utils.NormalizeString(chain)),
		Path:   utils.NormalizeString(path),
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

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := utils.ValidateString(m.Path); err != nil {
		return sdkerrors.Wrap(err, "invalid path")
	}
	f := host.NewPathValidator(func(path string) error {
		return nil
	})
	if err := f(m.Path); err != nil {
		return sdkerrors.Wrap(err, "invalid path")
	}

	// we only support direct IBC connections
	pathSplit := strings.Split(m.Path, "/")
	if len(pathSplit) != 2 {
		return fmt.Errorf(fmt.Sprintf("invalid IBC path %s", m.Path))
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
