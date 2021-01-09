package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

type MsgVerifyTx struct {
	Sender sdk.AccAddress
	Tx     Tx
	TxType TXType
}

func NewMsgVerifyMintTx(sender sdk.AccAddress, network Network, hash common.Hash, address common.Address, amount sdk.Int) sdk.Msg {
	return MsgVerifyTx{
		Sender: sender,
		Tx: Tx{
			Network:     network,
			Hash:        hash,
			Destination: address,
			Amount:      amount,
		},
		TxType: TypeERC20mint,
	}
}

func NewMsgVerifyDeployTx(sender sdk.AccAddress, network Network, hash common.Hash, contractID string) sdk.Msg {
	return MsgVerifyTx{
		Sender: sender,
		Tx: Tx{
			Network:    network,
			Hash:       hash,
			ContractID: contractID,
		},
		TxType: TypeSCDeploy,
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
	if err := msg.Tx.Network.Validate(); err != nil {
		return err
	}
	switch msg.TxType {
	case TypeSCDeploy:
		if msg.Tx.ContractID == "" {
			return fmt.Errorf("missing byte code")
		}
	case TypeERC20mint:
		if msg.Tx.Amount.IsNegative() {
			return fmt.Errorf("amount must be greater than 0")
		}
	default:
		return fmt.Errorf(fmt.Sprintf("wrong tx type"))
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
