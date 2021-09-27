package types

import (
	"fmt"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRefundMessageRequest creates a message of type RefundMessageRequest
func NewRefundMessageRequest(sender sdk.AccAddress, innerMessage sdk.Msg) *RefundMessageRequest {
	messageAny, err := cdctypes.NewAnyWithValue(innerMessage)
	if err != nil {
		panic(err)
	}
	return &RefundMessageRequest{
		Sender:       sender,
		InnerMessage: messageAny,
	}
}

// Route returns the route for this message
func (m RefundMessageRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RefundMessageRequest) Type() string {
	return "RefundMessageRequest"
}

// ValidateBasic executes a stateless message validation
func (m RefundMessageRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.InnerMessage == nil {
		return fmt.Errorf("missing inner message")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RefundMessageRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m RefundMessageRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m RefundMessageRequest) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if m.InnerMessage != nil {
		var sdkMsg sdk.Msg
		return unpacker.UnpackAny(m.InnerMessage, &sdkMsg)
	}
	return nil
}

// GetInnerMessage unwrap the inner message
func (m RefundMessageRequest) GetInnerMessage() sdk.Msg {
	innerMsg, ok := m.InnerMessage.GetCachedValue().(sdk.Msg)
	if !ok {
		return nil
	}
	return innerMsg
}
