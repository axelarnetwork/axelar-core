package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRetryFailedMessageRequest creates a message of type RetryFailedMessageRequest
func NewRetryFailedMessageRequest(sender sdk.AccAddress, id string) *RetryFailedMessageRequest {
	return &RetryFailedMessageRequest{
		Sender: sender.String(),
		ID:     id,
	}
}

// ValidateBasic implements sdk.Msg
func (m RetryFailedMessageRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if m.ID == "" {
		return fmt.Errorf("missing message ID")
	}

	return nil
}
