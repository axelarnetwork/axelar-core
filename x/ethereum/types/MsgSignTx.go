package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

// NewMsgSignTx - constructor
func NewMsgSignTx(sender sdk.AccAddress, jsonTx []byte) *MsgSignTx {
	return &MsgSignTx{
		Sender: sender,
		Tx:     jsonTx,
	}
}

// Route returns the route of the message
func (m MsgSignTx) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m MsgSignTx) Type() string {
	return "SignTx"
}

// ValidateBasic executes a stateless message validation
func (m MsgSignTx) ValidateBasic() error {
	if m.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if m.Tx == nil {
		return fmt.Errorf("missing tx")
	}
	tx := ethTypes.Transaction{}
	if err := tx.UnmarshalJSON(m.Tx); err != nil {
		return err
	}
	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m MsgSignTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m MsgSignTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// UnmarshaledTx returns the unmarshaled ethereum transaction contained in this message
func (m MsgSignTx) UnmarshaledTx() *ethTypes.Transaction {
	tx := &ethTypes.Transaction{}
	err := tx.UnmarshalJSON(m.Tx)
	if err != nil {
		panic(err)
	}
	return tx
}
