package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgRawTxForMasterKey struct {
	Sender sdk.AccAddress
	TxHash *chainhash.Hash
	Amount btcutil.Amount
	Chain  Chain
}

func NewMsgRawTxForMasterKey(sender sdk.AccAddress, chain Chain, txHash *chainhash.Hash, amount btcutil.Amount) sdk.Msg {
	return MsgRawTxForMasterKey{
		Sender: sender,
		TxHash: txHash,
		Amount: amount,
		Chain:  chain,
	}
}

func (msg MsgRawTxForMasterKey) Route() string {
	return RouterKey
}

func (msg MsgRawTxForMasterKey) Type() string {
	return "RawTxForMasterKey"
}

func (msg MsgRawTxForMasterKey) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.TxHash == nil {
		return fmt.Errorf("missing transaction ID")
	}
	if msg.Amount <= 0 {
		return fmt.Errorf("transaction amount must be greater than zero")
	}
	return msg.Chain.Validate()
}

func (msg MsgRawTxForMasterKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgRawTxForMasterKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
