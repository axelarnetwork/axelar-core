package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx         sdk.Context
		cdc         codec.Codec
		storekey    = sdk.NewKVStoreKey(types.StoreKey)
		k           *keeper.BaseKeeper
		handler     func(ctx sdk.Context) error
		err         error
		burnerInfos []types.BurnerInfo
	)

	Given("a context", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	}).
		Given("a keeper", func() {
			encCfg := params.MakeEncodingConfig()
			cdc = encCfg.Codec
			pk := paramsKeeper.NewKeeper(cdc, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))

			k = keeper.NewKeeper(cdc, storekey, pk)
			k.InitChains(ctx)
			funcs.MustNoErr(k.CreateChain(ctx, types.DefaultParams()[0]))
		}).
		Given("a migration handler", func() {
			n := &mock.NexusMock{
				GetChainsFunc: func(sdk.Context) []nexus.Chain { return []nexus.Chain{exported.Ethereum} },
			}
			handler = keeper.Migrate7To8(k, n)
		}).
		Given("there are only old burner infos", func() {
			burnerInfos = slices.Expand2(testutils.RandomBurnerInfo, 20)

			for i := 0; i < 10; i++ {
				info := burnerInfos[i]
				ctx.KVStore(storekey).Set(
					[]byte(fmt.Sprintf("chain_%s_burnerAddr_%s", strings.ToLower(exported.Ethereum.Name.String()), info.BurnerAddress.Hex())),
					cdc.MustMarshalLengthPrefixed(&info))
			}
			for i := 10; i < 20; i++ {
				info := burnerInfos[i]
				ctx.KVStore(storekey).Set(
					[]byte(fmt.Sprintf("chain_%s_burneraddr_%s", strings.ToLower(exported.Ethereum.Name.String()), info.BurnerAddress.Hex())),
					cdc.MustMarshalLengthPrefixed(&info))
			}
		}).
		When("calling migration", func() {
			err = handler(ctx)
		}).
		Then("it succeeds", func(t *testing.T) {
			assert.NoError(t, err)
		}).
		Then("all burner infos have been migrated", func(t *testing.T) {
			ck := funcs.Must(k.ForChain(ctx, exported.Ethereum.Name))

			for i, info := range burnerInfos {
				assert.Equal(t, &burnerInfos[i], ck.GetBurnerInfo(ctx, info.BurnerAddress))
			}
		}).Run(t, 20)
}
