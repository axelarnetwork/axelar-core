package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestMigrate4To5(t *testing.T) {
	var (
		ctx     sdk.Context
		k       keeper.Keeper
		handler func(ctx sdk.Context) error
		err     error
	)
	repeats := 10

	givenKeeper := Given("a keeper", func() {
		ctx, k, _ = setup()
	})

	givenKeeper.
		When("cosmos chains are set", func() {
			for _, chain := range randomChains() {
				k.SetCosmosChain(ctx, chain)
			}
		}).
		Then("migration should succeed", func(t *testing.T) {
			handler = keeper.Migrate4To5(k)
			err = handler(ctx)
			assert.NoError(t, err)
		}).
		Then("ibc path mapping should be set", func(t *testing.T) {
			for _, chainName := range k.GetCosmosChains(ctx) {
				chain := funcs.MustOk(k.GetCosmosChainByName(ctx, chainName))
				chain2, found := k.GetChainNameByIBCPath(ctx, chain.IBCPath)
				assert.True(t, found)
				assert.Equal(t, chainName, chain2)
			}
		}).
		Run(t, repeats)

	givenKeeper.
		When("cosmos chains are set", func() {
			for _, chain := range randomChains() {
				k.SetCosmosChain(ctx, chain)
			}
		}).
		When("cosmos chain with duplicate ibc path is set", func() {
			chain := testutils.RandomCosmosChain()
			chain2 := testutils.RandomCosmosChain()
			chain2.IBCPath = chain.IBCPath
			k.SetCosmosChain(ctx, chain)
			k.SetCosmosChain(ctx, chain2)
		}).
		Then("migration should fail", func(t *testing.T) {
			handler = keeper.Migrate4To5(k)
			err = handler(ctx)
			assert.ErrorContains(t, err, "already registered")
		}).
		Run(t, repeats)

	givenKeeper.
		When("axelarnet is set", func() {
			k.SetCosmosChain(ctx, types.CosmosChain{
				Name:       exported.Axelarnet.Name,
				IBCPath:    "",
				AddrPrefix: "axelar",
			})
		}).
		Then("migration should succeed", func(t *testing.T) {
			handler = keeper.Migrate4To5(k)
			err = handler(ctx)
			assert.NoError(t, err)
		}).
		Then("ibc path mapping should not be set for axelarnet", func(t *testing.T) {
			chain := funcs.MustOk(k.GetCosmosChainByName(ctx, exported.Axelarnet.Name))
			_, found := k.GetChainNameByIBCPath(ctx, chain.IBCPath)
			assert.False(t, found)
		}).
		Run(t)
}
