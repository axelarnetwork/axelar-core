package types

import (
	"github.com/axelarnetwork/axelar-core/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// NewConfirmGatewayDeploymentRequest creates a message of type ConfirmGatewayDeploymentRequest
func NewConfirmGatewayDeploymentRequest(sender sdk.AccAddress, chain string, txID common.Hash, address common.Address) *ConfirmGatewayDeploymentRequest {
	return &ConfirmGatewayDeploymentRequest{
		Sender:  sender,
		Chain:   utils.NormalizeString(chain),
		TxID:    Hash(txID),
		Address: Address(address),
	}
}

// Route implements sdk.Msg
func (m ConfirmGatewayDeploymentRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmGatewayDeploymentRequest) Type() string {
	return "ConfirmGatewayDeployment"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmGatewayDeploymentRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := utils.ValidateString(m.Chain); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmGatewayDeploymentRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements sdk.Msg
func (m ConfirmGatewayDeploymentRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
