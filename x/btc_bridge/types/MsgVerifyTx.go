package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Ensure MsgTrackAddress implements sdk.Msg interface
var _ sdk.Msg = &MsgVerifyTx{}

type MsgVerifyTx struct {
	Sender sdk.AccAddress
	UTXO   UTXO
}

func NewMsgVerifyTx(sender sdk.AccAddress, chain string, txHash *chainhash.Hash, voutIdx uint32, destination string, amount btcutil.Amount) MsgVerifyTx {
	return MsgVerifyTx{
		Sender: sender,
		UTXO: UTXO{
			Chain:   chain,
			Hash:    txHash,
			VoutIdx: voutIdx,
			Address: destination,
			Amount:  amount,
		},
	}
}

func (msg MsgVerifyTx) Route() string {
	return RouterKey
}

func (msg MsgVerifyTx) Type() string {
	return "VerifyTx"
}

func (msg MsgVerifyTx) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.UTXO.IsInvalid() {
		return fmt.Errorf("invalid utxo")
	}

	return nil
}

func (msg MsgVerifyTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgVerifyTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
