package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/feepolicy/keeper"
	"github.com/axelarnetwork/axelar-core/x/feepolicy/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestQueryParams(t *testing.T) {
	var (
		ctx sdk.Context
		k   keeper.Keeper
		q   keeper.Querier
		res *types.ParamsResponse
		err error
	)

	given := Given("a feepolicy keeper", func() {
		ctx, k = setup(t)
		q = keeper.NewGRPCQuerier(k)
	})

	given.
		When("params are queried", func() {
			res, err = q.Params(sdk.WrapSDKContext(ctx), &types.ParamsRequest{})
		}).
		Then("the default allowlist is returned", func(t *testing.T) {
			assert.NoError(t, err)
			assert.Equal(t, []string{"uaxl"}, res.Params.AllowedFeeDenoms)
		}).
		Run(t)

	given.
		When("the allowlist is updated to add uusdc", func() {
			_, err = keeper.NewMsgServerImpl(k).UpdateParams(sdk.WrapSDKContext(ctx), &types.UpdateParamsRequest{
				Params: types.Params{AllowedFeeDenoms: []string{"uaxl", "uusdc"}},
			})
		}).
		Then("querying params reflects the update", func(t *testing.T) {
			assert.NoError(t, err)

			res, err = q.Params(sdk.WrapSDKContext(ctx), &types.ParamsRequest{})
			assert.NoError(t, err)
			assert.Equal(t, []string{"uaxl", "uusdc"}, res.Params.AllowedFeeDenoms)
		}).
		Run(t)
}
