package types

import (
	"fmt"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewAddChainRequest is the constructor for NewAddChainRequest
func NewAddChainRequest(sender sdk.AccAddress, name, nativeAsset string) *AddChainRequest {
	return &AddChainRequest{
		Sender:      sender,
		Name:        name,
		NativeAsset: nativeAsset,
	}
}

// Route returns the route for this message
func (m AddChainRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m AddChainRequest) Type() string {
	return "VoteConfirmToken"
}

// ValidateBasic executes a stateless message validation
func (m AddChainRequest) ValidateBasic() error {
	if m.Sender == nil || len(m.Sender) != sdk.AddrLen {
		return fmt.Errorf("missing sender")
	}

	chain := nexus.Chain{
		Name:                  m.Name,
		NativeAsset:           m.NativeAsset,
		SupportsForeignAssets: true,
	}

	if err := chain.Validate(); err != nil {
		return fmt.Errorf("invalid chain spec: %v", err)
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
