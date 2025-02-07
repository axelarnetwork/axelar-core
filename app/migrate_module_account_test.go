package app_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	. "github.com/axelarnetwork/utils/test"
)

func TestMigrateDistributionAccountPermission(t *testing.T) {
	var (
		accountK authkeeper.AccountKeeper
		ctx      sdk.Context
		accNum   uint64
	)

	Given("a distribution account initialized with minter permission", func() {
		encodingConfig := app.MakeEncodingConfig()
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		storeKey := sdk.NewKVStoreKey(authtypes.StoreKey)

		moduleAccPerms := map[string][]string{
			distributionTypes.ModuleName: {authtypes.Minter},
		}

		accountK = authkeeper.NewAccountKeeper(
			encodingConfig.Codec,
			sdk.NewKVStoreKey(authtypes.StoreKey),
			params.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, storeKey, storeKey, authtypes.ModuleName),
			authtypes.ProtoBaseAccount,
			moduleAccPerms,
		)
		accountK.GetNextAccountNumber(ctx)

		acc := accountK.GetModuleAccount(ctx, distributionTypes.ModuleName)
		assert.False(t, acc.HasPermission(authtypes.Burner))
		accNum = acc.GetAccountNumber()

	}).When("run migration", func() {
		app.MigrateDistributionAccountPermission(ctx, accountK)

	}).Then("distribution account permission should also have burner permission", func(t *testing.T) {
		acc := accountK.GetModuleAccount(ctx, distributionTypes.ModuleName)
		assert.True(t, acc.HasPermission(authtypes.Burner))
		assert.Equal(t, accNum, acc.GetAccountNumber())
	}).Run(t)
}
