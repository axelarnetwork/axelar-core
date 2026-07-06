package app

import (
	"os"
	"testing"

	"cosmossdk.io/log"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	"github.com/CosmWasm/wasmd/x/wasm"
	dbm "github.com/cosmos/cosmos-db"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/utils/funcs"
)

// TestV15Upgrade guards the v1.5 upgrade entry. The deleted-stores list is
// applied verbatim at the upgrade height, so a wrong or extra store here would
// brick nodes during the upgrade. The handler registration is verified against
// a constructed app, since setUpgradeBehaviour runs at construction.
func TestV15Upgrade(t *testing.T) {
	var v15 *chainUpgrade
	for i := range chainUpgrades {
		if chainUpgrades[i].name == "v1.5" {
			v15 = &chainUpgrades[i]
		}
	}
	require.NotNil(t, v15, "v1.5 must be in the upgrade registry")

	assert.Equal(t, []string{"capability", "crisis"}, v15.storeUpgrades.Deleted)
	assert.Empty(t, v15.storeUpgrades.Added)
	assert.Empty(t, v15.storeUpgrades.Renamed)

	WasmEnabled, IBCWasmHooksEnabled = "true", "false"
	t.Cleanup(func() { funcs.MustNoErr(os.RemoveAll("wasm")) })

	axelarApp := NewAxelarApp(
		log.NewTestLogger(t),
		dbm.NewMemDB(),
		nil,
		true,
		MakeEncodingConfig(),
		simtestutil.EmptyAppOptions{},
		[]wasm.Option{},
	)

	assert.True(t, GetKeeper[upgradekeeper.Keeper](axelarApp.Keepers).HasHandler("v1.5"))
}
