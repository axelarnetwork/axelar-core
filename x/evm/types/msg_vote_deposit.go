package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/axelarnetwork/axelar-core/utils"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteConfirmDepositRequest creates a message of type ConfirmTokenRequest
func NewVoteConfirmDepositRequest(
	sender sdk.AccAddress,
	chain string,
	key vote.PollKey,
	txID common.Hash,
	burnAddr Address,
	confirmed bool) *VoteConfirmDepositRequest {
	return &VoteConfirmDepositRequest{
		Sender:      sender,
		Chain:       utils.NormalizeString(chain),
		PollKey:     key,
		TxID:        Hash(txID),
		BurnAddress: burnAddr,
		Confirmed:   confirmed,
	}
}

// Route returns the route for this message
func (m VoteConfirmDepositRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m VoteConfirmDepositRequest) Type() string {
	return "VoteConfirmDeposit"
}

// ValidateBasic executes a stateless message validation
func (m VoteConfirmDepositRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if err := utils.ValidateString(m.Chain); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	return m.PollKey.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (m VoteConfirmDepositRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m VoteConfirmDepositRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
