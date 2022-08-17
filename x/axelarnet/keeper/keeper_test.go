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
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	axelartestutils "github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

func setup() (sdk.Context, keeper.Keeper, *mock.ChannelKeeperMock) {
	encCfg := appParams.MakeEncodingConfig()
	axelarnetSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

	channelK := &mock.ChannelKeeperMock{}

	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), axelarnetSubspace, channelK)

	return ctx, k, channelK
}

func TestKeeper_GetIBCPath(t *testing.T) {
	repeats := 20

	var (
		ctx sdk.Context
		k   keeper.Keeper
	)

	t.Run("should return the registered IBC path when the given asset is registered", testutils.Func(func(t *testing.T) {
		ctx, k, _ = setup()
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
		ctx, k, _ = setup()
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
		ctx, k, _ = setup()

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
		ctx, k, _ = setup()
		empty := make([]nexus.ChainName, 0)

		assert.Equal(t, empty, k.GetCosmosChains(ctx))

	}).Repeat(repeats))

}

func TestSetFailedTransfer(t *testing.T) {
	ctx, k, _ := setup()
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

func TestSeqIDMap(t *testing.T) {
	ctx, k, channelK := setup()

	nextSeq := 2
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
	ctx, k, _ := setup()

	nonExistent := axelartestutils.RandomIBCTransfer()
	nonExistent.Status = types.TransferNonExistent
	assert.Error(t, k.EnqueueTransfer(ctx, nonExistent))

	_, ok := k.GetTransfer(ctx, nonExistent.ID)
	assert.False(t, ok)

	pending := axelartestutils.RandomIBCTransfer()
	assert.NoError(t, k.EnqueueTransfer(ctx, pending))
	actual, ok := k.GetTransfer(ctx, 1)
	assert.True(t, ok)
	assert.Equal(t, pending.ChannelID, actual.ChannelID)
	assert.True(t, pending.Token.IsEqual(actual.Token))
	assert.NoError(t, k.SetTransferCompleted(ctx, 1))
	assert.Error(t, k.SetTransferFailed(ctx, 1))

	pending2 := axelartestutils.RandomIBCTransfer()
	assert.NoError(t, k.EnqueueTransfer(ctx, pending2))
	_, ok = k.GetTransfer(ctx, 2)
	assert.True(t, ok)
	assert.NoError(t, k.SetTransferFailed(ctx, 2))
	assert.NoError(t, k.SetTransferPending(ctx, 2))
	assert.True(t, ok)
}

func randomIBCPath() string {
	port := ibctransfertypes.PortID
	identifier := fmt.Sprintf("%s%d", "channel-", rand.I64Between(0, 9999))
	return fmt.Sprintf("%s/%s", port, identifier)
}
