package types

import (
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteConfirmOutpointRequest - MsgVoteConfirmOutpoint constructor
func NewVoteConfirmOutpointRequest(sender sdk.AccAddress, key exported.PollKey, outPoint wire.OutPoint, confirmed bool) *VoteConfirmOutpointRequest {
	return &VoteConfirmOutpointRequest{
		Sender:    sender,
		PollKey:   key,
		OutPoint:  outPoint.String(),
		Confirmed: confirmed,
	}
}

// Route returns the route for this message
func (m VoteConfirmOutpointRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m VoteConfirmOutpointRequest) Type() string {
	return "VoteConfirmDeposit"
}

// ValidateBasic executes a stateless message validation
func (m VoteConfirmOutpointRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if _, err := OutPointFromStr(m.OutPoint); err != nil {
		return sdkerrors.Wrap(err, "outpoint malformed")
	}
	return m.PollKey.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (m VoteConfirmOutpointRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m VoteConfirmOutpointRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
