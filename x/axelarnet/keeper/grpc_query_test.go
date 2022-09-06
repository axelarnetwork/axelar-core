package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	abci "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	. "github.com/axelarnetwork/utils/test"
	"github.com/axelarnetwork/utils/test/rand"
)

func TestQuerier_PendingIBCTransferCount(t *testing.T) {
	var (
		querier                  keeper.Querier
		response                 *types.PendingIBCTransferCountResponse
		expectedChains           []nexus.ChainName
		expectedTransfersByChain map[string]uint32
	)

	Given("existing pending IBC transfers", func() {
		chainCount := int(rand.I64Between(0, 20))
		expectedChains = make([]nexus.ChainName, 0, chainCount)
		randStr := rand.Strings(5, 20).Distinct()
		for i := 0; i < chainCount; i++ {
			expectedChains = append(expectedChains, nexus.ChainName(randStr.Next()))
		}

		expectedTransfersByChain = make(map[string]uint32, len(expectedChains))

		for _, chain := range expectedChains {
			expectedTransfersByChain[chain.String()] = uint32(rand.I64Between(0, 30))
		}
	}).
		When("a querier", func() {
			k := &mock.BaseKeeperMock{GetCosmosChainsFunc: func(sdk.Context) []nexus.ChainName { return expectedChains }}
			n := &mock.NexusMock{GetTransfersForChainPaginatedFunc: func(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState, pageRequest *query.PageRequest) ([]nexus.CrossChainTransfer, *query.PageResponse, error) {
				if pageRequest.CountTotal {
					return []nexus.CrossChainTransfer{}, &query.PageResponse{Total: uint64(expectedTransfersByChain[chain.Name.String()])}, nil
				}

				return []nexus.CrossChainTransfer{}, nil, fmt.Errorf("count total not set in page request")
			},
				GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
					return nexus.Chain{Name: chain}, true
				},
			}
			querier = keeper.NewGRPCQuerier(k, n)
		}).
		When("IBC transfer counts are queried", func() {
			ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
			var err error
			response, err = querier.PendingIBCTransferCount(sdk.WrapSDKContext(ctx), &types.PendingIBCTransferCountRequest{})
			assert.NoError(t, err)
		}).
		Then("return the correct number for each chain", func(t *testing.T) {
			assert.Equal(t, expectedTransfersByChain, response.TransfersByChain)
		}).Run(t, 20)
}
