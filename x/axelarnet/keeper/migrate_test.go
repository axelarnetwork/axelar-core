package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (sdk.Context, Keeper) {
	encCfg := params.MakeEncodingConfig()
	subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), types.ModuleName)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), subspace, &mock.ChannelKeeperMock{})
	keeper.setParams(ctx, types.DefaultParams())

	return ctx, keeper
}

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
