package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewCallContractRequest is the constructor for NewCallContractRequest
func NewCallContractRequest(sender sdk.AccAddress, chain string, contractAddress string, payload []byte, fee *Fee) *CallContractRequest {
	return &CallContractRequest{
		Sender:          sender,
		Chain:           nexus.ChainName(utils.NormalizeString(chain)),
		ContractAddress: contractAddress,
		Payload:         payload,
		Fee:             fee,
	}
}

// Route returns the route for this message
func (m CallContractRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m CallContractRequest) Type() string {
	return "CallContract"
}

// ValidateBasic executes a stateless message validation
func (m CallContractRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := utils.ValidateString(m.ContractAddress); err != nil {
		return err
	}

	if m.Fee != nil {
		if err := m.Fee.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "fee")
		}
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m CallContractRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m CallContractRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
