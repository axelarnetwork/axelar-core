package keeper_test

import (
	"sort"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	axelartestutils "github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

func setup() (sdk.Context, keeper.Keeper, *mock.ChannelKeeperMock, *mock.FeegrantKeeperMock) {
	encCfg := appParams.MakeEncodingConfig()
	axelarnetSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

	channelK := &mock.ChannelKeeperMock{}
	feegrantK := &mock.FeegrantKeeperMock{}

	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), axelarnetSubspace, channelK, feegrantK)
	return ctx, k, channelK, feegrantK
}

func TestKeeper_GetIBCPath(t *testing.T) {
	repeats := 20

	var (
		ctx sdk.Context
		k   keeper.Keeper
	)

	t.Run("should return the registered IBC path when the given asset is registered", testutils.Func(func(t *testing.T) {
		ctx, k, _, _ = setup()
		chain := axelartestutils.RandomCosmosChain()
		funcs.MustNoErr(k.SetCosmosChain(ctx, chain))
		result, ok := k.GetIBCPath(ctx, chain.Name)
		assert.Equal(t, chain.IBCPath, result)
		assert.True(t, ok)
	}).Repeat(repeats))
}

func TestKeeper_RegisterCosmosChain(t *testing.T) {
	repeats := 20

	var (
		ctx sdk.Context
		k   keeper.Keeper
	)

	t.Run("should return list of registered cosmos chains", testutils.Func(func(t *testing.T) {
		ctx, k, _, _ = setup()

		count := rand.I64Between(10, 100)
		chains := make([]string, count)

		for i := 0; i < int(count); i++ {
			chains[i] = strings.ToLower(rand.NormalizedStr(10))
			chain := axelartestutils.RandomCosmosChain()
			chain.Name = nexus.ChainName(chains[i])
			assert.NoError(t, k.SetCosmosChain(ctx, chain))
		}
		sort.Strings(chains)
		assert.Equal(t, chains,
			slices.Map(k.GetCosmosChains(ctx), func(c nexus.ChainName) string { return c.String() }),
		)

	}).Repeat(repeats))

	t.Run("should empty list when no chain registered", testutils.Func(func(t *testing.T) {
		ctx, k, _, _ = setup()
		empty := make([]nexus.ChainName, 0)

		assert.Equal(t, empty, k.GetCosmosChains(ctx))

	}).Repeat(repeats))

}

func TestSeqIDMap(t *testing.T) {
	ctx, k, channelK, _ := setup()

	nextSeq := 1
	channelK.GetNextSequenceSendFunc = func(ctx sdk.Context, portID, channelID string) (uint64, bool) {
		return uint64(nextSeq), true
	}

	n := int(rand.I64Between(0, 100))
	transfers := slices.Expand(func(_ int) types.IBCTransfer {
		return axelartestutils.RandomIBCTransfer()
	}, n)

	for _, t := range transfers {
		funcs.MustNoErr(k.SetSeqIDMapping(ctx, t))
		nextSeq++
	}

	seq := uint64(1)
	for _, transfer := range transfers {
		id, ok := k.GetSeqIDMapping(ctx, transfer.PortID, transfer.ChannelID, seq)
		assert.True(t, ok)
		assert.Equal(t, transfer.ID, id)
		k.DeleteSeqIDMapping(ctx, transfer.PortID, transfer.ChannelID, seq)
		seq++
	}

	seq = uint64(1)
	for _, transfer := range transfers {
		_, ok := k.GetSeqIDMapping(ctx, transfer.PortID, transfer.ChannelID, seq)
		assert.False(t, ok)
		seq++
	}
}

func TestSetTransferStatus(t *testing.T) {
	ctx, k, _, _ := setup()

	pending := axelartestutils.RandomIBCTransfer()
	assert.NoError(t, k.EnqueueIBCTransfer(ctx, pending))
	actual, ok := k.GetTransfer(ctx, pending.ID)
	assert.True(t, ok)
	assert.Equal(t, pending.ChannelID, actual.ChannelID)
	assert.True(t, pending.Token.IsEqual(actual.Token))
	assert.NoError(t, k.SetTransferCompleted(ctx, pending.ID))
	assert.Error(t, k.SetTransferFailed(ctx, pending.ID))

	pending2 := axelartestutils.RandomIBCTransfer()
	assert.NoError(t, k.EnqueueIBCTransfer(ctx, pending2))
	_, ok = k.GetTransfer(ctx, pending2.ID)
	assert.True(t, ok)
	assert.NoError(t, k.SetTransferFailed(ctx, pending2.ID))
	assert.NoError(t, k.SetTransferPending(ctx, pending2.ID))
	assert.True(t, ok)
}

func TestSetChainByIBCPath(t *testing.T) {
	ctx, k, _, _ := setup()
	ibcPaths := []string{
		"",
		"transfer",
		"transfer//channel-1",
		"/channel-1",
		"transfer/",
		"transfer/channel-1/transfer/channel-2",
		"a/b/c",
	}

	for _, ibcPath := range ibcPaths {
		chain := nexustestutils.RandomChainName()
		err := k.SetChainByIBCPath(ctx, ibcPath, chain)
		assert.ErrorContains(t, err, "invalid IBC path")

		_, found := k.GetChainNameByIBCPath(ctx, ibcPath)
		assert.False(t, found)
	}

	chain := nexustestutils.RandomChainName()
	ibcPath := axelartestutils.RandomIBCPath()
	err := k.SetChainByIBCPath(ctx, ibcPath, chain)
	assert.NoError(t, err)

	chainName, found := k.GetChainNameByIBCPath(ctx, ibcPath)
	assert.True(t, found)
	assert.Equal(t, chain, chainName)
}
