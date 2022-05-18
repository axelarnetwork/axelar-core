package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewSetGatewayRequest creates a message of type SetGatewayRequest
func NewSetGatewayRequest(sender sdk.AccAddress, chain string, address Address) *SetGatewayRequest {
	return &SetGatewayRequest{
		Sender:  sender,
		Chain:   nexus.ChainName(utils.NormalizeString(chain)),
		Address: address,
	}
}

// Route implements sdk.Msg
func (m SetGatewayRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m SetGatewayRequest) Type() string {
	return "SetGateway"
}

// ValidateBasic implements sdk.Msg
func (m SetGatewayRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain name")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m SetGatewayRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m SetGatewayRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
