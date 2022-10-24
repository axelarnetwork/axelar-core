package keeper_test

import (
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/keeper"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeygenOptOut(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(types.StoreKey), testutils.NewSubspace(encCfg))

	participant := rand.AccAddr()

	ctx := testutils.NewContext()
	assert.False(t, k.IsOptOut(ctx, participant))
	k.KeygenOptOut(ctx, participant)
	assert.True(t, k.IsOptOut(ctx, participant))
	k.KeygenOptIn(ctx, participant)
	assert.False(t, k.IsOptOut(ctx, participant))
	k.KeygenOptOut(ctx, participant)
	assert.True(t, k.IsOptOut(ctx, participant))
}
