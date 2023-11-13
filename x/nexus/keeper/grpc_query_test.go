package keeper_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestKeeper_TransfersForChain(t *testing.T) {
	var (
		k               nexusKeeper.Keeper
		axelarnetKeeper types.AxelarnetKeeper
		q               nexusKeeper.Querier
		ctx             sdk.Context
		totalTransfers  int64
		pageRequest     *query.PageRequest
		response        *types.TransfersForChainResponse
	)

	Given("a nexus keeper", func() {
		encCfg := app.MakeEncodingConfig()
		nexusSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
		k = nexusKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("nexus"), nexusSubspace)
		q = nexusKeeper.NewGRPCQuerier(k, axelarnetKeeper)
	}).
		When("a correct context", func() {
			store := fake.NewMultiStore()
			ctx = sdk.NewContext(store, tmproto.Header{}, false, log.TestingLogger())
		}).
		When("the keeper is correctly set up", func() {
			k.SetParams(ctx, types.DefaultParams())
			k.SetChain(ctx, evm.Ethereum)
			k.ActivateChain(ctx, evm.Ethereum)
			k.SetChain(ctx, axelarnet.Axelarnet)
			k.ActivateChain(ctx, axelarnet.Axelarnet)
			funcs.MustNoErr(k.RegisterAsset(ctx, evm.Ethereum, exported.NewAsset(axelarnet.NativeAsset, false), utils.MaxUint, time.Hour))
			funcs.MustNoErr(k.RegisterAsset(ctx, axelarnet.Axelarnet, exported.NewAsset(axelarnet.NativeAsset, true), utils.MaxUint, time.Hour))

			addressValidators := types.NewAddressValidators().
				AddAddressValidator("evm", func(sdk.Context, exported.CrossChainAddress) error {
					return nil
				}).AddAddressValidator("axelarnet", func(sdk.Context, exported.CrossChainAddress) error {
				return nil
			})
			addressValidators.Seal()
			k.SetAddressValidators(addressValidators)

		}).
		When("there are some pending transfers", func() {
			totalTransfers = rand.I64Between(10, 50)
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
				_, err := k.EnqueueForTransfer(ctx, sender, sdk.NewCoin(axelarnet.NativeAsset, sdk.NewInt(rand.PosI64())))
				assert.NoError(t, err)
			}
		}).
		When("pagination flags are set up", func() {
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
		When("TransferForChain is called", func() {
			var err error
			response, err = q.TransfersForChain(sdk.WrapSDKContext(ctx), &types.TransfersForChainRequest{
				Chain:      axelarnet.Axelarnet.Name.String(),
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

func TestKeeper_Chains(t *testing.T) {
	var (
		k               nexusKeeper.Keeper
		axelarnetKeeper types.AxelarnetKeeper
		q               nexusKeeper.Querier
		ctx             sdk.Context
		response        *types.ChainsResponse
		err             error
	)

	testChain := exported.Chain{Name: exported.ChainName("test")}

	Given("a nexus keeper", func() {
		encCfg := app.MakeEncodingConfig()
		nexusSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
		k = nexusKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("nexus"), nexusSubspace)
		q = nexusKeeper.NewGRPCQuerier(k, axelarnetKeeper)
	}).
		When("a correct context", func() {
			store := fake.NewMultiStore()
			ctx = sdk.NewContext(store, tmproto.Header{}, false, log.TestingLogger())
		}).
		When("the keeper is correctly set up", func() {
			k.SetChain(ctx, evm.Ethereum)
			k.ActivateChain(ctx, evm.Ethereum)
			k.SetChain(ctx, axelarnet.Axelarnet)
			k.ActivateChain(ctx, axelarnet.Axelarnet)
			k.SetChain(ctx, testChain)
		}).
		Branch(
			Then("query all chains", func(t *testing.T) {
				response, err = q.Chains(sdk.WrapSDKContext(ctx), &types.ChainsRequest{})
				assert.NoError(t, err)
				assert.Equal(t, response.Chains, []exported.ChainName{axelarnet.Axelarnet.Name, evm.Ethereum.Name, testChain.Name})
			}),
			Then("query only activated chains", func(t *testing.T) {
				response, err = q.Chains(sdk.WrapSDKContext(ctx), &types.ChainsRequest{
					Status: types.Activated,
				})
				assert.NoError(t, err)
				assert.Equal(t, response.Chains, []exported.ChainName{axelarnet.Axelarnet.Name, evm.Ethereum.Name})
			}),
			Then("query only deactivated chains", func(t *testing.T) {
				response, err = q.Chains(sdk.WrapSDKContext(ctx), &types.ChainsRequest{
					Status: types.Deactivated,
				})
				assert.NoError(t, err)
				assert.Equal(t, response.Chains, []exported.ChainName{testChain.Name})
			}),
		).Run(t)
}

func TestKeeper_Message(t *testing.T) {
	var (
		ctx sdk.Context
		k   nexusKeeper.Keeper
		q   nexusKeeper.Querier
		id  string
		msg exported.GeneralMessage
	)

	Given("keeper and context", func() {

		cfg := app.MakeEncodingConfig()
		k, ctx = setup(cfg)
		q = nexusKeeper.NewGRPCQuerier(k, nil)
	}).Branch(
		When("message exists", func() {
			sourceChain := nexustestutils.RandomChain()
			sourceChain.Module = axelarnet.ModuleName
			destinationChain := nexustestutils.RandomChain()
			destinationChain.Module = evmtypes.ModuleName
			k.SetChain(ctx, sourceChain)
			k.SetChain(ctx, destinationChain)
			id, txID, nonce := k.GenerateMessageID(ctx)
			msg := exported.GeneralMessage{
				ID:            id,
				Sender:        exported.CrossChainAddress{Chain: sourceChain, Address: genCosmosAddr(sourceChain.Name.String())},
				Recipient:     exported.CrossChainAddress{Chain: destinationChain, Address: evmtestutils.RandomAddress().Hex()},
				Status:        exported.Processing,
				PayloadHash:   crypto.Keccak256Hash(rand.Bytes(int(rand.I64Between(1, 100)))).Bytes(),
				Asset:         nil,
				SourceTxID:    txID,
				SourceTxIndex: nonce,
			}
			err := k.SetNewMessage(ctx, msg)
			assert.NoError(t, err)
		}).Then("should succeed", func(t *testing.T) {
			response, err := q.Message(sdk.WrapSDKContext(ctx), &types.MessageRequest{ID: id})
			assert.NoError(t, err)
			assert.Equal(t, msg, response.Message)
		}),
		When("message doesn't exist", func() {

		}).Then("should fail", func(t *testing.T) {
			response, err := q.Message(sdk.WrapSDKContext(ctx), &types.MessageRequest{ID: id})
			assert.Error(t, err)
			assert.Nil(t, response)
		}),
	)

}
