package keeper_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	axelartestutils "github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
)

func setup() (sdk.Context, keeper.Keeper) {
	encCfg := appParams.MakeEncodingConfig()
	axelarnetSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), axelarnetSubspace)

	return ctx, keeper
}

func TestKeeper_GetIBCPath(t *testing.T) {
	repeats := 20

	var (
		ctx sdk.Context
		k   keeper.Keeper
	)

	t.Run("should return the registered IBC path when the given asset is registered", testutils.Func(func(t *testing.T) {
		ctx, k = setup()
		path := randomIBCPath()
		chain := randomChain()
		chain.IBCPath = ""
		k.SetCosmosChain(ctx, chain)
		err := k.SetIBCPath(ctx, chain.Name, path)
		assert.NoError(t, err)
		result, ok := k.GetIBCPath(ctx, chain.Name)
		assert.Equal(t, path, result)
		assert.True(t, ok)
	}).Repeat(repeats))

	t.Run("should return error when registered the same asset twice", testutils.Func(func(t *testing.T) {
		ctx, k = setup()
		path := randomIBCPath()
		chain := randomChain()
		chain.IBCPath = ""
		k.SetCosmosChain(ctx, chain)
		err := k.SetIBCPath(ctx, chain.Name, path)
		assert.NoError(t, err)
		path2 := randomIBCPath()
		err2 := k.SetIBCPath(ctx, chain.Name, path2)
		assert.Error(t, err2)
	}).Repeat(repeats))

}

func TestKeeper_RegisterCosmosChain(t *testing.T) {
	repeats := 20

	var (
		ctx sdk.Context
		k   keeper.Keeper
	)

	t.Run("should return list of registered cosmos chains", testutils.Func(func(t *testing.T) {
		ctx, k = setup()

		count := rand.I64Between(10, 100)
		chains := make([]string, count)

		for i := 0; i < int(count); i++ {
			chains[i] = strings.ToLower(rand.NormalizedStr(10))
			k.SetCosmosChain(ctx, types.CosmosChain{
				Name:       nexus.ChainName(chains[i]),
				AddrPrefix: rand.NormalizedStr(5),
			})
		}
		sort.Strings(chains)
		assert.Equal(t, chains,
			slices.Map(k.GetCosmosChains(ctx), func(c nexus.ChainName) string { return c.String() }),
		)

	}).Repeat(repeats))

	t.Run("should empty list when no chain registered", testutils.Func(func(t *testing.T) {
		ctx, k = setup()
		empty := make([]nexus.ChainName, 0)

		assert.Equal(t, empty, k.GetCosmosChains(ctx))

	}).Repeat(repeats))

}

func TestSetFailedTransfer(t *testing.T) {
	ctx, k := setup()
	n := int(rand.I64Between(0, 100))
	for i := 0; i < n; i++ {
		k.SetFailedTransfer(ctx, axelartestutils.RandomIBCTransfer())
	}

	for i := 0; i < n; i++ {
		transfer, ok := k.GetFailedTransfer(ctx, nexus.TransferID(i))
		assert.True(t, ok)
		assert.Equal(t, transfer.ID, nexus.TransferID(i))
	}
}

func randomIBCPath() string {
	port := ibctransfertypes.PortID
	identifier := fmt.Sprintf("%s%d", "channel-", rand.I64Between(0, 9999))
	return fmt.Sprintf("%s/%s", port, identifier)
}
