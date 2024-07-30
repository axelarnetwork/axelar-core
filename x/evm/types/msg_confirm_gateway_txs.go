package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
)

const TxLimit = 10

// NewConfirmGatewayTxsRequest creates a message of type ConfirmGatewayTxsRequest
func NewConfirmGatewayTxsRequest(sender sdk.AccAddress, chain nexus.ChainName, txIDs []Hash) *ConfirmGatewayTxsRequest {
	return &ConfirmGatewayTxsRequest{
		Sender: sender,
		Chain:  chain,
		TxIDs:  txIDs,
	}
}

// Route implements sdk.Msg
func (m ConfirmGatewayTxsRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmGatewayTxsRequest) Type() string {
	return "ConfirmGatewayTxs"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmGatewayTxsRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if len(m.TxIDs) == 0 {
		return errors.New("tx ids cannot be empty")
	}

	if len(m.TxIDs) > TxLimit {
		return errors.New("tx ids limit exceeded")
	}

	if slices.Any(m.TxIDs, Hash.IsZero) {
		return errors.New("invalid tx id")
	}

	if slices.HasDuplicates(m.TxIDs) {
		return errors.New("duplicate tx ids")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmGatewayTxsRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m ConfirmGatewayTxsRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
