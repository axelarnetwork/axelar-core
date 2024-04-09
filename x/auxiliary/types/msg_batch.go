package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

var (
	_ sdk.Msg = &BatchRequest{}

	_ cdctypes.UnpackInterfacesMessage = &BatchRequest{}
)

// NewBatchRequest is the constructor for BatchRequest
func NewBatchRequest(sender sdk.AccAddress, messages []sdk.Msg) *BatchRequest {
	return &BatchRequest{
		Sender:   sender,
		Messages: slices.Map(messages, func(msg sdk.Msg) cdctypes.Any { return *funcs.Must(cdctypes.NewAnyWithValue(msg)) }),
	}
}

// ValidateBasic executes a stateless message validation
func (m BatchRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if len(m.Messages) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty batch")
	}

	if anyBatch(m.UnwrapMessages()) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "nested batch requests are not allowed")
	}

	for _, msg := range m.UnwrapMessages() {
		if !equalAccAddresses(msg.GetSigners(), m.GetSigners()) {
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
	return slices.Map(m.Messages, func(msg cdctypes.Any) sdk.Msg {
		return msg.GetCachedValue().(sdk.Msg)
	})
}

// equalAccAddresses checks the equality of two slices of sdk.AccAddress
func equalAccAddresses(first, second []sdk.AccAddress) bool {
	if len(first) != len(second) {
		return false
	}

	for i := range first {
		if !first[i].Equals(second[i]) {
			return false
		}
	}

	return true
}

// anyBatch checks if any of the messages are a BatchRequest
func anyBatch(msgs []sdk.Msg) bool {
	return slices.Any(msgs, func(msg sdk.Msg) bool {
		_, ok := msg.(*BatchRequest)
		return ok
	})
}
