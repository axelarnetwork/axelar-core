package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/keeper"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

func TestKeeper_GetRole_nil_Address_Return_Unrestricted(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	key := store.NewKVStoreKey("permission")
	subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, key, store.NewKVStoreKey("trewardKey"), "reward")
	k := keeper.NewKeeper(encCfg.Codec, key, subspace)

	ctx := sdk.NewContext(fake.NewMultiStore(), sdk.Context{}.BlockHeader(), false, log.NewTestLogger(t))
	assert.Equal(t, k.GetRole(ctx, nil), exported.ROLE_UNRESTRICTED)
}

func TestMsgServer_UpdateParams(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	key := store.NewKVStoreKey("permission")
	subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, key, store.NewKVStoreKey("tpermissionKey"), "permission")
	k := keeper.NewKeeper(encCfg.Codec, key, subspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), sdk.Context{}.BlockHeader(), false, log.NewTestLogger(t))

	server := keeper.NewMsgServerImpl(k)

	p := types.DefaultParams()
	_, err := server.UpdateParams(ctx, &types.UpdateParamsRequest{Authority: "", Params: p})
	assert.NoError(t, err)
	got := k.GetParams(ctx)
	assert.Equal(t, p, got)
}
