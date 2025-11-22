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
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types/mock"
)

func TestMsgServer_UpdateParams(t *testing.T) {
	enc := appParams.MakeEncodingConfig()
	subspace := paramstypes.NewSubspace(enc.Codec, enc.Amino, store.NewKVStoreKey("voteParams"), store.NewKVStoreKey("tvoteParams"), "vote")
	k := keeper.NewKeeper(enc.Codec, store.NewKVStoreKey(types.StoreKey), subspace, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.RewarderMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), sdk.Context{}.BlockHeader(), false, log.NewTestLogger(t))
	server := keeper.NewMsgServerImpl(k)

	p := types.DefaultParams()
	p.EndBlockerLimit = p.EndBlockerLimit + 1
	_, err := server.UpdateParams(ctx, &types.UpdateParamsRequest{Authority: rand.AccAddr().String(), Params: p})
	assert.NoError(t, err)
	got := k.GetParams(ctx)
	assert.Equal(t, p, got)
}
