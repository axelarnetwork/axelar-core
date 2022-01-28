package keeper

import (
	"testing"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/evm/legacy"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func setup() (sdk.Context, types.BaseKeeper, string) {
	encCfg := params.MakeEncodingConfig()
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("evm"), paramsK)
	chain := "Ethereum"

	return ctx, keeper, chain
}

func TestGetMigrationHandler_addAbsorberBytecode(t *testing.T) {
	ctx, baseKeeper, chain := setup()
	nexus := mock.NexusMock{
		GetChainsFunc: func(_ sdk.Context) []exported.Chain {
			return []exported.Chain{{
				Name:   chain,
				Module: types.ModuleName,
			}}
		},
	}

	legacyParams := legacy.Params{Params: types.DefaultParams()[0]}
	keeper := baseKeeper.ForChain(chain).(chainKeeper)
	keeper.getBaseStore(ctx).SetRaw(subspacePrefix.AppendStr(keeper.chainLowerKey), []byte(chain))
	subspace, ok := keeper.getSubspace(ctx)
	if !ok {
		panic("subspace not found")
	}

	subspace.SetParamSet(ctx, &legacyParams)
	assert.Panics(t, func() { keeper.GetParams(ctx) })

	handler := GetMigrationHandler(keeper, &nexus)
	handler(ctx)

	assert.NotPanics(t, func() { keeper.GetParams(ctx) })
	actual := keeper.GetParams(ctx)
	assert.Equal(t, actual, types.DefaultParams()[0])
}
