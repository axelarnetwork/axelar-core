package keeper_test

import (
	"strconv"
	"testing"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	. "github.com/axelarnetwork/utils/test"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestKeeper_TransfersForChain(t *testing.T) {
	var (
		k              nexusKeeper.Keeper
		ctx            sdk.Context
		totalTransfers int64
		pageRequest    *query.PageRequest
		response       *types.TransfersForChainResponse
	)

	Given("a nexus keeper", func(t *testing.T) {
		encCfg := app.MakeEncodingConfig()
		nexusSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
		k = nexusKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("nexus"), nexusSubspace)
	}).And().
		Given("a correct context", func(t *testing.T) {
			store := fake.NewMultiStore()
			ctx = sdk.NewContext(store, tmproto.Header{}, false, log.TestingLogger())
		}).And().
		Given("the keeper is correctly set up", func(t *testing.T) {
			k.SetParams(ctx, types.DefaultParams())
			k.SetChain(ctx, evm.Ethereum)
			k.ActivateChain(ctx, evm.Ethereum)
			k.SetChain(ctx, axelarnet.Axelarnet)
			k.ActivateChain(ctx, axelarnet.Axelarnet)

			nexusRouter := types.NewRouter().
				AddAddressValidator("evm", func(sdk.Context, exported.CrossChainAddress) error {
					return nil
				}).AddAddressValidator("axelarnet", func(sdk.Context, exported.CrossChainAddress) error {
				return nil
			})
			k.SetRouter(nexusRouter)

		}).And().
		Given("there are some pending transfers", func(t *testing.T) {
			totalTransfers = rand.I64Between(10, 200)
			for i := int64(0); i < totalTransfers; i++ {
				sender := exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: rand.Str(20),
				}
				assert.NoError(t,
					k.LinkAddresses(
						ctx,
						sender,
						exported.CrossChainAddress{
							Chain:   axelarnet.Axelarnet,
							Address: rand.AccAddr().String(),
						},
					))
				_, err := k.EnqueueForTransfer(ctx, sender, sdk.NewCoin(axelarnet.Uaxl, sdk.NewInt(rand.PosI64())), sdk.NewDec(0))
				assert.NoError(t, err)
			}
		}).And().
		Given("pagination flags are set up", func(t *testing.T) {
			pageFlags := pflag.NewFlagSet("pagination", pflag.PanicOnError)
			pageFlags.Uint64(flags.FlagPage, 1, "")
			pageFlags.Uint64(flags.FlagLimit, 100, "")

			assert.NoError(t, pageFlags.Set(flags.FlagPage, strconv.FormatInt(rand.I64Between(0, 3), 10)))
			assert.NoError(t, pageFlags.Set(flags.FlagLimit, strconv.FormatInt(rand.I64Between(1, totalTransfers), 10)))
			var err error
			pageRequest, err = client.ReadPageRequest(pageFlags)
			if len(pageRequest.Key) == 0 && pageRequest.Offset > 0 {
				pageRequest.Key = nil
			}

			assert.NoError(t, err)
		}).
		When("TransferForChain is called", func(t *testing.T) {
			var err error
			response, err = k.TransfersForChain(sdk.WrapSDKContext(ctx), &types.TransfersForChainRequest{
				Chain:      axelarnet.Axelarnet.Name,
				State:      exported.Pending,
				Pagination: pageRequest,
			})
			assert.NoError(t, err)

		}).
		Then("return only paginated transfers", func(t *testing.T) {
			count := int(pageRequest.Limit)
			if int(pageRequest.Limit) > int(totalTransfers)-int(pageRequest.Offset) {
				count = int(totalTransfers) - int(pageRequest.Offset)
			}
			assert.Len(t, response.Transfers, count)
		}).Run(t, 20)

}
