package types

import (
	"fmt"

	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewVoteConfirmGatewayDeploymentRequest creates a message of type VoteConfirmGatewayDeploymentRequest
func NewVoteConfirmGatewayDeploymentRequest(sender sdk.AccAddress, chain string, key vote.PollKey, confirmed bool) *VoteConfirmGatewayDeploymentRequest {
	return &VoteConfirmGatewayDeploymentRequest{
		Sender:    sender,
		Chain:     chain,
		PollKey:   key,
		Confirmed: confirmed,
	}
}

// Route implements sdk.Msg
func (m VoteConfirmGatewayDeploymentRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m VoteConfirmGatewayDeploymentRequest) Type() string {
	return "ConfirmGatewayDeployment"
}

// ValidateBasic implements sdk.Msg
func (m VoteConfirmGatewayDeploymentRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	if err := m.PollKey.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m VoteConfirmGatewayDeploymentRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements sdk.Msg
func (m VoteConfirmGatewayDeploymentRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
