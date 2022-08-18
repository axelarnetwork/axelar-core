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
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	axelarnettestutils "github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	"github.com/axelarnetwork/utils/slices"
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

		transfers []types.IBCTransfer
	)

	givenMigrationHandler := Given("the migration handler", func() {
		ctx, keeper = setup()
		handler = GetMigrationHandler(keeper)
	})

	whenTransfersExist := When("IBC transfers exists", func() {
		transfers = slices.Expand(func(_ int) types.IBCTransfer {
			transfer := axelarnettestutils.RandomIBCTransfer()
			transfer.Status = types.TransferNonExistent
			return transfer
		}, int(rand.I64Between(50, 100)))
	})

	whenMigrationRuns := When("migration runs", func() {
		err := handler(ctx)
		assert.NoError(t, err)
	})

	givenMigrationHandler.
		Branch(
			When("no IBC transfers exists", func() {}).
				When2(whenMigrationRuns).
				Then("should do nothing", func(t *testing.T) {
					assert.Zero(t, keeper.getIBCTransfers(ctx))
				}),
			whenTransfersExist.
				When("", func() {
					slices.ForEach(transfers, func(t types.IBCTransfer) { keeper.setTransfer(ctx, t) })
				}).
				When2(whenMigrationRuns).
				Then("transfers set to completed", func(t *testing.T) {
					transfers := keeper.getIBCTransfers(ctx)
					assert.NotEmpty(t, transfers)
					slices.All(keeper.getIBCTransfers(ctx), func(t types.IBCTransfer) bool { return t.Status == types.TransferCompleted })
				}),
		).Run(t)

	givenMigrationHandler.
		Branch(
			When("no failed IBC transfers exists", func() {}).
				When2(whenMigrationRuns).
				Then("should do nothing", func(t *testing.T) {
					assert.Zero(t, getFailedTransfers(ctx, keeper))
				}),

			whenTransfersExist.
				When("", func() {
					slices.ForEach(transfers, func(t types.IBCTransfer) { setFailedTransfer(ctx, keeper, t) })
					assert.NotEmpty(t, getFailedTransfers(ctx, keeper))
				}).
				When2(whenMigrationRuns).
				Then("should delete failed Transfers transfers", func(t *testing.T) {
					assert.Empty(t, getFailedTransfers(ctx, keeper))
				}),
		).Run(t)
}

func getFailedTransfers(ctx sdk.Context, k Keeper) (failedTransfers []types.IBCTransfer) {
	iter := k.getStore(ctx).IteratorNew(failedTransferPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var t types.IBCTransfer
		iter.UnmarshalValue(&t)

		failedTransfers = append(failedTransfers, t)
	}

	return failedTransfers
}

func setFailedTransfer(ctx sdk.Context, k Keeper, transfer types.IBCTransfer) {
	k.getStore(ctx).SetNew(failedTransferPrefix.Append(key.FromBz(transfer.ID.Bytes())), &transfer)
}
