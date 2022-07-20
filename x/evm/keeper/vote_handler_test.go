package keeper_test

import (
	mathRand "math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
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
		ctx     sdk.Context
		basek   *mock.BaseKeeperMock
		chaink  *mock.ChainKeeperMock
		n       *mock.NexusMock
		r       *mock.RewarderMock
		result  codec.ProtoMarshaler
		handler vote.VoteHandler
	)

	setup := func() {
		ctx = sdk.NewContext(&fakeMock.MultiStoreMock{}, tmproto.Header{}, false, log.TestingLogger())

		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				if chain.Equals(evmChain) {
					return chaink
				}
				return nil
			},
			LoggerFunc:   func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
			HasChainFunc: func(ctx sdk.Context, chain nexus.ChainName) bool { return true },
		}
		chaink = &mock.ChainKeeperMock{
			GetEventFunc: func(sdk.Context, types.EventID) (types.Event, bool) {
				return types.Event{}, false
			},
			SetConfirmedEventFunc: func(sdk.Context, types.Event) error {
				return nil
			},
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
		}

		chains := map[nexus.ChainName]nexus.Chain{
			exported.Ethereum.Name: exported.Ethereum,
		}
		n = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
		}
		r = &mock.RewarderMock{}
		encCfg := params.MakeEncodingConfig()
		handler = keeper.NewVoteHandler(encCfg.Codec, basek, n, r)
	}

	repeats := 20

	t.Run("Given vote When events are not from the same source chain THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		result = &types.VoteEvents{
			Chain:  nexus.ChainName(rand.Str(5)),
			Events: randTransferEvents(int(rand.I64Between(5, 10))),
		}
		err := handler.HandleResult(ctx, result)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("Given vote When events empty THEN should nothing and return nil", testutils.Func(func(t *testing.T) {
		setup()

		result = &types.VoteEvents{
			Chain:  evmChain,
			Events: []types.Event{},
		}
		err := handler.HandleResult(ctx, result)

		assert.NoError(t, err)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN chain is not registered THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		result = &types.VoteEvents{
			Chain:  evmChain,
			Events: randTransferEvents(int(rand.I64Between(5, 10))),
		}
		err := handler.HandleResult(ctx, result)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN chain is not activated THEN still confirm the event", testutils.Func(func(t *testing.T) {
		setup()

		n.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return false }

		result = &types.VoteEvents{
			Chain:  evmChain,
			Events: randTransferEvents(int(rand.I64Between(5, 10))),
		}
		err := handler.HandleResult(ctx, result)

		assert.NoError(t, err)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN result is invalid THEN panic", testutils.Func(func(t *testing.T) {
		setup()

		result = types.NewConfirmGatewayTxRequest(rand.AccAddr(), rand.Str(5), types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))))
		assert.Panics(t, func() {
			handler.HandleResult(ctx, result)
		})
	}).Repeat(repeats))

	t.Run("GIVEN already confirmed event WHEN handle deposit THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		chaink.GetEventFunc = func(sdk.Context, types.EventID) (types.Event, bool) {
			return types.Event{}, true
		}

		result = &types.VoteEvents{
			Chain:  evmChain,
			Events: randTransferEvents(int(rand.I64Between(5, 10))),
		}
		err := handler.HandleResult(ctx, result)

		assert.Error(t, err)
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
			TxID:  types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Index: uint64(rand.I64Between(1, 50)),
			Event: &types.Event_Transfer{
				Transfer: &transfer,
			},
		}
	}

	return events
}
