package app_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestMigratePreInitializedModuleAccounts(t *testing.T) {
	var (
		accountK authkeeper.AccountKeeper
		ctx      sdk.Context
	)

	Given("an account keeper", func() {
		encodingConfig := app.MakeEncodingConfig()
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		storeKey := sdk.NewKVStoreKey(authtypes.StoreKey)

		moduleAccPerms := map[string][]string{
			"module1":             nil,
			nexusTypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		}

		accountK = authkeeper.NewAccountKeeper(
			encodingConfig.Codec,
			sdk.NewKVStoreKey(authtypes.StoreKey),
			params.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, storeKey, storeKey, authtypes.ModuleName),
			authtypes.ProtoBaseAccount,
			moduleAccPerms,
		)
	}).When("there is an pre-initialized module account", func() {
		account := accountK.NewAccountWithAddress(ctx, authtypes.NewModuleAddress(nexusTypes.ModuleName))
		accountK.SetAccount(ctx, account)

		account = accountK.GetAccount(ctx, authtypes.NewModuleAddress(nexusTypes.ModuleName))
		_, isModuleAccount := account.(authtypes.ModuleAccountI)
		assert.False(t, isModuleAccount)

	}).Then("migrating pre-initialized base account to module account", func(t *testing.T) {
		err := app.MigratePreInitializedModuleAccounts(ctx, accountK, []string{"module1", nexusTypes.ModuleName})
		assert.NoError(t, err)

		account := accountK.GetAccount(ctx, authtypes.NewModuleAddress(nexusTypes.ModuleName))
		_, isModuleAccount := account.(authtypes.ModuleAccountI)
		assert.True(t, isModuleAccount)

		account = accountK.GetAccount(ctx, authtypes.NewModuleAddress("module1"))
		_, isModuleAccount = account.(authtypes.ModuleAccountI)
		assert.True(t, isModuleAccount)
	}).Run(t)
}
