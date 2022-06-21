package keeper_test

import (
	"fmt"
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
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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
			k = keeper.NewKeeper(cfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace)

		}).
		When("the state is initialized from a genesis state",
			func() {
				initialGenesis = types.NewGenesisState(types.DefaultParams(), rand.AccAddr(), randomChains(), randomTransferQueue(cfg.Codec))
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
			}).Run(t, 10)
}

func randomTransfers() []types.IBCTransfer {
	transferCount := rand.I64Between(0, 100)
	var transfers []types.IBCTransfer
	for i := int64(0); i < transferCount; i++ {
		transfers = append(transfers, randomIBCTransfer())
	}
	return transfers
}

func randomIBCTransfer() types.IBCTransfer {
	denom := rand.Strings(5, 20).WithAlphabet([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY")).Next()
	return types.IBCTransfer{
		Sender:    rand.AccAddr(),
		Receiver:  randomNormalizedStr(5, 20),
		Token:     sdk.NewCoin(denom, sdk.NewInt(rand.PosI64())),
		PortID:    randomNormalizedStr(5, 20),
		ChannelID: randomNormalizedStr(5, 20),
		ID:        nexus.TransferID(uint64(rand.PosI64())),
	}
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
	queueName := "ibc_transfer_queue"
	queueLen := rand.I64Between(0, 20)
	keyPrefix := utils.KeyFromStr("transfer")

	for i := int64(0); i < queueLen; i++ {
		transfer := randomIBCTransfer()

		qs.Items[fmt.Sprintf("%s_%d_%s", queueName, rand.PosI64(), transfer.ID.String())] = utils.QueueState_Item{
			Key:   keyPrefix.AppendStr(transfer.ID.String()).AsKey(),
			Value: cdc.MustMarshalLengthPrefixed(&transfer),
		}
	}

	return qs
}
