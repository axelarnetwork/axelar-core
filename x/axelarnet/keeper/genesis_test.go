package keeper_test

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
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

	Given("a keeper",
		func() {
			subspace := paramstypes.NewSubspace(cfg.Codec, cfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "axelarnet")
			k = keeper.NewKeeper(cfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace, &mock.ChannelKeeperMock{})

		}).
		When("the state is initialized from a genesis state",
			func() {
				initialGenesis = types.NewGenesisState(types.DefaultParams(), rand.AccAddr(), randomChains(), randomTransferQueue(cfg.Codec), randomTransfers())
				assert.NoError(t, initialGenesis.Validate())

				ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
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

				for i := range initialGenesis.FailedTransfers {
					assert.Equal(t, initialGenesis.FailedTransfers[i].ID, exportedGenesis.FailedTransfers[i].ID)
					assert.Equal(t, initialGenesis.FailedTransfers[i].ChannelID, exportedGenesis.FailedTransfers[i].ChannelID)
					assert.Equal(t, initialGenesis.FailedTransfers[i].PortID, exportedGenesis.FailedTransfers[i].PortID)
					assert.Equal(t, initialGenesis.FailedTransfers[i].Receiver, exportedGenesis.FailedTransfers[i].Receiver)
					assert.True(t, initialGenesis.FailedTransfers[i].Token.Equal(exportedGenesis.FailedTransfers[i].Token))
					assert.True(t, initialGenesis.FailedTransfers[i].Sender.Equals(exportedGenesis.FailedTransfers[i].Sender))
				}
			}).Run(t, 10)
}

func randomTransfers() []types.IBCTransfer {
	transfers := slices.Expand(
		func(_ int) types.IBCTransfer { return testutils.RandomIBCTransfer() },
		int(rand.I64Between(0, 100)),
	)

	sort.SliceStable(transfers, func(i, j int) bool {
		return bytes.Compare(transfers[i].ID.Bytes(), transfers[j].ID.Bytes()) < 0
	})

	return transfers
}

func randomChains() []types.CosmosChain {
	chainCount := rand.I64Between(0, 100)
	var chains []types.CosmosChain
	for i := int64(0); i < chainCount; i++ {
		chains = append(chains, randomChain())
	}
	return chains
}

func randomChain() types.CosmosChain {
	return types.CosmosChain{
		Name:       nexus.ChainName(randomNormalizedStr(5, 20)),
		IBCPath:    randomIBCPath(),
		AddrPrefix: randomNormalizedStr(5, 20),
	}
}

func randomNormalizedStr(min, max int) string {
	return strings.ReplaceAll(utils.NormalizeString(rand.StrBetween(min, max)), utils.DefaultDelimiter, "-")
}

// randomTransferQueue returns a random (valid) transfer queue state for testing
func randomTransferQueue(cdc codec.Codec) utils.QueueState {
	qs := utils.QueueState{Items: make(map[string]utils.QueueState_Item)}
	queueName := "route_transfer_queue"
	queueLen := rand.I64Between(0, 20)
	keyPrefix := utils.KeyFromStr("transfer")

	for i := int64(0); i < queueLen; i++ {
		transfer := testutils.RandomIBCTransfer()

		qs.Items[fmt.Sprintf("%s_%d_%s", queueName, rand.PosI64(), transfer.ID.String())] = utils.QueueState_Item{
			Key:   keyPrefix.AppendStr(transfer.ID.String()).AsKey(),
			Value: cdc.MustMarshalLengthPrefixed(&transfer),
		}
	}

	return qs
}
