package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx      sdk.Context
		cdc      codec.Codec
		storeKey = sdk.NewKVStoreKey(types.StoreKey)
		bk       *keeper.BaseKeeper
		chain    nexus.ChainName
		handler  func(ctx sdk.Context) error
	)

	Given("a context", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	}).
		Given("a keeper", func() {
			encCfg := params.MakeEncodingConfig()
			cdc = encCfg.Codec
			chain = types.DefaultParams()[0].Chain
			pk := paramsKeeper.NewKeeper(cdc, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))

			bk = keeper.NewKeeper(cdc, storeKey, pk)
			bk.InitChains(ctx)
			funcs.MustNoErr(bk.CreateChain(ctx, types.DefaultParams()[0]))
		}).
		Given("a migration handler", func() {
			n := &mock.NexusMock{
				GetChainsFunc: func(sdk.Context) []nexus.Chain { return []nexus.Chain{exported.Ethereum} },
			}
			handler = keeper.Migrate8to9(bk, n)
		}).
		Branch(
			When("deposits with no matching transfer events exist", func() {
				// events with no events
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, testutils.RandomDeposit(), types.DepositStatus_Confirmed)
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, testutils.RandomDeposit(), types.DepositStatus_Burned)
				setDeposit(ctx.KVStore(storeKey), cdc, chain, testutils.RandomDeposit(), types.DepositStatus_Confirmed)
				setDeposit(ctx.KVStore(storeKey), cdc, chain, testutils.RandomDeposit(), types.DepositStatus_Burned)
				// events with no transfer events
				deposit, event := randomDepositAndGatewayEvent(chain)
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Confirmed)
				setEvent(ctx, bk, event)

				deposit, event = randomDepositAndGatewayEvent(chain)
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Burned)
				setEvent(ctx, bk, event)

				deposit, event = randomDepositAndGatewayEvent(chain)
				setDeposit(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Confirmed)
				setEvent(ctx, bk, event)

				deposit, event = randomDepositAndGatewayEvent(chain)
				setDeposit(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Burned)
				setEvent(ctx, bk, event)
				// events with no matching transfer events
				deposit = testutils.RandomDeposit()
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Confirmed)
				setEvent(ctx, bk, randomTransferEvent(chain, deposit.TxID))

				deposit = testutils.RandomDeposit()
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Burned)
				setEvent(ctx, bk, randomTransferEvent(chain, deposit.TxID))

				deposit = testutils.RandomDeposit()
				setDeposit(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Confirmed)
				setEvent(ctx, bk, randomTransferEvent(chain, deposit.TxID))

				deposit = testutils.RandomDeposit()
				setDeposit(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Burned)
				setEvent(ctx, bk, randomTransferEvent(chain, deposit.TxID))
			}).
				Then("should ignore them all", func(t *testing.T) {
					assert.NoError(t, handler(ctx))

					actual := getChainState(ctx, bk, chain)
					assert.Len(t, actual.LegacyConfirmedDeposits, 6)
					assert.Len(t, actual.LegacyBurnedDeposits, 6)
					assert.Zero(t, len(actual.ConfirmedDeposits)+len(actual.BurnedDeposits))
				}),

			When("deposits with matching transfer events exist", func() {
				deposit := testutils.RandomDeposit()
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Confirmed)
				setEvent(ctx, bk, randomTransferEvent(chain, deposit.TxID, deposit.BurnerAddress))

				deposit = testutils.RandomDeposit()
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Burned)
				setEvent(ctx, bk, randomTransferEvent(chain, deposit.TxID, deposit.BurnerAddress))

				deposit = testutils.RandomDeposit()
				setDeposit(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Confirmed)
				setEvent(ctx, bk, randomTransferEvent(chain, deposit.TxID, deposit.BurnerAddress))

				deposit = testutils.RandomDeposit()
				setDeposit(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Burned)
				setEvent(ctx, bk, randomTransferEvent(chain, deposit.TxID, deposit.BurnerAddress))
			}).
				Then("should migrate them all", func(t *testing.T) {
					assert.NoError(t, handler(ctx))

					actual := getChainState(ctx, bk, chain)
					assert.Len(t, actual.LegacyConfirmedDeposits, 0)
					assert.Len(t, actual.LegacyBurnedDeposits, 0)
					assert.Len(t, actual.ConfirmedDeposits, 2)
					assert.Len(t, actual.BurnedDeposits, 2)
				}),

			When("deposits with multiple matching transfer events exist", func() {
				deposit := testutils.RandomDeposit()
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Confirmed)
				for i := 0; i < 10; i++ {
					event := randomTransferEvent(chain, deposit.TxID, deposit.BurnerAddress)
					event.Index = uint64(i)
					setEvent(ctx, bk, event)
				}

				txID := testutils.RandomHash()
				for i := 0; i < 20; i++ {
					deposit := testutils.RandomDeposit()
					deposit.TxID = txID
					deposit.LogIndex = uint64(i)
					setDeposit(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Burned)

					event := randomTransferEvent(chain, deposit.TxID, deposit.BurnerAddress)
					event.Index = uint64(i)
					setEvent(ctx, bk, event)
				}

				txID = testutils.RandomHash()
				address := testutils.RandomAddress()
				for i := 0; i < 5; i++ {
					deposit := testutils.RandomDeposit()
					deposit.TxID = txID
					deposit.LogIndex = uint64(i)
					deposit.BurnerAddress = address
					setDeposit(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Burned)

					event := randomTransferEvent(chain, deposit.TxID, deposit.BurnerAddress)
					event.Index = uint64(i)
					setEvent(ctx, bk, event)
				}
			}).
				Then("should migrate them all", func(t *testing.T) {
					assert.NoError(t, handler(ctx))

					actual := getChainState(ctx, bk, chain)
					assert.Len(t, actual.LegacyConfirmedDeposits, 0)
					assert.Len(t, actual.LegacyBurnedDeposits, 0)
					assert.Len(t, actual.ConfirmedDeposits, 10)
					assert.Len(t, actual.BurnedDeposits, 25)
				}),

			When("deposits with multiple matching transfer events exist and are 'burnt/confirmed' at the same time", func() {
				deposit := testutils.RandomDeposit()
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Confirmed)
				setDeposit(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Burned)
				setEvent(ctx, bk, randomTransferEvent(chain, deposit.TxID, deposit.BurnerAddress))

				deposit = testutils.RandomDeposit()
				setDeposit(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Confirmed)
				setDepositLegacy(ctx.KVStore(storeKey), cdc, chain, deposit, types.DepositStatus_Burned)
				setEvent(ctx, bk, randomTransferEvent(chain, deposit.TxID, deposit.BurnerAddress))
			}).
				Then("should migrate them all as burnt", func(t *testing.T) {
					assert.NoError(t, handler(ctx))

					actual := getChainState(ctx, bk, chain)
					assert.Len(t, actual.LegacyConfirmedDeposits, 0)
					assert.Len(t, actual.LegacyBurnedDeposits, 0)
					assert.Len(t, actual.ConfirmedDeposits, 0)
					assert.Len(t, actual.BurnedDeposits, 2)
				}),
		).
		Run(t)
}

func getChainState(ctx sdk.Context, bk *keeper.BaseKeeper, chain nexus.ChainName) types.GenesisState_Chain {
	for _, chainState := range bk.ExportGenesis(ctx).Chains {
		if chainState.Params.Chain == chain {
			return chainState
		}
	}

	panic(fmt.Errorf("chain state %s not found", chain))
}

func setDepositLegacy(store sdk.KVStore, cdc codec.Codec, chain nexus.ChainName, deposit types.ERC20Deposit, status types.DepositStatus) {
	var prefix key.Key

	switch status {
	case types.DepositStatus_Confirmed:
		prefix = key.FromStr(fmt.Sprintf("chain_%s_confirmed_deposit", chain))
	case types.DepositStatus_Burned:
		prefix = key.FromStr(fmt.Sprintf("chain_%s_burned_deposit", chain))
	default:
		panic("invalid deposit status")
	}

	store.Set(
		prefix.Append(key.FromStr(deposit.TxID.Hex())).Append(key.FromStr(deposit.BurnerAddress.Hex())).Bytes(),
		cdc.MustMarshalLengthPrefixed(&deposit),
	)
}

func setDeposit(store sdk.KVStore, cdc codec.Codec, chain nexus.ChainName, deposit types.ERC20Deposit, status types.DepositStatus) {
	var prefix key.Key

	switch status {
	case types.DepositStatus_Confirmed:
		prefix = key.FromStr(fmt.Sprintf("chain_%s_confirmed_deposit", chain))
	case types.DepositStatus_Burned:
		prefix = key.FromStr(fmt.Sprintf("chain_%s_burned_deposit", chain))
	default:
		panic("invalid deposit status")
	}

	store.Set(
		prefix.Append(key.FromStr(deposit.TxID.Hex())).Append(key.FromUInt(deposit.LogIndex)).Bytes(),
		cdc.MustMarshalLengthPrefixed(&deposit),
	)
}

func setEvent(ctx sdk.Context, bk *keeper.BaseKeeper, event types.Event) {
	ck := funcs.Must(bk.ForChain(ctx, event.Chain))
	ck.SetConfirmedEvent(ctx, event)

	switch event.Status {
	case types.EventCompleted:
		ck.SetEventCompleted(ctx, event.GetID())
	case types.EventFailed:
		ck.SetEventFailed(ctx, event.GetID())
	}
}

func randomDepositAndGatewayEvent(chain nexus.ChainName) (types.ERC20Deposit, types.Event) {
	deposit := testutils.RandomDeposit()
	event := testutils.RandomGatewayEvent()
	event.TxID = deposit.TxID
	event.Chain = chain

	return deposit, event
}

func randomTransferEvent(chain nexus.ChainName, txID types.Hash, address ...types.Address) types.Event {
	eventTransfer := testutils.RandomEventTransfer()
	if len(address) > 0 {
		eventTransfer.To = address[0]
	}

	return types.Event{
		Chain:  chain,
		TxID:   txID,
		Index:  uint64(rand.PosI64()),
		Status: rand.Of(types.EventConfirmed, types.EventCompleted, types.EventFailed),
		Event: &types.Event_Transfer{
			Transfer: &eventTransfer,
		},
	}
}
