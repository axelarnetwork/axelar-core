package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

// NewRouteMessage creates a message of type RouteMessageRequest
func NewRouteMessage(sender sdk.AccAddress, feegranter sdk.AccAddress, id string, payload []byte) *RouteMessageRequest {
	return &RouteMessageRequest{
		Sender:     sender.String(),
		ID:         id,
		Payload:    payload,
		Feegranter: feegranter,
	}
}

// Route returns the route for this message
func (m RouteMessageRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RouteMessageRequest) Type() string {
	return "RouteMessage"
}

// ValidateBasic executes a stateless message validation
func (m RouteMessageRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := utils.ValidateString(m.ID); err != nil {
		return errorsmod.Wrap(err, "invalid general message id")
	}

	if m.Feegranter != nil {
		if err := sdk.VerifyAddressFormat(m.Feegranter); err != nil {
			return errorsmod.Wrap(err, "invalid feegranter")
		}
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RouteMessageRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
