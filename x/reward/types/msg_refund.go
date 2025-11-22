package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/reward/exported"
)

// NewRefundMsgRequest creates a message of type RefundMsgRequest
func NewRefundMsgRequest(sender sdk.AccAddress, innerMessage sdk.Msg) *RefundMsgRequest {
	messageAny, err := cdctypes.NewAnyWithValue(innerMessage)
	if err != nil {
		panic(err)
	}
	return &RefundMsgRequest{
		Sender:       sender.String(),
		InnerMessage: messageAny,
	}
}

// Route returns the route for this message
func (m RefundMsgRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RefundMsgRequest) Type() string {
	return "RefundMsgRequest"
}

// ValidateBasic executes a stateless message validation
func (m RefundMsgRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if m.InnerMessage == nil {
		return fmt.Errorf("missing inner message")
	}

	innerMessage := m.GetInnerMessage()
	if innerMessage == nil {
		return fmt.Errorf("invalid inner message")
	}

	msg, ok := innerMessage.(sdk.HasValidateBasic)
	if !ok {
		return fmt.Errorf("inner message %T does not implement HasValidateBasic", innerMessage)
	}
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RefundMsgRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m RefundMsgRequest) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if m.InnerMessage != nil {
		var refundableMsg exported.Refundable
		return unpacker.UnpackAny(m.InnerMessage, &refundableMsg)
	}
	return nil
}

// GetInnerMessage unwrap the inner message
func (m RefundMsgRequest) GetInnerMessage() exported.Refundable {
	innerMsg, ok := m.InnerMessage.GetCachedValue().(exported.Refundable)
	if !ok {
		return nil
	}
	return innerMsg
}
