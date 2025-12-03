package app_test

import (
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/app"
)

func TestDirectStoreWriteMatchesGetUpgradePlan(t *testing.T) {
	// This test verifies that writing an upgrade plan directly to the store
	// (bypassing ScheduleUpgrade) produces data that GetUpgradePlan can read.
	// This is important because scheduleUpgradeFromDisk uses direct store writes
	// to avoid issues with legacy "done" key formats.

	version.Version = "v1.3.0"
	app.WasmEnabled = "false"
	t.Cleanup(cleanup)

	axelarApp := app.NewAxelarApp(
		log.NewTestLogger(t),
		dbm.NewMemDB(),
		nil,
		true,
		app.MakeEncodingConfig(),
		simtestutil.EmptyAppOptions{},
		[]wasm.Option{},
	)

	// Create a test plan
	expectedPlan := upgradetypes.Plan{
		Name:   "v1.4",
		Height: 12345,
		Info:   "test upgrade",
	}

	// Get the context - must set both BlockHeader and HeaderInfo for SDK v0.50+
	ctx := axelarApp.NewUncachedContext(false, cmtproto.Header{}).WithHeaderInfo(header.Info{
		Height: 1,
		Time:   time.Now(),
	})

	// Write directly to store (same method as scheduleUpgradeFromDisk)
	kvStore := ctx.KVStore(axelarApp.Keys[upgradetypes.StoreKey])
	bz, err := axelarApp.AppCodec().Marshal(&expectedPlan)
	require.NoError(t, err)
	kvStore.Set(upgradetypes.PlanKey(), bz)

	// Read back using the keeper's GetUpgradePlan
	upgradeKeeper := app.GetKeeper[upgradekeeper.Keeper](axelarApp.Keepers)
	actualPlan, err := upgradeKeeper.GetUpgradePlan(ctx)
	require.NoError(t, err)

	// Verify they match
	assert.Equal(t, expectedPlan.Name, actualPlan.Name)
	assert.Equal(t, expectedPlan.Height, actualPlan.Height)
	assert.Equal(t, expectedPlan.Info, actualPlan.Info)
}

func TestLegacyDoneKeysCauseGetLastCompletedUpgradePanic(t *testing.T) {
	// This test verifies that legacy "done" keys from pre-v0.50 SDK cause a panic
	// when GetLastCompletedUpgrade is called. This proves the problem exists and
	// justifies why scheduleUpgradeFromDisk must delete legacy keys.
	//
	// Pre-v0.50 SDK format: DoneByte + upgradeName (variable length, e.g., 5 bytes for "v1.0")
	// v0.50 SDK format: DoneByte + uint64(height) + upgradeName (10+ bytes minimum)
	//
	// The upgrade PreBlocker calls GetLastCompletedUpgrade for downgrade verification,
	// so this panic would occur during normal node startup after an SDK upgrade.

	version.Version = "v1.3.0"
	app.WasmEnabled = "false"
	t.Cleanup(cleanup)

	axelarApp := app.NewAxelarApp(
		log.NewTestLogger(t),
		dbm.NewMemDB(),
		nil,
		true,
		app.MakeEncodingConfig(),
		simtestutil.EmptyAppOptions{},
		[]wasm.Option{},
	)

	// Must set HeaderInfo for SDK v0.50+ - PreBlocker uses HeaderInfo().Height
	ctx := axelarApp.NewUncachedContext(false, cmtproto.Header{Height: 100}).WithHeaderInfo(header.Info{
		Height: 100,
		Time:   time.Now(),
	})
	kvStore := ctx.KVStore(axelarApp.Keys[upgradetypes.StoreKey])

	// Write a legacy "done" key (old format without height prefix)
	legacyDoneKey := append([]byte{upgradetypes.DoneByte}, []byte("v1.2")...)
	kvStore.Set(legacyDoneKey, []byte{1})

	upgradeKeeper := app.GetKeeper[upgradekeeper.Keeper](axelarApp.Keepers)

	// GetLastCompletedUpgrade should panic because it tries to parse the legacy key
	// expecting at least 10 bytes but gets only 5 bytes
	assert.Panics(t, func() {
		_, _, _ = upgradeKeeper.GetLastCompletedUpgrade(ctx)
	}, "GetLastCompletedUpgrade should panic on legacy done keys")
}

func TestLegacyDoneKeysCausePreBlockerPanic(t *testing.T) {
	// This test verifies that PreBlocker panics when legacy "done" keys exist.
	// This is the actual code path that fails during a real upgrade.
	//
	// PreBlocker flow:
	// 1. Gets upgrade plan (if any)
	// 2. If DowngradeVerified() is false (first run), checks for downgrade
	// 3. Downgrade check calls GetLastCompletedUpgrade which iterates done keys
	// 4. Legacy done keys cause panic in parseDoneKey

	version.Version = "v1.3.0"
	app.WasmEnabled = "false"
	t.Cleanup(cleanup)

	axelarApp := app.NewAxelarApp(
		log.NewTestLogger(t),
		dbm.NewMemDB(),
		nil,
		true,
		app.MakeEncodingConfig(),
		simtestutil.EmptyAppOptions{},
		[]wasm.Option{},
	)

	// Must set HeaderInfo for SDK v0.50+ - PreBlocker uses HeaderInfo().Height
	ctx := axelarApp.NewUncachedContext(false, cmtproto.Header{Height: 100}).WithHeaderInfo(header.Info{
		Height: 100,
		Time:   time.Now(),
	})
	kvStore := ctx.KVStore(axelarApp.Keys[upgradetypes.StoreKey])

	// Write a legacy "done" key
	legacyDoneKey := append([]byte{upgradetypes.DoneByte}, []byte("v1.2")...)
	kvStore.Set(legacyDoneKey, []byte{1})

	upgradeKeeper := app.GetKeeper[upgradekeeper.Keeper](axelarApp.Keepers)

	// PreBlocker should panic because:
	// 1. No plan exists, so downgrade verification runs
	// 2. GetLastCompletedUpgrade is called
	// 3. Legacy done key causes panic
	assert.Panics(t, func() {
		_, _ = upgrade.PreBlocker(ctx, upgradeKeeper)
	}, "PreBlocker should panic on legacy done keys")
}

func TestDeleteUpgradeDoneKeysPreventsGetLastCompletedUpgradePanic(t *testing.T) {
	// This test verifies that after deleting done keys using DeleteUpgradeDoneKeys,
	// GetLastCompletedUpgrade no longer panics.

	version.Version = "v1.3.0"
	app.WasmEnabled = "false"
	t.Cleanup(cleanup)

	logger := log.NewTestLogger(t)
	axelarApp := app.NewAxelarApp(
		logger,
		dbm.NewMemDB(),
		nil,
		true,
		app.MakeEncodingConfig(),
		simtestutil.EmptyAppOptions{},
		[]wasm.Option{},
	)

	// Must set HeaderInfo for SDK v0.50+ - PreBlocker uses HeaderInfo().Height
	ctx := axelarApp.NewUncachedContext(false, cmtproto.Header{Height: 100}).WithHeaderInfo(header.Info{
		Height: 100,
		Time:   time.Now(),
	})
	kvStore := ctx.KVStore(axelarApp.Keys[upgradetypes.StoreKey])

	// Write legacy "done" keys
	legacyDoneKey1 := append([]byte{upgradetypes.DoneByte}, []byte("v1.0")...)
	legacyDoneKey2 := append([]byte{upgradetypes.DoneByte}, []byte("v1.1")...)
	legacyDoneKey3 := append([]byte{upgradetypes.DoneByte}, []byte("v1.2")...)
	kvStore.Set(legacyDoneKey1, []byte{1})
	kvStore.Set(legacyDoneKey2, []byte{1})
	kvStore.Set(legacyDoneKey3, []byte{1})

	// Verify keys exist before deletion
	require.True(t, kvStore.Has(legacyDoneKey1))
	require.True(t, kvStore.Has(legacyDoneKey2))
	require.True(t, kvStore.Has(legacyDoneKey3))

	// Use the actual exported function to delete done keys
	app.DeleteUpgradeDoneKeys(kvStore, logger)

	// Verify keys are deleted
	require.False(t, kvStore.Has(legacyDoneKey1))
	require.False(t, kvStore.Has(legacyDoneKey2))
	require.False(t, kvStore.Has(legacyDoneKey3))

	// Now GetLastCompletedUpgrade should NOT panic
	upgradeKeeper := app.GetKeeper[upgradekeeper.Keeper](axelarApp.Keepers)
	assert.NotPanics(t, func() {
		name, height, err := upgradeKeeper.GetLastCompletedUpgrade(ctx)
		// With no done keys, it should return empty values
		assert.NoError(t, err)
		assert.Empty(t, name)
		assert.Zero(t, height)
	}, "GetLastCompletedUpgrade should not panic after legacy keys are deleted")
}

func TestDeleteUpgradeDoneKeysPreventsPreBlockerPanic(t *testing.T) {
	// This test verifies that after deleting done keys using DeleteUpgradeDoneKeys,
	// PreBlocker no longer panics. This is the key test that proves our fix works.
	//
	// The fix in scheduleUpgradeFromDisk:
	// 1. Calls DeleteUpgradeDoneKeys to remove all done keys
	// 2. Writes the upgrade plan directly to store
	//
	// After this, PreBlocker can safely call GetLastCompletedUpgrade.

	version.Version = "v1.3.0"
	app.WasmEnabled = "false"
	t.Cleanup(cleanup)

	logger := log.NewTestLogger(t)
	axelarApp := app.NewAxelarApp(
		logger,
		dbm.NewMemDB(),
		nil,
		true,
		app.MakeEncodingConfig(),
		simtestutil.EmptyAppOptions{},
		[]wasm.Option{},
	)

	// Must set HeaderInfo for SDK v0.50+ - PreBlocker uses HeaderInfo().Height
	ctx := axelarApp.NewUncachedContext(false, cmtproto.Header{Height: 100}).WithHeaderInfo(header.Info{
		Height: 100,
		Time:   time.Now(),
	})
	kvStore := ctx.KVStore(axelarApp.Keys[upgradetypes.StoreKey])

	// Write legacy "done" keys (simulating pre-v0.50 state)
	legacyDoneKey1 := append([]byte{upgradetypes.DoneByte}, []byte("v1.0")...)
	legacyDoneKey2 := append([]byte{upgradetypes.DoneByte}, []byte("v1.1")...)
	kvStore.Set(legacyDoneKey1, []byte{1})
	kvStore.Set(legacyDoneKey2, []byte{1})

	// Use the actual exported function to delete done keys
	app.DeleteUpgradeDoneKeys(kvStore, logger)

	// Write the upgrade plan directly to store (as scheduleUpgradeFromDisk does)
	plan := upgradetypes.Plan{
		Name:   "v1.3",
		Height: 100,
		Info:   "test upgrade",
	}
	bz, err := axelarApp.AppCodec().Marshal(&plan)
	require.NoError(t, err)
	kvStore.Set(upgradetypes.PlanKey(), bz)

	upgradeKeeper := app.GetKeeper[upgradekeeper.Keeper](axelarApp.Keepers)

	// Verify plan was written correctly
	foundPlan, err := upgradeKeeper.GetUpgradePlan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "v1.3", foundPlan.Name)

	// PreBlocker should NOT panic from legacy keys - they've been deleted.
	// With HeaderInfo().Height = 100 and plan.Height = 100, the upgrade will execute.
	// The upgrade may panic during execution due to uninitialized genesis state in tests,
	// but that's fine - we just need to confirm it's not panicking from legacy done keys
	// and that the upgrade actually started executing.
	func() {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				panicMsg := fmt.Sprintf("%v", r)

				// The panic should NOT be from parseDoneKey (legacy key issue)
				assert.NotContains(t, panicMsg, "parseDoneKey", "panic should not be from legacy done key parsing")
				assert.NotContains(t, stack, "parseDoneKey", "stack should not contain parseDoneKey")

				// Verify the upgrade actually started - stack should show ApplyUpgrade was called
				assert.Contains(t, stack, "ApplyUpgrade", "panic should occur during upgrade execution, confirming upgrade was triggered")

				t.Logf("PreBlocker panicked during upgrade execution (expected in test): %v", r)
			}
		}()
		_, _ = upgrade.PreBlocker(ctx, upgradeKeeper)
	}()
}
