package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/utils/slices"
)

// NewBatchRequest is the constructor for BatchRequest
func NewBatchRequest(sender sdk.AccAddress, messages []sdk.Msg) *BatchRequest {
	f := func(msg sdk.Msg) types.Any {
		messageAny, err := cdctypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
		return *messageAny
	}

	return &BatchRequest{
		Sender:   sender,
		Messages: slices.Map(messages, f),
	}
}

// Route returns the route for this message
func (m BatchRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m BatchRequest) Type() string {
	return "Batch"
}

// ValidateBasic executes a stateless message validation
func (m BatchRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if len(m.Messages) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty batch")
	}

	return nil
}

// GetSigners returns the set of signers for this message
func (m BatchRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

func (m BatchRequest) ValidateInnerMessage(msg sdk.Msg) error {
	if !msg.GetSigners()[0].Equals(m.Sender) {
		return fmt.Errorf("message signer %s does not match batch signer %s", msg.GetSigners()[0], m.Sender)
	}

	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
