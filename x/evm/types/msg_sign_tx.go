package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmTypes "github.com/ethereum/go-ethereum/core/types"
)

// NewSignTxRequest - constructor
func NewSignTxRequest(sender sdk.AccAddress, chain string, jsonTx []byte) *SignTxRequest {
	return &SignTxRequest{
		Sender: sender,
		Chain:  chain,
		Tx:     jsonTx,
	}
}

// Route returns the route of the message
func (m SignTxRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m SignTxRequest) Type() string {
	return "SignTx"
}

// ValidateBasic executes a stateless message validation
func (m SignTxRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}
	if m.Tx == nil {
		return fmt.Errorf("missing tx")
	}
	tx := evmTypes.Transaction{}
	if err := tx.UnmarshalJSON(m.Tx); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m SignTxRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m SignTxRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// UnmarshaledTx returns the unmarshaled evm transaction contained in this message
func (m SignTxRequest) UnmarshaledTx() *evmTypes.Transaction {
	tx := &evmTypes.Transaction{}
	err := tx.UnmarshalJSON(m.Tx)
	if err != nil {
		panic(err)
	}
	return tx
}
