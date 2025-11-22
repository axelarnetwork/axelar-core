package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewRetryFailedEventRequest - RetryFailedEventRequest constructor
func NewRetryFailedEventRequest(sender sdk.AccAddress, chain string, eventID string) *RetryFailedEventRequest {
	return &RetryFailedEventRequest{
		Sender:  sender.String(),
		Chain:   nexus.ChainName(utils.NormalizeString(chain)),
		EventID: EventID(utils.NormalizeString(eventID)),
	}
}

// Route returns the route for this message
func (m RetryFailedEventRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RetryFailedEventRequest) Type() string {
	return "RetryFailedEvent"
}

// ValidateBasic executes a stateless message validation
func (m RetryFailedEventRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
	}

	if err := m.EventID.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid eventID")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RetryFailedEventRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
