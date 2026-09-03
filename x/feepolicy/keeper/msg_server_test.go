package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	appparams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/feepolicy/keeper"
	"github.com/axelarnetwork/axelar-core/x/feepolicy/types"
	. "github.com/axelarnetwork/utils/test"
)

func setup(t *testing.T) (sdk.Context, keeper.Keeper) {
	encCfg := appparams.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("feepolicy"), store.NewKVStoreKey("tfeepolicy"), types.ModuleName)
	k := keeper.NewKeeper(subspace)
	k.InitGenesis(ctx, types.DefaultGenesisState())

	return ctx, k
}

func TestUpdateParams(t *testing.T) {
	var (
		ctx sdk.Context
		k   keeper.Keeper
		err error
	)

	given := Given("a feepolicy keeper", func() {
		ctx, k = setup(t)
	})

	given.
		When("a valid update sets a new allowlist", func() {
			_, err = keeper.NewMsgServerImpl(k).UpdateParams(sdk.WrapSDKContext(ctx), &types.UpdateParamsRequest{
				Params: types.Params{AllowedFeeDenoms: []string{"uaxl", "uusdc"}},
			})
		}).
		Then("the allowlist is replaced", func(t *testing.T) {
			assert.NoError(t, err)
			assert.Equal(t, []string{"uaxl", "uusdc"}, k.GetAllowedFeeDenoms(ctx))
		}).
		Run(t)

	given.
		When("an invalid empty allowlist is submitted", func() {
			_, err = keeper.NewMsgServerImpl(k).UpdateParams(sdk.WrapSDKContext(ctx), &types.UpdateParamsRequest{
				Params: types.Params{AllowedFeeDenoms: nil},
			})
		}).
		Then("it is rejected and the allowlist is unchanged", func(t *testing.T) {
			assert.Error(t, err)
			assert.Equal(t, []string{"uaxl"}, k.GetAllowedFeeDenoms(ctx))
		}).
		Run(t)
}
