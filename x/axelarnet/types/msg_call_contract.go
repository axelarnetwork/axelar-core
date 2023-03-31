package types

import (
	fmt "fmt"

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

	if len(m.ContractAddress) == 0 {
		return fmt.Errorf("contract address empty")
	}

	if m.Fee != nil {

		if err := sdk.VerifyAddressFormat(m.Fee.Recipient); err != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "fee recipient").Error())
		}

		if !m.Fee.Amount.IsValid() || !m.Fee.Amount.IsPositive() {
			return fmt.Errorf("invalid fee amount")
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
