package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteConfirmTransferOwnershipRequest creates a message of type VoteConfirmTransferOwnershipRequest
func NewVoteConfirmTransferOwnershipRequest(
	sender sdk.AccAddress,
	chain string,
	poll vote.PollKey,
	txID common.Hash,
	newOwnerAddr Address,
	confirmed bool) *VoteConfirmTransferOwnershipRequest {
	return &VoteConfirmTransferOwnershipRequest{
		Sender:          sender,
		Chain:           chain,
		Poll:            poll,
		TxID:            Hash(txID),
		NewOwnerAddress: newOwnerAddr,
		Confirmed:       confirmed,
	}
}

// Route returns the route for this message
func (m VoteConfirmTransferOwnershipRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m VoteConfirmTransferOwnershipRequest) Type() string {
	return "VoteConfirmTransferOwnership"
}

// ValidateBasic executes a stateless message validation
func (m VoteConfirmTransferOwnershipRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	return m.Poll.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (m VoteConfirmTransferOwnershipRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m VoteConfirmTransferOwnershipRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
