package keeper_test

import (
	mathRand "math/rand"
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	fakeMock "github.com/axelarnetwork/axelar-core/testutils/fake/interfaces/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestHandleResult(t *testing.T) {
	var (
		ctx        sdk.Context
		cacheStore *fakeMock.CacheMultiStoreMock
		basek      *mock.BaseKeeperMock
		chaink     *mock.ChainKeeperMock
		n          *mock.NexusMock
		r          *mock.RewarderMock
		result     vote.Vote
		handler    vote.VoteHandler
	)

	setup := func() {
		store := &fakeMock.MultiStoreMock{}
		cacheStore = &fakeMock.CacheMultiStoreMock{
			WriteFunc: func() {},
		}
		store.CacheMultiStoreFunc = func() sdk.CacheMultiStore { return cacheStore }

		ctx = sdk.NewContext(store, tmproto.Header{}, false, log.TestingLogger())

		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(chain string) types.ChainKeeper {
				if strings.EqualFold(chain, evmChain) {
					return chaink
				}
				return nil
			},
			LoggerFunc:   func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
			HasChainFunc: func(ctx sdk.Context, chain string) bool { return true },
		}
		chaink = &mock.ChainKeeperMock{
			GetEventFunc: func(sdk.Context, types.EventID) (types.Event, bool) {
				return types.Event{}, false
			},
			SetConfirmedEventFunc: func(sdk.Context, types.Event) error {
				return nil
			},
			SetFailedEventFunc: func(sdk.Context, types.Event) error {
				return nil
			},
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
		}

		chains := map[string]nexus.Chain{
			exported.Ethereum.Name: exported.Ethereum,
		}
		n = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
		}
		r = &mock.RewarderMock{}
		encCfg := params.MakeEncodingConfig()
		handler = keeper.NewVoteHandler(encCfg.Codec, basek, n, r)

		result = vote.Vote{}
	}

	repeats := 20

	t.Run("Given vote When events are not from the same source chain THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		voteEvents, err := types.PackEvents(rand.Str(5), randTransferEvents(int(rand.I64Between(5, 10))))
		if err != nil {
			panic(err)
		}
		result.Result = voteEvents

		err = handler.HandleResult(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)
	}).Repeat(repeats))

	t.Run("Given vote When events empty THEN should nothing and return nil", testutils.Func(func(t *testing.T) {
		setup()

		voteEvents, err := types.PackEvents(evmChain, []types.Event{})
		if err != nil {
			panic(err)
		}
		result.Result = voteEvents

		err = handler.HandleResult(ctx, &result)

		assert.NoError(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN chain is not registered THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		voteEvents, err := types.PackEvents(evmChain, randTransferEvents(int(rand.I64Between(5, 10))))
		if err != nil {
			panic(err)
		}
		result.Result = voteEvents

		err = handler.HandleResult(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN chain is not activated THEN still confirm the event", testutils.Func(func(t *testing.T) {
		setup()

		n.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return false }
		eventNum := int(rand.I64Between(5, 10))
		voteEvents, err := types.PackEvents(evmChain, randTransferEvents(eventNum))
		if err != nil {
			panic(err)
		}
		result.Result = voteEvents

		err = handler.HandleResult(ctx, &result)

		assert.NoError(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 1)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN result is invalid THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		incorrectResult, _ := codectypes.NewAnyWithValue(types.NewConfirmGatewayTxRequest(rand.AccAddr(), rand.Str(5), types.Hash(common.BytesToHash(rand.Bytes(common.HashLength)))))
		result.Result = incorrectResult

		err := handler.HandleResult(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)
	}).Repeat(repeats))

	t.Run("GIVEN already confirmed event WHEN handle deposit THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		chaink.GetEventFunc = func(sdk.Context, types.EventID) (types.Event, bool) {
			return types.Event{}, true
		}

		voteEvents, err := types.PackEvents(evmChain, randTransferEvents(int(rand.I64Between(5, 10))))
		if err != nil {
			panic(err)
		}
		result.Result = voteEvents

		err = handler.HandleResult(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)

	}).Repeat(repeats))
}

func randTransferEvents(n int) []types.Event {
	events := make([]types.Event, n)
	burnerAddress := types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
	for i := 0; i < n; i++ {
		transfer := types.EventTransfer{
			To:     burnerAddress,
			Amount: sdk.NewUint(mathRand.Uint64()),
		}
		events[i] = types.Event{
			Chain: evmChain,
			TxId:  types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Index: uint64(rand.I64Between(1, 50)),
			Event: &types.Event_Transfer{
				Transfer: &transfer,
			},
		}
	}

	return events
}
