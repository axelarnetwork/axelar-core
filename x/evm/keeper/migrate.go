package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/utils/slices"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.24 to v0.25
// The migration includes:
// - set TransferLimit parameter
func GetMigrationHandler(k BaseKeeper, n types.Nexus, s types.Signer, m types.MultisigKeeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// set TransferLimit param
		for _, chain := range slices.Filter(n.GetChains(ctx), types.IsEVMChain) {
			ck := k.ForChain(chain.Name).(chainKeeper)
			if err := addTransferLimitParam(ctx, ck); err != nil {
				return err
			}
		}

		return nil
	}
}

func addTransferLimitParam(ctx sdk.Context, ck chainKeeper) error {
	subspace, ok := ck.getSubspace(ctx)
	if !ok {
		return fmt.Errorf("param subspace for chain %s should exist", ck.GetName())
	}

	subspace.Set(ctx, types.KeyTransferLimit, types.DefaultParams()[0].TransferLimit)

	return nil
}
