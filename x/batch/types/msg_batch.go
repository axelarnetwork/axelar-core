package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/utils/slices"
)

var _ types.UnpackInterfacesMessage = BatchRequest{}

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

	for _, msg := range m.UnwrapMessages() {
		if !msg.GetSigners()[0].Equals(m.Sender) {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "message signer mismatch")
		}

		if err := msg.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

// GetSigners returns the set of signers for this message
func (m BatchRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m BatchRequest) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	for _, msg := range m.Messages {
		var sdkMsg sdk.Msg
		if err := unpacker.UnpackAny(&msg, &sdkMsg); err != nil {
			return err
		}
	}

	return nil
}

// UnwrapMessages unwrap the batch messages
func (m BatchRequest) UnwrapMessages() []sdk.Msg {
	return slices.Map(m.Messages, func(msg types.Any) sdk.Msg {
		return msg.GetCachedValue().(sdk.Msg)
	})
}
