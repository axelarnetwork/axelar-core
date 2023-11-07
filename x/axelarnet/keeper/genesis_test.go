package keeper_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestGenesis(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	var (
		k              keeper.Keeper
		ctx            sdk.Context
		initialGenesis *types.GenesisState
	)

	givenKeeper := Given("a keeper",
		func() {
			subspace := paramstypes.NewSubspace(cfg.Codec, cfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "axelarnet")
			k = keeper.NewKeeper(cfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace, &mock.ChannelKeeperMock{}, &mock.FeegrantKeeperMock{})
		})

	givenGenesisState := Given("a genesis state",
		func() {
			ordered := randomTransfers()
			initialGenesis = types.NewGenesisState(types.DefaultParams(), rand.AccAddr(), randomChains(), randomTransferQueue(cfg.Codec, ordered), ordered, randomSeqIDMapping())
			assert.NoError(t, initialGenesis.Validate())

			ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		})

	givenKeeper.
		Given2(givenGenesisState).
		When("duplicate chains are provided", func() {
			chain := testutils.RandomCosmosChain()
			initialGenesis.Chains = append(initialGenesis.Chains, chain, chain)
		}).
		Then("init genesis should panic", func(t *testing.T) {
			assert.Panics(t, func() { k.InitGenesis(ctx, initialGenesis) })
		})

	givenKeeper.
		Given2(givenGenesisState).
		When("duplicate ibc paths are provided", func() {
			chain := testutils.RandomCosmosChain()
			chain2 := testutils.RandomCosmosChain()
			chain2.IBCPath = chain.IBCPath
			initialGenesis.Chains = append(initialGenesis.Chains, chain, chain2)
		}).
		Then("init genesis should panic", func(t *testing.T) {
			assert.Panics(t, func() { k.InitGenesis(ctx, initialGenesis) })
		})

	givenKeeper.
		Given2(givenGenesisState).
		When("genesis state is initialized", func() {
			k.InitGenesis(ctx, initialGenesis)
		}).
		Then("export the identical state",
			func(t *testing.T) {
				exportedGenesis := k.ExportGenesis(ctx)
				assert.NoError(t, exportedGenesis.Validate())

				assert.Equal(t, initialGenesis.CollectorAddress, exportedGenesis.CollectorAddress)
				assert.Equal(t, initialGenesis.Params, exportedGenesis.Params)
				assert.Equal(t, initialGenesis.TransferQueue, exportedGenesis.TransferQueue)
				assert.Equal(t, len(initialGenesis.Chains), len(exportedGenesis.Chains))
				for i := range initialGenesis.Chains {
					assert.Equal(t, initialGenesis.Chains[i].Name, exportedGenesis.Chains[i].Name)
					assert.Equal(t, initialGenesis.Chains[i].IBCPath, exportedGenesis.Chains[i].IBCPath)
					assert.Equal(t, initialGenesis.Chains[i].AddrPrefix, exportedGenesis.Chains[i].AddrPrefix)
				}

				for i := range initialGenesis.IBCTransfers {
					assert.Equal(t, initialGenesis.IBCTransfers[i].ID, exportedGenesis.IBCTransfers[i].ID)
					assert.Equal(t, initialGenesis.IBCTransfers[i].ChannelID, exportedGenesis.IBCTransfers[i].ChannelID)
					assert.Equal(t, initialGenesis.IBCTransfers[i].PortID, exportedGenesis.IBCTransfers[i].PortID)
					assert.Equal(t, initialGenesis.IBCTransfers[i].Receiver, exportedGenesis.IBCTransfers[i].Receiver)
					assert.True(t, initialGenesis.IBCTransfers[i].Token.Equal(exportedGenesis.IBCTransfers[i].Token))
					assert.True(t, initialGenesis.IBCTransfers[i].Sender.Equals(exportedGenesis.IBCTransfers[i].Sender))
				}

				for seqKey, ID := range initialGenesis.SeqIDMapping {
					assert.Equal(t, ID, exportedGenesis.SeqIDMapping[seqKey])
				}
			}).Run(t, 10)
}

func randomTransfers() []types.IBCTransfer {
	transfers := slices.Expand(
		func(_ int) types.IBCTransfer { return testutils.RandomIBCTransfer() },
		int(rand.I64Between(0, 100)),
	)

	sort.SliceStable(transfers, func(i, j int) bool {
		return strings.Compare(transfers[i].ID.String(), transfers[j].ID.String()) < 0
	})

	return transfers
}

func randomChains() []types.CosmosChain {
	chainCount := rand.I64Between(0, 100)
	var chains []types.CosmosChain
	seen := make(map[nexus.ChainName]bool)
	for i := int64(0); i < chainCount; {
		chain := testutils.RandomCosmosChain()
		if seen[chain.Name] {
			continue
		}

		chains = append(chains, chain)
		seen[chain.Name] = true
		i++
	}
	return chains
}

// randomTransferQueue returns a random (valid) transfer queue state for testing
func randomTransferQueue(cdc codec.Codec, transfers []types.IBCTransfer) utils.QueueState {
	qs := utils.QueueState{Items: make(map[string]utils.QueueState_Item)}
	queueName := "route_transfer_queue"
	keyPrefix := utils.KeyFromStr("ibc_transfer")

	for i := 0; i < len(transfers); i++ {
		qs.Items[fmt.Sprintf("%s_%d_%s", queueName, rand.PosI64(), transfers[i].ID.String())] = utils.QueueState_Item{
			Key:   keyPrefix.AppendStr(transfers[i].ID.String()).AsKey(),
			Value: cdc.MustMarshalLengthPrefixed(&transfers[i]),
		}
	}

	return qs
}

func randomSeqIDMapping() map[string]uint64 {
	mapping := make(map[string]uint64)
	seqIDMappingPrefix := key.FromUInt[uint64](3)

	for i := 0; i < 1000; i++ {
		seqKey := seqIDMappingPrefix.
			Append(key.FromStr(ibctypes.PortID)).
			Append(key.FromStr(fmt.Sprintf("channel-%d", rand.PosI64()))).
			Append(key.FromUInt(uint64(rand.PosI64())))

		mapping[string(seqKey.Bytes())] = uint64(rand.PosI64())
	}

	return mapping
}
