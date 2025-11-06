package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	tssmock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

func TestMsgServer_UpdateParams(t *testing.T) {
	enc := appParams.MakeEncodingConfig()
	subspace := paramstypes.NewSubspace(enc.Codec, enc.Amino, store.NewKVStoreKey("tssParams"), store.NewKVStoreKey("ttssParams"), "tss")
	k := keeper.NewKeeper(enc.Codec, store.NewKVStoreKey(types.StoreKey), subspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), sdk.Context{}.BlockHeader(), false, log.NewTestLogger(t))
	server := keeper.NewMsgServerImpl(k, &tssmock.SnapshotterMock{}, &tssmock.StakingKeeperMock{}, &tssmock.MultiSigKeeperMock{})
	querier := keeper.NewGRPCQuerier(k)

	p := types.DefaultParams()
	p.HeartbeatPeriodInBlocks = p.HeartbeatPeriodInBlocks + 1
	_, err := server.UpdateParams(ctx, &types.UpdateParamsRequest{Authority: rand.AccAddr().String(), Params: p})
	assert.NoError(t, err)
	got, err := querier.Params(ctx, &types.ParamsRequest{})
	assert.NoError(t, err)
	assert.Equal(t, p, got.Params)
}
