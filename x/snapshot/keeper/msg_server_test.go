package keeper

import (
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	snapmock "github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
)

func TestMsgServer_UpdateParams(t *testing.T) {
	enc := appParams.MakeEncodingConfig()
	subspace := paramstypes.NewSubspace(enc.Codec, enc.Amino, store.NewKVStoreKey("snapshotParams"), store.NewKVStoreKey("tsnapshotParams"), "snapshot")
	k := NewKeeper(enc.Codec, store.NewKVStoreKey("snapshot"), subspace, &snapmock.StakingKeeperMock{}, &snapmock.BankKeeperMock{}, &snapmock.SlasherMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), sdk.Context{}.BlockHeader(), false, log.NewTestLogger(t))
	server := NewMsgServerImpl(k)
	querier := NewGRPCQuerier(k)

	p := types.DefaultParams()
	p.MinProxyBalance = p.MinProxyBalance + 1
	_, err := server.UpdateParams(ctx, &types.UpdateParamsRequest{Authority: "", Params: p})
	assert.NoError(t, err)
	got, err := querier.Params(ctx, &types.ParamsRequest{})
	assert.NoError(t, err)
	assert.Equal(t, p, got.Params)
}
