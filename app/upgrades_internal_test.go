package app

import (
	"os"
	"testing"

	"cosmossdk.io/log"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/module"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	feepolicyKeeper "github.com/axelarnetwork/axelar-core/x/feepolicy/keeper"
	feepolicyTypes "github.com/axelarnetwork/axelar-core/x/feepolicy/types"
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
	assert.Equal(t, []string{"authz"}, v15.storeUpgrades.Added)
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

	assert.NotNil(t, GetKeeper[authzkeeper.Keeper](axelarApp.Keepers))
	assert.Contains(t, axelarApp.Keys, "authz")
	assert.Contains(t, axelarApp.mm.GetVersionMap(), "authz")
}

// TestV15UpgradeSeedsFeeDenomsFromBondDenom checks that the v1.5 handler leaves
// the new x/feepolicy allowlist holding the chain's own bond denom instead of the
// uaxl hardcoded in the module's default genesis.
func TestV15UpgradeSeedsFeeDenomsFromBondDenom(t *testing.T) {
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

	ctx := axelarApp.NewUncachedContext(false, cmtproto.Header{})

	stakingK := GetKeeper[stakingkeeper.Keeper](axelarApp.Keepers)
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = "utest"
	require.NoError(t, stakingK.SetParams(ctx, stakingParams))

	// pretend every module but the brand new x/feepolicy is already at its
	// current version, so the upgrade only runs the feepolicy initialization
	fromVM := axelarApp.mm.GetVersionMap()
	delete(fromVM, feepolicyTypes.ModuleName)

	// only x/feepolicy's InitGenesis runs, so the configurator does not need any
	// registered module migrations
	configurator := module.NewConfigurator(axelarApp.appCodec, axelarApp.MsgServiceRouter(), axelarApp.GRPCQueryRouter())

	handler := v15UpgradeHandler(axelarApp, configurator, axelarApp.Keepers)
	toVM, err := handler(ctx, upgradetypes.Plan{Name: "v1.5"}, fromVM)
	require.NoError(t, err)
	assert.Equal(t, axelarApp.mm.GetVersionMap(), toVM)

	assert.Equal(t, []string{"utest"}, GetKeeper[feepolicyKeeper.Keeper](axelarApp.Keepers).GetAllowedFeeDenoms(ctx))
}
