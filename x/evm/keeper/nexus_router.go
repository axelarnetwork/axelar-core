package keeper

import (
	"fmt"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

// NewNexusHandler returns the handler for validating hex-encoded EVM addresses
func NewNexusHandler() nexus.Handler {
	return func(ctx sdk.Context, address nexus.CrossChainAddress) error {
		if !common.IsHexAddress(address.Address) {
			return fmt.Errorf("not an hex address")
		}

		return nil
	}
}
