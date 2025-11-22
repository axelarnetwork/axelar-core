package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewRegisterIBCPathRequest creates a message of type RegisterIBCPathRequest
func NewRegisterIBCPathRequest(sender sdk.AccAddress, chain, path string) *RegisterIBCPathRequest {
	return &RegisterIBCPathRequest{
		Sender: sender.String(),
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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
	}

	if err := ValidateIBCPath(m.Path); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RegisterIBCPathRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
