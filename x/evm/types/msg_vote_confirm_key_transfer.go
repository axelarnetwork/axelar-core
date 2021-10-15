package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteConfirmTransferKeyRequest creates a message of type VoteConfirmTransferKeyRequest
func NewVoteConfirmTransferKeyRequest(
	sender sdk.AccAddress,
	chain string,
	key vote.PollKey,
	txID common.Hash,
	transferType KeyTransferType,
	newAddr Address,
	confirmed bool) *VoteConfirmKeyTransferRequest {
	return &VoteConfirmKeyTransferRequest{
		Sender:       sender,
		Chain:        chain,
		PollKey:      key,
		TxID:         Hash(txID),
		TransferType: transferType,
		NewAddress:   newAddr,
		Confirmed:    confirmed,
	}
}

// Route returns the route for this message
func (m VoteConfirmKeyTransferRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m VoteConfirmKeyTransferRequest) Type() string {
	return "VoteConfirmTransferKey"
}

// ValidateBasic executes a stateless message validation
func (m VoteConfirmKeyTransferRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	if err := m.PollKey.Validate(); err != nil {
		return err
	}

	if err := m.TransferType.Validate(); err != nil {
		return err
	}

	return m.PollKey.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (m VoteConfirmKeyTransferRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m VoteConfirmKeyTransferRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
