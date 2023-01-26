package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewExecuteMessage creates a message of type ExecuteMessageRequest
func NewExecuteMessage(sender sdk.AccAddress, chain nexus.ChainName, id string, payload []byte) *ExecuteMessageRequest {
	return &ExecuteMessageRequest{
		Sender:  sender,
		Chain:   chain,
		ID:      id,
		Payload: payload,
	}
}

// Route returns the route for this message
func (m ExecuteMessageRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m ExecuteMessageRequest) Type() string {
	return "ExecuteGeneralMessage"
}

// ValidateBasic executes a stateless message validation
func (m ExecuteMessageRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return err
	}

	if err := utils.ValidateString(m.ID); err != nil {
		return err
	}

	if len(m.Payload) == 0 {
		return fmt.Errorf("invalid payload")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m ExecuteMessageRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m ExecuteMessageRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
