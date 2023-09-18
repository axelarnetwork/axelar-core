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
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func TestMigrate5to6(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), subspace, &mock.ChannelKeeperMock{}, &mock.FeegrantKeeperMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

	Given("subspace is setup with params before migration", func() {
		subspace.Set(ctx, types.KeyRouteTimeoutWindow, types.DefaultParams().RouteTimeoutWindow)
		subspace.Set(ctx, types.KeyTransferLimit, types.DefaultParams().TransferLimit)
		subspace.Set(ctx, types.KeyEndBlockerLimit, types.DefaultParams().EndBlockerLimit)
	}).
		When("", func() {}).
		Then("the migration should add the new param with the default value", func(t *testing.T) {
			actual := types.CallContractProposalMinDeposits{}

			assert.PanicsWithError(t, "UnmarshalJSON cannot decode empty bytes", func() {
				subspace.Get(ctx, types.KeyCallContractsProposalMinDeposits, &actual)
			})
			assert.PanicsWithError(t, "UnmarshalJSON cannot decode empty bytes", func() {
				k.GetParams(ctx)
			})

			keeper.Migrate5to6(k)(ctx)

			assert.NotPanics(t, func() {
				subspace.Get(ctx, types.KeyCallContractsProposalMinDeposits, &actual)
			})
			assert.NotPanics(t, func() {
				k.GetParams(ctx)
			})

			assert.Equal(t, types.DefaultParams().CallContractsProposalMinDeposits, actual)
		}).
		Run(t)
}
