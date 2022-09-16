package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx     sdk.Context
		keeper  Keeper
		handler func(ctx sdk.Context) error
	)

	givenMigrationHandler := Given("the migration handler", func() {
		ctx, keeper = setup()
		handler = GetMigrationHandler(keeper)
	})

	givenMigrationHandler.
		When("TransferLimit param is not set", func() {
			keeper.params.Set(ctx, types.KeyTransferLimit, uint64(0))
		}).
		Then("should set EndBlockerLimit param", func(t *testing.T) {
			assert.Zero(t, keeper.getParams(ctx).TransferLimit)

			err := handler(ctx)
			assert.NoError(t, err)

			assert.Equal(t, types.DefaultParams().TransferLimit, keeper.getParams(ctx).TransferLimit)
		}).
		Run(t)
}
