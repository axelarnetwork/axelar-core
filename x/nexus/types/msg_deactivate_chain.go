package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
)

// NewDeactivateChainRequest creates a message of type DeactivateChainRequest
func NewDeactivateChainRequest(sender sdk.AccAddress, chains ...string) *DeactivateChainRequest {
	return &DeactivateChainRequest{
		Sender: sender.String(),
		Chains: slices.Map(chains, func(c string) exported.ChainName {
			return exported.ChainName(utils.NormalizeString(c))
		}),
	}
}

// Route implements sdk.Msg
func (m DeactivateChainRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m DeactivateChainRequest) Type() string {
	return "DeactivateChain"
}

// ValidateBasic implements sdk.Msg
func (m DeactivateChainRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if len(m.Chains) == 0 {
		return fmt.Errorf("missing chains")
	}

	for _, chain := range m.Chains {
		if err := chain.Validate(); err != nil {
			return errorsmod.Wrap(err, "invalid chain")
		}
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m DeactivateChainRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
