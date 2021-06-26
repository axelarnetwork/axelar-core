package exported

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate performs a stateless check to ensure the Chain object has been initialized correctly
func (m Chain) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("missing chain name")
	}
	if m.NativeAsset == "" {
		return fmt.Errorf("missing native asset name")
	}
	return nil
}

// MergeTransfersBy merges cross chain transfers grouped by the given function
func MergeTransfersBy(transfers []CrossChainTransfer, groupFn func(transfer CrossChainTransfer) string) []CrossChainTransfer {
	results := []CrossChainTransfer{}
	transferAmountByAddressAndAsset := map[string]sdk.Int{}

	for _, transfer := range transfers {
		id := groupFn(transfer)

		if _, ok := transferAmountByAddressAndAsset[id]; !ok {
			transferAmountByAddressAndAsset[id] = sdk.ZeroInt()
		}

		transferAmountByAddressAndAsset[id] = transferAmountByAddressAndAsset[id].Add(transfer.Asset.Amount)
	}

	seen := map[string]bool{}

	for _, transfer := range transfers {
		id := groupFn(transfer)

		if seen[id] {
			continue
		}

		transfer.Asset.Amount = transferAmountByAddressAndAsset[id]
		results = append(results, transfer)
		seen[id] = true
	}

	return results
}
