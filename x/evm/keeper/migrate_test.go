package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (sdk.Context, BaseKeeper) {
	encCfg := params.MakeEncodingConfig()
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("evm"), paramsK)

	for _, params := range types.DefaultParams() {
		keeper.ForChain(params.Chain).SetParams(ctx, params)
	}

	return ctx, keeper
}

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx     sdk.Context
		keeper  BaseKeeper
		nexus   *mock.NexusMock
		handler func(ctx sdk.Context) error
	)

	givenHandler := Given("the migration handler", func() {
		ctx, keeper = setup()
		nexus = &mock.NexusMock{}
		handler = GetMigrationHandler(keeper, nexus)
	})

	givenHandler.
		When("contract bytecode is out-of-date for some EVM chain", func() {
			chain := types.DefaultParams()[0].Chain

			ck := keeper.ForChain(chain)
			subspace, ok := ck.(chainKeeper).getSubspace(ctx)
			if !ok {
				panic(fmt.Errorf("param subspace for chain %s should exist", ck.GetName()))
			}
			subspace.Set(ctx, types.KeyToken, rand.Bytes(100))
			subspace.Set(ctx, types.KeyBurnable, rand.Bytes(100))

			nexus.GetChainsFunc = func(ctx sdk.Context) []exported.Chain {
				return []exported.Chain{{Name: chain, Module: types.ModuleName}}
			}
		}).
		Then("should update bytecode", func(t *testing.T) {
			err := handler(ctx)
			assert.NoError(t, err)

			params := types.DefaultParams()[0]
			assert.Equal(t, params, keeper.ForChain(params.Chain).GetParams(ctx))
		}).
		Run(t)

}
