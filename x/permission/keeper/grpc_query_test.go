package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/permission/keeper"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestGrpcQuery(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	var (
		k              keeper.Keeper
		ctx            sdk.Context
		initialGenesis *types.GenesisState
		governanceKey  multisig.LegacyAminoPubKey
		params         types.Params
	)

	Given("a keeper",
		func() {
			subspace := paramstypes.NewSubspace(cfg.Codec, cfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "permission")
			k = keeper.NewKeeper(cfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace)
		}).
		Given("a state that is initialized",
			func() {
				initialGenesis = types.NewGenesisState(types.Params{}, randomMultisigGovernanceKey(), randomGovAccounts())
				assert.NoError(t, initialGenesis.Validate())

				ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
				k.InitGenesis(ctx, initialGenesis)
			}).
		When("querying the governance key",
			func() {
				req := &types.QueryGovernanceKeyRequest{}
				resp, err := k.GovernanceKey(sdk.WrapSDKContext(ctx), req)
				assert.NotNil(t, resp)
				assert.Nil(t, err)
				governanceKey = resp.GovernanceKey
			}).
		Then("return the expected key",
			func(t *testing.T) {
				assert.Equal(t, *initialGenesis.GovernanceKey, governanceKey)
			}).Run(t, 10)

	Given("a keeper",
		func() {
			subspace := paramstypes.NewSubspace(cfg.Codec, cfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "permission")
			k = keeper.NewKeeper(cfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace)
		}).
		Given("a state that is initialized",
			func() {
				initialGenesis = types.NewGenesisState(types.Params{}, randomMultisigGovernanceKey(), randomGovAccounts())
				assert.NoError(t, initialGenesis.Validate())

				ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
				k.InitGenesis(ctx, initialGenesis)
			}).
		When("querying params",
			func() {
				req := &types.ParamsRequest{}
				resp, err := k.Params(sdk.WrapSDKContext(ctx), req)
				assert.NotNil(t, resp)
				assert.Nil(t, err)
				params = resp.Params
			}).
		Then("return the expected params",
			func(t *testing.T) {
				assert.Equal(t, initialGenesis.Params, params)
			}).Run(t, 10)
}
