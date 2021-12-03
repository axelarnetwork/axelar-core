package keeper

import (
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewNexusHandler returns the handler for validating cosmos SDK addresses
func NewNexusHandler() nexus.Handler {
	return &handler{addrPrefix: types.ModuleName}
}

type handler struct {
	addrPrefix string
}

func (h *handler) Validate(ctx sdk.Context, address nexus.CrossChainAddress) error {
	if _, err := sdk.GetFromBech32(address.Address, h.addrPrefix); err != nil {
		return err
	}

	return nil
}

func (h *handler) With(properties ...nexus.HandlerProperty) nexus.Handler {
	var newHandler nexus.Handler = h
	for _, property := range properties {
		newHandler = property(newHandler)
	}
	return newHandler
}

// AddrPrefix sets the address prefix to be assumed for validation
func AddrPrefix(prefix string) nexus.HandlerProperty {
	return func(h nexus.Handler) nexus.Handler {
		axelarnetHandler := h.(*handler)
		axelarnetHandler.addrPrefix = prefix
		return axelarnetHandler
	}
}
