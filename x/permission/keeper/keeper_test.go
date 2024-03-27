package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/keeper"
)

func TestKeeper_GetRole_nil_Address_Return_Unrestricted(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	key := sdk.NewKVStoreKey("permission")
	subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, key, sdk.NewKVStoreKey("trewardKey"), "reward")
	k := keeper.NewKeeper(encCfg.Codec, key, subspace)

	ctx := sdk.NewContext(fake.NewMultiStore(), sdk.Context{}.BlockHeader(), false, log.TestingLogger())
	assert.Equal(t, k.GetRole(ctx, nil), exported.ROLE_UNRESTRICTED)
}
