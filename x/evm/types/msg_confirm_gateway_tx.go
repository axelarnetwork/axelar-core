package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewConfirmGatewayTxRequest creates a message of type ConfirmGatewayTxRequest
func NewConfirmGatewayTxRequest(sender sdk.AccAddress, chain string, txID Hash) *ConfirmGatewayTxRequest {
	return &ConfirmGatewayTxRequest{
		Sender: sender,
		Chain:  nexus.ChainName(utils.NormalizeString(chain)),
		TxID:   txID,
	}
}

// Route implements sdk.Msg
func (m ConfirmGatewayTxRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmGatewayTxRequest) Type() string {
	return "ConfirmGatewayTx"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmGatewayTxRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if m.TxID.IsZero() {
		return fmt.Errorf("invalid tx id")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmGatewayTxRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m ConfirmGatewayTxRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
