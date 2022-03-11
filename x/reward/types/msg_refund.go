package types

import (
	"fmt"

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
		Sender:       sender,
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
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.InnerMessage == nil {
		return fmt.Errorf("missing inner message")
	}

	innerMessage := m.GetInnerMessage()
	if innerMessage == nil {
		return fmt.Errorf("invalid inner message")
	}

	if err := innerMessage.ValidateBasic(); err != nil {
		return err
	}

	signers := innerMessage.GetSigners()

	if len(signers) != 1 {
		return fmt.Errorf("invalid number of signers for inner message")
	}

	if !m.GetSigners()[0].Equals(signers[0]) {
		return fmt.Errorf("refund msg and inner message signers do not match")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RefundMsgRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m RefundMsgRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
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
