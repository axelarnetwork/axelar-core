package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// TestGetChainMaintainers verifies that GetChainMaintainers, which recovers maintainer
// addresses directly from the store keys (without decoding the maintainer states),
// returns exactly the same addresses in the same order as deriving them from the fully
// decoded states via GetChainMaintainerStates.
func TestGetChainMaintainers(t *testing.T) {
	ctx, k := setup(t)

	chain := exported.Chain{Name: exported.ChainName("ethereum")}

	// Addresses deliberately include the delimiter byte ('_' == 0x5f) and the prefix's
	// integer-encoded bytes to make sure the key parser strips a fixed prefix length
	// rather than splitting on the delimiter.
	addresses := []sdk.ValAddress{
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13},
		{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f},
		{0xff, 0xfe, 0xfd, 0xfc, 0xfb, 0xfa, 0xf9, 0xf8, 0xf7, 0xf6, 0xf5, 0xf4, 0xf3, 0xf2, 0xf1, 0xf0, 0x5f, 0x00, 0x5f, 0x42},
	}

	for _, addr := range addresses {
		funcs.MustNoErr(k.AddChainMaintainer(ctx, chain, addr))
	}

	// Reference: addresses derived from the fully decoded states (the pre-optimization path).
	want := slices.Map(k.GetChainMaintainerStates(ctx, chain), exported.MaintainerState.GetAddress)

	got := k.GetChainMaintainers(ctx, chain)

	// Same store iteration order, so equality (not just set membership) must hold.
	assert.Equal(t, want, got)
	assert.ElementsMatch(t, addresses, got)
}
