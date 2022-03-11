package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewAddressValidator returns the callback for validating hex-encoded EVM addresses
func NewAddressValidator() nexus.AddressValidator {
	return func(ctx sdk.Context, address nexus.CrossChainAddress) error {
		if !common.IsHexAddress(address.Address) {
			return fmt.Errorf("not an hex address")
		}

		return nil
	}
}
