package types

import (
	errorsmod "cosmossdk.io/errors"
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
		Sender:   sender.String(),
		Messages: slices.Map(messages, func(msg sdk.Msg) cdctypes.Any { return *funcs.Must(cdctypes.NewAnyWithValue(msg)) }),
	}
}

// ValidateBasic executes a stateless message validation
func (m BatchRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if len(m.Messages) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty batch")
	}

	if anyBatch(m.UnwrapMessages()) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "nested batch requests are not allowed")
	}

	for _, msg := range m.UnwrapMessages() {
		m, ok := msg.(sdk.HasValidateBasic)
		if ok {
			if err := m.ValidateBasic(); err != nil {
				return err
			}
		}

	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m *BatchRequest) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	for i := range m.Messages {
		var sdkMsg sdk.Msg
		if err := unpacker.UnpackAny(&m.Messages[i], &sdkMsg); err != nil {
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

// anyBatch checks if any of the messages are a BatchRequest
func anyBatch(msgs []sdk.Msg) bool {
	return slices.Any(msgs, func(msg sdk.Msg) bool {
		_, ok := msg.(*BatchRequest)
		return ok
	})
}
