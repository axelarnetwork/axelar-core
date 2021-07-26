package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteConfirmTokenRequest creates a message of type ConfirmTokenRequest
func NewVoteConfirmTokenRequest(
	sender sdk.AccAddress,
	chain, asset string,
	key vote.PollKey,
	txID common.Hash,
	confirmed bool) *VoteConfirmTokenRequest {
	return &VoteConfirmTokenRequest{
		Sender:    sender,
		Chain:     chain,
		PollKey:   key,
		TxID:      Hash(txID),
		Asset:     asset,
		Confirmed: confirmed,
	}
}

// Route returns the route for this message
func (m VoteConfirmTokenRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m VoteConfirmTokenRequest) Type() string {
	return "VoteConfirmToken"
}

// ValidateBasic executes a stateless message validation
func (m VoteConfirmTokenRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}
	if m.Asset == "" {
		return fmt.Errorf("asset missing")
	}

	return m.PollKey.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (m VoteConfirmTokenRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m VoteConfirmTokenRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
