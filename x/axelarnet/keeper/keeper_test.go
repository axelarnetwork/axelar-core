package keeper_test

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	axelarnetKeeper "github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
)

func TestKeeper_GetIbcPath(t *testing.T) {
	var (
		ctx    sdk.Context
		keeper axelarnetKeeper.Keeper
	)
	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = axelarnetKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("axelarnet"))
	}
	t.Run("should return the registered IBC path when the given asset is registered", testutils.Func(func(t *testing.T) {
		setup()
		path := randomIbcPath()
		asset := rand.StrBetween(5, 20)
		err := keeper.RegisterIbcPath(ctx, asset, path)
		assert.NoError(t, err)
		result := keeper.GetIbcPath(ctx, asset)
		assert.Equal(t, path, result)
	}).Repeat(20))

	t.Run("should return error when registered the same asset twice", testutils.Func(func(t *testing.T) {
		setup()
		path := randomIbcPath()
		asset := rand.StrBetween(5, 20)
		err := keeper.RegisterIbcPath(ctx, asset, path)
		assert.NoError(t, err)
		path2 := randomIbcPath()
		err2 := keeper.RegisterIbcPath(ctx, asset, path2)
		assert.Error(t, err2)
	}).Repeat(20))

}
func randomIbcPath() string {
	port := rand.StrBetween(5, 10)
	identifier := fmt.Sprintf("%s%d", "channel-", rand.I64Between(0, 9999))
	return port + "/" + identifier
}
