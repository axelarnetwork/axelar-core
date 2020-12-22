package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgRawTx struct {
	Sender      sdk.AccAddress
	TxHash      *chainhash.Hash
	Amount      btcutil.Amount
	Destination BtcAddress
	Chain       Chain
	Mode        int
}

func NewMsgRawTx(sender sdk.AccAddress, txHash *chainhash.Hash, amount btcutil.Amount, destination BtcAddress) sdk.Msg {
	return MsgRawTx{
		Sender:      sender,
		TxHash:      txHash,
		Amount:      amount,
		Destination: destination,
		Mode:        ModeSpecificAddress,
	}
}

func NewMsgRawTxForNextMasterKey(sender sdk.AccAddress, chain Chain, txHash *chainhash.Hash, amount btcutil.Amount) sdk.Msg {
	return MsgRawTx{
		Sender: sender,
		TxHash: txHash,
		Amount: amount,
		Chain:  chain,
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
	if msg.Amount <= 0 {
		return fmt.Errorf("transaction amount must be greater than zero")
	}

	switch msg.Mode {
	case ModeSpecificAddress:
		if err := msg.Destination.Validate(); err != nil {
			return sdkerrors.Wrap(err, "invalid destination")
		}
	case ModeNextMasterKey:
		return msg.Chain.Validate()
	default:
		return fmt.Errorf("chosen mode not recognized")
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
