package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/axelarnetwork/axelar-core/utils"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteConfirmGatewayTxRequest creates a message of type VoteConfirmGatewayTxRequest
func NewVoteConfirmGatewayTxRequest(sender sdk.AccAddress, pollKey vote.PollKey, vote VoteConfirmGatewayTxRequest_Vote) *VoteConfirmGatewayTxRequest {
	return &VoteConfirmGatewayTxRequest{
		Sender:  sender,
		PollKey: pollKey,
		Vote:    vote,
	}
}

// Route implements sdk.Msg
func (m VoteConfirmGatewayTxRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m VoteConfirmGatewayTxRequest) Type() string {
	return "VoteConfirmGatewayTx"
}

// GetChain returns chain in poll key
func (m VoteConfirmGatewayTxRequest) GetChain() string {
	pollKeyIdItems := strings.Split(m.PollKey.ID, "_")
	if len(pollKeyIdItems) != 2 {
		return ""
	}

	return pollKeyIdItems[0]
}

// GetTxID returns the tx id in poll key
func (m VoteConfirmGatewayTxRequest) GetTxID() string {
	pollKeyIdItems := strings.Split(m.PollKey.ID, "_")
	if len(pollKeyIdItems) != 2 {
		return ""
	}

	return pollKeyIdItems[1]
}

// ValidateBasic implements sdk.Msg
func (m VoteConfirmGatewayTxRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.PollKey.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid poll key")
	}

	if err := utils.ValidateString(m.GetChain()); err != nil {
		return sdkerrors.Wrap(err, "invalid chain in poll key")
	}

	txID, err := hexutil.Decode(m.GetTxID())
	if err != nil {
		return sdkerrors.Wrap(err, "invalid tx id in poll key")
	}

	if len(txID) != common.HashLength {
		return fmt.Errorf("invalid tx id in poll key")
	}

	for i, event := range m.Vote.Events {
		if err := event.Info.Validate(); err != nil {
			return sdkerrors.Wrap(err, "invalid event info")
		}

		if event.Info.Chain != m.GetChain() {
			return fmt.Errorf("invalid source chain in event ContractCallWithToken")
		}

		if event.Info.TxId.Hex() != m.GetTxID() {
			return fmt.Errorf("invalid tx id in event ContractCallWithToken")
		}

		if event.Info.Index != uint64(i) {
			return fmt.Errorf("invalid index in event ContractCallWithToken")
		}

		switch event := event.GetEvent().(type) {
		case *VoteConfirmGatewayTxRequest_Vote_Event_ContractCallWithToken:
			if event.ContractCallWithToken == nil {
				return sdkerrors.Wrap(err, "missing event ContractCallWithToken")
			}

			if err := event.ContractCallWithToken.Validate(); err != nil {
				return sdkerrors.Wrap(err, "invalid event ContractCallWithToken")
			}
		default:
			return fmt.Errorf("unknown type of event")
		}
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m VoteConfirmGatewayTxRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m VoteConfirmGatewayTxRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
