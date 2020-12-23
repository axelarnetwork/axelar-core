package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgVerifyTx struct {
	Sender sdk.AccAddress
	UTXO   UTXO
	Mode   Mode
	Chain  Chain
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
		Mode: ModeSpecificAddress,
	}
}

func NewMsgVerifyTxForMasterKey(sender sdk.AccAddress, txHash *chainhash.Hash, voutIdx uint32, amount btcutil.Amount, mode Mode, chain Chain) sdk.Msg {
	return MsgVerifyTx{
		Sender: sender,
		UTXO: UTXO{
			Hash:    txHash,
			VoutIdx: voutIdx,
			Amount:  amount,
		},
		Mode:  mode,
		Chain: chain,
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

	switch msg.Mode {
	case ModeSpecificAddress:
		return msg.UTXO.Validate()
	case ModeCurrentMasterKey, ModeNextMasterKey:
		if msg.UTXO.Hash == nil {
			return fmt.Errorf("missing hash")
		}
		if msg.UTXO.Amount <= 0 {
			return fmt.Errorf("amount must be greater than 0")
		}
		if msg.UTXO.Address.Validate() == nil {
			return fmt.Errorf("destination should not be set when using master key flags")
		}
		if err := msg.Chain.Validate(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("chosen mode not recognized")
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
