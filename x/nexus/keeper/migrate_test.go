package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	nexuskeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestMigrate6to7(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
	keeper := nexuskeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("nexus"), subspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

	Given("subspace is setup with params before migration", func() {
		subspace.Set(ctx, types.KeyChainActivationThreshold, types.DefaultParams().ChainActivationThreshold)
		subspace.Set(ctx, types.KeyChainMaintainerMissingVoteThreshold, types.DefaultParams().ChainMaintainerMissingVoteThreshold)
		subspace.Set(ctx, types.KeyChainMaintainerIncorrectVoteThreshold, types.DefaultParams().ChainMaintainerIncorrectVoteThreshold)
		subspace.Set(ctx, types.KeyChainMaintainerCheckWindow, types.DefaultParams().ChainMaintainerCheckWindow)
	}).
		When("", func() {}).
		Then("the migration should add the new param with the default value", func(t *testing.T) {
			actual := make(map[string]types.Params_Coins)

			assert.PanicsWithError(t, "UnmarshalJSON cannot decode empty bytes", func() {
				subspace.Get(ctx, types.KeyCallContractsProposalMinDeposits, &actual)
			})
			assert.PanicsWithError(t, "UnmarshalJSON cannot decode empty bytes", func() {
				keeper.GetParams(ctx)
			})

			nexuskeeper.Migrate6to7(keeper)(ctx)

			assert.NotPanics(t, func() {
				subspace.Get(ctx, types.KeyCallContractsProposalMinDeposits, &actual)
			})
			assert.NotPanics(t, func() {
				keeper.GetParams(ctx)
			})

			assert.Equal(t, types.DefaultParams().CallContractsProposalMinDeposits, actual)
		}).
		Run(t)
}
