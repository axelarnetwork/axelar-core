package keeper_test

import (
	"testing"

	"github.com/axelarnetwork/axelar-core/x/permission/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/permission/keeper"
	. "github.com/axelarnetwork/utils/test"
)

func TestGrpcQuery(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	var (
		k              keeper.Keeper
		ctx            sdk.Context
		initialGenesis *types.GenesisState
	)

	Given("a keeper",
		func() {
			subspace := paramstypes.NewSubspace(cfg.Codec, cfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "permission")
			k = keeper.NewKeeper(cfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace)
		}).
		When("the state is initialized from a genesis state",
			func() {
				initialGenesis = types.NewGenesisState(types.Params{}, randomMultisigGovernanceKey(), randomGovAccounts())
				assert.NoError(t, initialGenesis.Validate())

				ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
				k.InitGenesis(ctx, initialGenesis)
			}).
		Then("query the governance key",
			func(t *testing.T) {
				req := &types.QueryGovernanceKeyRequest{}
				resp, err := k.GovernanceKey(sdk.WrapSDKContext(ctx), req)
				assert.NotNil(t, resp)
				assert.Nil(t, err)
				assert.Equal(t, *initialGenesis.GovernanceKey, resp.GovernanceKey)
			}).
		Then("query the params",
			func(t *testing.T) {
				req := &types.ParamsRequest{}
				resp, err := k.Params(sdk.WrapSDKContext(ctx), req)
				assert.NotNil(t, resp)
				assert.Nil(t, err)
				assert.Equal(t, initialGenesis.Params, resp.Params)
			}).Run(t, 10)
}
