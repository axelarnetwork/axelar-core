package types

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

type MsgRawTx struct {
	Sender      sdk.AccAddress
	TxHash      *common.Hash
	Amount      big.Int
	Data        []byte
	Destination EthAddress
	TXType      TXType
	Mode        Mode
}

func NewMsgRawTx(sender sdk.AccAddress, txHash *common.Hash, amount big.Int, data []byte, destination EthAddress, txType TXType) sdk.Msg {
	return MsgRawTx{
		Sender:      sender,
		TxHash:      txHash,
		Amount:      amount,
		Data:        data,
		Destination: destination,
		TXType:      txType,
		Mode:        ModeSpecificAddress,
	}
}

func NewMsgRawTxForNextMasterKey(sender sdk.AccAddress, txHash *common.Hash, amount big.Int, data []byte, txType TXType) sdk.Msg {
	return MsgRawTx{
		Sender: sender,
		TxHash: txHash,
		Amount: amount,
		Data:   data,
		TXType: txType,
		Mode:   ModeNextMasterKey,
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
	if msg.TxHash == nil {
		return fmt.Errorf("missing transaction ID")
	}
	if msg.Amount.Int64() <= 0 {
		return fmt.Errorf("transaction amount must be greater than zero")
	}
	if err := msg.Destination.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid destination")
	}

	if !msg.TXType.IsValid() {
		return fmt.Errorf("Invalid transaction type")
	}

	if !msg.Mode.IsValid() {
		return fmt.Errorf("Invalid transaction mode")
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
