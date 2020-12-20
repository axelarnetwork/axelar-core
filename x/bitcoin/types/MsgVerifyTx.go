package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgVerifyTx struct {
	Sender       sdk.AccAddress
	UTXO         UTXO
	UseMasterKey bool
	Chain        Chain
}

func NewMsgVerifyTx(sender sdk.AccAddress, txHash *chainhash.Hash, voutIdx uint32, destination BtcAddress, amount btcutil.Amount) sdk.Msg {
	return MsgVerifyTx{
		Sender: sender,
		UTXO: UTXO{
			Hash:    txHash,
			VoutIdx: voutIdx,
			Address: destination,
			Amount:  amount,
		},
		UseMasterKey: false,
	}
}

func NewMsgVerifyTxForMasterKey(sender sdk.AccAddress, txHash *chainhash.Hash, voutIdx uint32, amount btcutil.Amount, chain Chain) sdk.Msg {
	return MsgVerifyTx{
		Sender: sender,
		UTXO: UTXO{
			Hash:    txHash,
			VoutIdx: voutIdx,
			Amount:  amount,
		},
		UseMasterKey: true,
		Chain:        chain,
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
	if msg.UseMasterKey {
		if msg.UTXO.Hash == nil {
			return fmt.Errorf("missing hash")
		}
		if msg.UTXO.Amount <= 0 {
			return fmt.Errorf("amount must be greater than 0")
		}
		if err := msg.Chain.Validate(); err != nil {
			return err
		}
	} else {
		if err := msg.UTXO.Validate(); err != nil {
			return err
		}
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
