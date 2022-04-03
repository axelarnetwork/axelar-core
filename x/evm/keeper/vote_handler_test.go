package keeper_test

import (
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	mathRand "math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestHandleVoteResult(t *testing.T) {
	var (
		ctx     sdk.Context
		basek   *mock.BaseKeeperMock
		chaink  *mock.ChainKeeperMock
		n       *mock.NexusMock
		result  vote.Vote
		handler vote.VoteHandler
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(chain string) types.ChainKeeper {
				if strings.EqualFold(chain, evmChain) {
					return chaink
				}
				return nil
			},
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
		}
		chaink = &mock.ChainKeeperMock{
			GetDepositFunc: func(sdk.Context, common.Hash, common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
				return types.ERC20Deposit{}, 0, false
			},
			GetBurnerInfoFunc: func(sdk.Context, types.Address) *types.BurnerInfo {
				return &types.BurnerInfo{
					TokenAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
					Symbol:       rand.StrBetween(5, 10),
					Asset:        rand.Denom(5, 10),
					Salt:         types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
				}
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) (int64, bool) { return rand.PosI64(), true },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) (uint64, bool) { return mathRand.Uint64(), true },
			GetEventFunc: func(sdk.Context, string) (types.Event, bool) {
				return types.Event{}, false
			},
			SetConfirmedEventFunc: func(sdk.Context, types.Event) error {
				return nil
			},
			GetERC20TokenBySymbolFunc: func(ctx sdk.Context, symbol string) types.ERC20Token {
				return types.NilToken
			},
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
			GetRecipientFunc: func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
				return nexus.CrossChainAddress{}, true
			},
			EnqueueForTransferFunc: func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) (nexus.TransferID, error) {
				return nexus.TransferID(mathRand.Uint64()), nil
			},
		}
		encCfg := params.MakeEncodingConfig()
		handler = keeper.NewVoteHandler(encCfg.Codec, basek, n)

		result = vote.Vote{}
	}

	repeats := 20

	t.Run("GIVEN vote WHEN chain is not registered THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		result.Results = randTransferEvents(int(rand.I64Between(5, 10)))

		err := handler(ctx, rand.Str(5), &result)

		assert.Error(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeDepositConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 0)

	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN chain is not activated THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		n.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return false }

		result.Results = randTransferEvents(int(rand.I64Between(5, 10)))

		err := handler(ctx, evmChain, &result)

		assert.Error(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeDepositConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 0)

	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN result is invalid THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		incorrectResult, _ := codectypes.NewAnyWithValue(types.NewConfirmGatewayTxRequest(rand.AccAddr(), rand.Str(5), types.Hash(common.BytesToHash(rand.Bytes(common.HashLength)))))
		result.Results = []*codectypes.Any{incorrectResult}
		err := handler(ctx, evmChain, &result)

		assert.Error(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeDepositConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 0)

	}).Repeat(repeats))

	t.Run("GIVEN already confirmed event WHEN handle deposit THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		chaink.GetEventFunc = func(sdk.Context, string) (types.Event, bool) {
			return types.Event{}, true
		}

		result.Results = randTransferEvents(int(rand.I64Between(5, 10)))

		err := handler(ctx, evmChain, &result)

		assert.Error(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeDepositConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 0)

	}).Repeat(repeats))

	t.Run("GIVEN transfer event unknown recipient WHEN handler deposit THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		n.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, false
		}

		result.Results = randTransferEvents(int(rand.I64Between(5, 10)))

		err := handler(ctx, evmChain, &result)

		assert.Error(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeDepositConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 0)

	}).Repeat(repeats))

	t.Run("GIVEN transfer event WHEN handle deposit THEN depositConfirmation event is emitted", testutils.Func(func(t *testing.T) {
		setup()
		eventNum := int(rand.I64Between(5, 10))
		result.Results = randTransferEvents(eventNum)
		err := handler(ctx, evmChain, &result)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeDepositConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), eventNum)

	}).Repeat(repeats))

	t.Run("GIVEN tokenDeployed event WHEN token is not exited THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		event := randTokenDeployedEvents(1)
		eventsAny, _ := types.PackEvents(event)
		result.Results = eventsAny

		err := handler(ctx, evmChain, &result)

		assert.Error(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeTokenConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 0)

	}).Repeat(repeats))

	t.Run("GIVEN tokenDeployed event WHEN token address does not match THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		event := randTokenDeployedEvents(1)
		eventsAny, _ := types.PackEvents(event)
		result.Results = eventsAny

		chaink.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Pending})
		}

		err := handler(ctx, evmChain, &result)

		assert.Error(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeTokenConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 0)

	}).Repeat(repeats))

	t.Run("GIVEN tokenDeployed event WHEN handle confirm token THEN tokenConfirmation event is emitted", testutils.Func(func(t *testing.T) {
		setup()
		event := randTokenDeployedEvents(1)
		eventsAny, _ := types.PackEvents(event)
		result.Results = eventsAny
		deployedEvent := event[0].GetEvent().(*types.Event_TokenDeployed)

		chaink.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Pending, TokenAddress: deployedEvent.TokenDeployed.TokenAddress})
		}

		err := handler(ctx, evmChain, &result)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeTokenConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 1)

	}).Repeat(repeats))
}

func randTransferEvents(n int) []*codectypes.Any {
	var events []types.Event
	events = make([]types.Event, n)

	for i := 0; i < n; i++ {
		transfer := types.EventTransfer{
			To:     types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
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

	eventsAny, _ := types.PackEvents(events)
	return eventsAny
}

func randTokenDeployedEvents(n int) []types.Event {
	var events []types.Event
	events = make([]types.Event, n)

	for i := 0; i < n; i++ {
		tokenDeployed := types.EventTokenDeployed{
			TokenAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			Symbol:       rand.Denom(5, 20),
		}
		events[i] = types.Event{
			Chain: evmChain,
			TxId:  types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Index: uint64(rand.I64Between(1, 50)),
			Event: &types.Event_TokenDeployed{
				TokenDeployed: &tokenDeployed,
			},
		}
	}
	return events

}
