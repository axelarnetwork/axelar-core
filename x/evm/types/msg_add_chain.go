package types

import (
	"fmt"
	"strings"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewAddChainRequest is the constructor for NewAddChainRequest
func NewAddChainRequest(sender sdk.AccAddress, name string, keyType tss.KeyType, params Params) *AddChainRequest {
	return &AddChainRequest{
		Sender:      sender,
		Name:        utils.NormalizeString(name),
		KeyType:     keyType,
		Params:      params,
	}
}

// Route returns the route for this message
func (m AddChainRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m AddChainRequest) Type() string {
	return "AddChain"
}

// ValidateBasic executes a stateless message validation
func (m AddChainRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	chain := nexus.Chain{
		Name:                  m.Name,
		SupportsForeignAssets: true,
		KeyType:               m.KeyType,
		Module:                ModuleName,
	}

	if err := chain.Validate(); err != nil {
		return fmt.Errorf("invalid chain spec: %v", err)
	}

	if err := m.Params.Validate(); err != nil {
		return fmt.Errorf("invalid EVM param: %v", err)
	}

	if !strings.EqualFold(m.Name, m.Params.Chain) {
		return fmt.Errorf("chain mismatch: chain name is %s, parameters chain is %s", m.Name, m.Params.Chain)
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m AddChainRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m AddChainRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
