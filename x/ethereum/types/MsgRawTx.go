package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

type MsgRawTx struct {
	Sender      sdk.AccAddress
	TXType      TXType
	Amount      sdk.Int
	Destination common.Address
	Network     Network
	ContractID  string
}

func NewMsgRawTxForMint(sender sdk.AccAddress, network Network, contractID string, amount sdk.Int, destination common.Address) sdk.Msg {
	return MsgRawTx{
		Sender:      sender,
		Network:     network,
		Amount:      amount,
		Destination: destination,
		ContractID:  contractID,
		TXType:      TypeERC20mint,
	}
}

func NewMsgRawTxForDeploy(sender sdk.AccAddress, network Network, contractID string) sdk.Msg {
	return MsgRawTx{
		Sender:     sender,
		Network:    network,
		ContractID: contractID,
		TXType:     TypeSCDeploy,
	}
}

func (msg MsgRawTx) Route() string {
	return RouterKey
}

func (msg MsgRawTx) Type() string {
	return "RawTx"
}

func (msg MsgRawTx) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if err := msg.Network.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid network")
	}
	if msg.ContractID == "" {
		return fmt.Errorf("missing contract ID")
	}

	switch msg.TXType {
	case TypeERC20mint:
		if msg.Amount.Int64() <= 0 {
			return fmt.Errorf("transaction amount must be greater than zero")
		}
	case TypeSCDeploy:
	default:
		return fmt.Errorf("wrong tx type")
	}
	return nil
}

func (msg MsgRawTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgRawTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
