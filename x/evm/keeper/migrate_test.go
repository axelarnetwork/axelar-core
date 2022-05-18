package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

const uaxl = "uaxl"

func setup() (sdk.Context, BaseKeeper) {
	encCfg := params.MakeEncodingConfig()
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("evm"), paramsK)

	for _, params := range types.DefaultParams() {
		keeper.ForChain(params.Chain).SetParams(ctx, params)
	}

	return ctx, keeper
}

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx     sdk.Context
		keeper  BaseKeeper
		handler func(ctx sdk.Context) error
	)

	evmChains := []nexus.Chain{exported.Ethereum}
	tokens := []types.ERC20TokenMetadata{
		{
			Asset: "uaxl",
			Details: types.TokenDetails{
				TokenName: rand.NormalizedStr(5),
				Symbol:    rand.NormalizedStr(5),
				Decimals:  8,
				Capacity:  sdk.ZeroInt(),
			},
			Status: types.Pending,
		},

		{
			Asset: rand.NormalizedStr(5),
			Details: types.TokenDetails{
				TokenName: rand.NormalizedStr(5),
				Symbol:    rand.NormalizedStr(5),
				Decimals:  8,
				Capacity:  sdk.ZeroInt(),
			},
			Status: types.Pending,
		},
	}

	pendingChains := []types.PendingChain{
		{
			Chain: nexus.Chain{
				Name:   "evm-1",
				Module: types.ModuleName,
			},
			Params: testutils.RandomParams(),
		},

		{
			Chain: nexus.Chain{
				Name:   "evm-2",
				Module: types.ModuleName,
			},
			Params: testutils.RandomParams(),
		},

		{
			Chain: nexus.Chain{
				Name:   "evm-3",
				Module: types.ModuleName,
			},
			Params: testutils.RandomParams(),
		},
	}

	whenTokensAreSetup := Given("the migration handler", func() {
		ctx, keeper = setup()
		nexus := mock.NexusMock{
			GetChainsFunc: func(_ sdk.Context) []nexus.Chain {
				return evmChains
			},
		}
		handler = GetMigrationHandler(keeper, &nexus)
	}).
		When("tokens are setup for evm chains", func() {
			for _, chain := range evmChains {
				for _, token := range tokens {
					keeper.ForChain(chain.Name).(chainKeeper).setTokenMetadata(ctx, token)
				}
			}
		})

	whenTokensAreSetup.
		When("voting grace period is not set", func() {
			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				subspace, _ := ck.getSubspace(ctx)
				subspace.Set(ctx, types.KeyVotingGracePeriod, int64(0))
			}
		}).
		Then("should set new param voting grace period", func(t *testing.T) {
			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				assert.EqualValues(t, 0, ck.GetParams(ctx).VotingGracePeriod)
			}

			err := handler(ctx)
			assert.Error(t, err)

			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				assert.EqualValues(t, types.DefaultParams()[0].VotingGracePeriod, ck.GetParams(ctx).VotingGracePeriod)
			}
		}).
		Run(t)

	whenTokensAreSetup.
		Then("should delete uaxl token", func(t *testing.T) {
			err := handler(ctx)
			assert.Error(t, err)

			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)

				for _, token := range tokens {
					switch token.Asset {
					case uaxl:
						_, ok := ck.getTokenMetadataByAsset(ctx, token.Asset)
						assert.False(t, ok)
						_, ok = ck.getTokenMetadataBySymbol(ctx, token.Details.Symbol)
						assert.False(t, ok)
					default:
						_, ok := ck.getTokenMetadataByAsset(ctx, token.Asset)
						assert.True(t, ok)
						_, ok = ck.getTokenMetadataBySymbol(ctx, token.Details.Symbol)
						assert.True(t, ok)
					}
				}
			}
		}).
		Run(t)

	chainToCommandIDs := make(map[nexus.ChainName][]types.CommandID)
	whenTokensAreSetup.
		When("token deployment commands are set", func() {
			for _, chain := range evmChains {
				chainToCommandIDs[chain.Name] = make([]types.CommandID, len(tokens))

				for i, meta := range tokens {
					ck := keeper.ForChain(chain.Name).(chainKeeper)

					meta.ChainID, _ = ck.GetChainID(ctx)
					token := types.CreateERC20Token(func(_ types.ERC20TokenMetadata) {}, meta)

					command, err := token.CreateDeployCommand(tssTestUtils.RandKeyID())
					command.ID = types.NewCommandID([]byte(meta.Details.Symbol), meta.ChainID)
					if err != nil {
						panic(err)
					}

					if err := ck.EnqueueCommand(ctx, command); err != nil {
						panic(err)
					}

					chainToCommandIDs[chain.Name][i] = command.ID

					_, ok := ck.GetCommand(ctx, command.ID)
					assert.True(t, ok)
				}
			}
		}).
		Then("should delete uaxl token deployment command", func(t *testing.T) {
			err := handler(ctx)
			assert.NoError(t, err)

			for chain, commandIDs := range chainToCommandIDs {
				for i, commandID := range commandIDs {
					_, ok := keeper.ForChain(chain).GetCommand(ctx, commandID)

					switch tokens[i].Asset {
					case uaxl:
						assert.False(t, ok)
					default:
						assert.True(t, ok)
					}
				}
			}
		}).
		Run(t)

	chainToNonUaxlBurnerCount := make(map[nexus.ChainName]uint)
	whenTokensAreSetup.
		When("some token burners are set", func() {
			for _, chain := range evmChains {
				burnerCount := int(rand.I64Between(5, 1000))
				ck := keeper.ForChain(chain.Name).(chainKeeper)

				for i := 0; i < burnerCount; i++ {
					switch rand.Bools(0.5).Next() {
					case true:
						ck.SetBurnerInfo(ctx, types.BurnerInfo{Asset: uaxl, BurnerAddress: testutils.RandomAddress()})
					default:
						chainToNonUaxlBurnerCount[chain.Name]++
						ck.SetBurnerInfo(ctx, types.BurnerInfo{Asset: rand.NormalizedStr(5), BurnerAddress: testutils.RandomAddress()})
					}
				}

				assert.Len(t, ck.getBurnerInfos(ctx), burnerCount)
			}
		}).
		Then("should delete uaxl burner infos", func(t *testing.T) {
			err := handler(ctx)
			assert.Error(t, err)

			for chain, count := range chainToNonUaxlBurnerCount {
				ck := keeper.ForChain(chain).(chainKeeper)
				assert.Len(t, ck.getBurnerInfos(ctx), int(count))
			}
		}).
		Run(t)

	chainToUaxlConfirmedDepositCount := make(map[nexus.ChainName]uint)
	whenTokensAreSetup.
		When("some confirmed uaxl deposits exist", func() {
			for _, chain := range evmChains {
				confirmedDepositCount := int(rand.I64Between(5, 1000))
				ck := keeper.ForChain(chain.Name).(chainKeeper)

				for i := 0; i < confirmedDepositCount; i++ {
					deposit := types.ERC20Deposit{
						TxID:          testutils.RandomHash(),
						BurnerAddress: testutils.RandomAddress(),
					}

					switch rand.Bools(0.5).Next() {
					case true:
						deposit.Asset = uaxl
						chainToUaxlConfirmedDepositCount[chain.Name]++
					default:
						deposit.Asset = rand.NormalizedStr(5)
					}

					ck.SetDeposit(ctx, deposit, types.DepositStatus_Confirmed)
				}

				assert.Len(t, ck.GetConfirmedDeposits(ctx), confirmedDepositCount)
				assert.Len(t, slices.Filter(ck.GetConfirmedDeposits(ctx), func(d types.ERC20Deposit) bool { return d.Asset == uaxl }), int(chainToUaxlConfirmedDepositCount[chain.Name]))
			}
		}).
		Then("should migrate confirmed uaxl deposits to burnt", func(t *testing.T) {
			err := handler(ctx)
			assert.Error(t, err)

			for chain, count := range chainToUaxlConfirmedDepositCount {
				ck := keeper.ForChain(chain).(chainKeeper)

				assert.NotEmpty(t, ck.GetConfirmedDeposits(ctx))
				assert.Len(t, slices.Filter(ck.GetConfirmedDeposits(ctx), func(d types.ERC20Deposit) bool { return d.Asset == uaxl }), 0)
				assert.Len(t, ck.getBurnedDeposits(ctx), int(count))
			}
		}).
		Run(t)

	whenTokensAreSetup.
		When("some pending chains exist", func() {
			for _, pending := range pendingChains {
				setPendingChain(ctx, keeper, pending.Chain, pending.Params)
				assert.True(t, hasPendingChain(ctx, keeper, pending.Chain.Name))
			}
		}).
		Then("should delete all pending chains", func(t *testing.T) {
			err := handler(ctx)
			assert.Error(t, err)

			for _, pending := range pendingChains {
				assert.False(t, hasPendingChain(ctx, keeper, pending.Chain.Name))
			}
		}).
		Run(t)
}

func setPendingChain(ctx sdk.Context, k BaseKeeper, chain nexus.Chain, p types.Params) {
	k.getBaseStore(ctx).Set(pendingChainKey.Append(utils.LowerCaseKey(chain.Name.String())), &types.PendingChain{Chain: chain, Params: p})
}

func hasPendingChain(ctx sdk.Context, k BaseKeeper, chainName nexus.ChainName) bool {
	return k.getBaseStore(ctx).Has(pendingChainKey.Append(utils.LowerCaseKey(chainName.String())))
}
