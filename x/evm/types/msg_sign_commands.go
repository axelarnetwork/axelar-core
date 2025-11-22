package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewSignCommandsRequest creates a message of type SignCommandsRequest
func NewSignCommandsRequest(sender sdk.AccAddress, chain string) *SignCommandsRequest {
	return &SignCommandsRequest{
		Sender: sender.String(),
		Chain:  nexus.ChainName(utils.NormalizeString(chain)),
	}
}

// Route implements sdk.Msg
func (m SignCommandsRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m SignCommandsRequest) Type() string {
	return "SignCommands"
}

// ValidateBasic implements sdk.Msg
func (m SignCommandsRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m SignCommandsRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
