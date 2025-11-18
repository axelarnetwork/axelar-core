package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewConfirmGatewayTxRequest creates a message of type ConfirmGatewayTxRequest
func NewConfirmGatewayTxRequest(sender sdk.AccAddress, chain string, txID Hash) *ConfirmGatewayTxRequest {
	return &ConfirmGatewayTxRequest{
		Sender: sender.String(),
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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
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
