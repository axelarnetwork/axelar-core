package keeper_test

import (
	mathRand "math/rand"
	"strings"
	"testing"

	fakeMock "github.com/axelarnetwork/axelar-core/testutils/fake/interfaces/mock"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"

	"github.com/btcsuite/btcd/btcec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/utils/slices"

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
		ctx        sdk.Context
		cacheStore *fakeMock.CacheMultiStoreMock
		basek      *mock.BaseKeeperMock
		chaink     *mock.ChainKeeperMock
		n          *mock.NexusMock
		s       *mock.SignerMock
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
			SetEventCompletedFunc: func(sdk.Context, string) error {
				return nil
			},
			SetFailedEventFunc: func(sdk.Context, types.Event) error {
				return nil
			},
			GetERC20TokenBySymbolFunc: func(ctx sdk.Context, symbol string) types.ERC20Token {
				return types.NilToken
			},
			SetDepositFunc: func(ctx sdk.Context, deposit types.ERC20Deposit, state types.DepositStatus) {},
			LoggerFunc:     func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
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
		s = &mock.SignerMock{}
		encCfg := params.MakeEncodingConfig()
		handler = keeper.NewVoteHandler(encCfg.Codec, basek, n, s)

		result = vote.Vote{}
	}

	repeats := 20

	t.Run("GIVEN vote WHEN chain is not registered THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		result.Results = randTransferEvents(int(rand.I64Between(5, 10)))

		err := handler(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN chain is not activated THEN still confirm the event", testutils.Func(func(t *testing.T) {
		setup()

		n.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return false }
		eventNum := int(rand.I64Between(5, 10))
		result.Results = randTransferEvents(eventNum)

		err := handler(ctx, &result)
		assert.NoError(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 1)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN result is invalid THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		incorrectResult, _ := codectypes.NewAnyWithValue(types.NewConfirmGatewayTxRequest(rand.AccAddr(), rand.Str(5), types.Hash(common.BytesToHash(rand.Bytes(common.HashLength)))))
		result.Results = []*codectypes.Any{incorrectResult}
		err := handler(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)
	}).Repeat(repeats))

	t.Run("GIVEN already confirmed event WHEN handle deposit THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		chaink.GetEventFunc = func(sdk.Context, string) (types.Event, bool) {
			return types.Event{}, true
		}

		result.Results = randTransferEvents(int(rand.I64Between(5, 10)))

		err := handler(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)

	}).Repeat(repeats))

	t.Run("GIVEN transfer event unknown recipient WHEN handler deposit THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		n.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, false
		}

		result.Results = randTransferEvents(int(rand.I64Between(5, 10)))

		err := handler(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)

	}).Repeat(repeats))

	t.Run("GIVEN transfer event WHEN handle deposit THEN depositConfirmation event is emitted", testutils.Func(func(t *testing.T) {
		setup()
		eventNum := int(rand.I64Between(5, 10))
		result.Results = randTransferEvents(eventNum)
		err := handler(ctx, &result)

		assert.NoError(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 1)
	}).Repeat(repeats))

	t.Run("GIVEN tokenDeployed event WHEN token is not exited THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		event := randTokenDeployedEvents(1)
		eventsAny, _ := types.PackEvents(event)
		result.Results = eventsAny

		err := handler(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)
	}).Repeat(repeats))

	t.Run("GIVEN tokenDeployed event WHEN token address does not match THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		event := randTokenDeployedEvents(1)
		eventsAny, _ := types.PackEvents(event)
		result.Results = eventsAny

		chaink.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Pending})
		}

		err := handler(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 0)

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

		err := handler(ctx, &result)

		assert.NoError(t, err)
		assert.Len(t, cacheStore.WriteCalls(), 1)

	}).Repeat(repeats))

}

func TestHandleVoteTransferKey(t *testing.T) {
	var (
		ctx       sdk.Context
		basek     *mock.BaseKeeperMock
		chaink    *mock.ChainKeeperMock
		n         *mock.NexusMock
		s         *mock.SignerMock
		result    vote.Vote
		handler   vote.VoteHandler
		masterKey tss.Key
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())
		masterKey = randomMultisigKey(tss.MasterKey)
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
			GetRevoteLockingPeriodFunc:        func(sdk.Context) (int64, bool) { return rand.PosI64(), true },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) (uint64, bool) { return mathRand.Uint64(), true },
			GetEventFunc: func(sdk.Context, string) (types.Event, bool) {
				return types.Event{}, false
			},
			SetConfirmedEventFunc: func(sdk.Context, types.Event) error {
				return nil
			},
			SetEventCompletedFunc: func(sdk.Context, string) error {
				return nil
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
		}
		s = &mock.SignerMock{
			GetNextKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				return "", true
			},
			GetKeyFunc: func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
				return masterKey, true
			},
			RotateKeyFunc: func(sdk.Context, nexus.Chain, tss.KeyRole) error { return nil },
		}
		encCfg := params.MakeEncodingConfig()
		handler = keeper.NewVoteHandler(encCfg.Codec, basek, n, s)

		result = vote.Vote{}
	}

	repeats := 20

	t.Run("GIVEN MultisigOwnershipTransferred event WHEN next keyID not found THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		event := randTransferKeyEvent(tss.MasterKey)
		eventsAny, _ := types.PackEvents(event)
		result.Results = eventsAny

		s.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.KeyID, bool) {
			return "", false
		}

		err := handler(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeTransferKeyConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 0)

	}).Repeat(repeats))

	t.Run("GIVEN MultisigOwnershipTransferred event WHEN next key not found THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		event := randTransferKeyEvent(tss.MasterKey)
		eventsAny, _ := types.PackEvents(event)
		result.Results = eventsAny

		s.GetKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
			return tss.Key{}, false
		}

		err := handler(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeTransferKeyConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 0)

	}).Repeat(repeats))

	t.Run("GIVEN MultisigOwnershipTransferred event WHEN new owners do not match expected keys THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		event := randTransferKeyEvent(tss.MasterKey)
		eventsAny, _ := types.PackEvents(event)
		result.Results = eventsAny

		err := handler(ctx, &result)

		assert.Error(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeTransferKeyConfirmation
			if !isValidType {
				return false
			}
			return isValidType
		}), 0)

	}).Repeat(repeats))

	t.Run("GIVEN MultisigOwnershipTransferred event When event is confirmed THEN transferKeyConfirmation event is emitted", testutils.Func(func(t *testing.T) {
		setup()
		event := randTransferKeyEvent(tss.MasterKey)

		multisigPubKeys, _ := masterKey.GetMultisigPubKey()
		expectedAddresses := types.KeysToAddresses(multisigPubKeys...)
		threshold := masterKey.GetMultisigKey().Threshold

		newOwners := slices.Map(expectedAddresses, func(addr common.Address) types.Address { return types.Address(addr) })

		ownershipTransferred := types.EventMultisigOwnershipTransferred{
			PreOwners:     event[0].GetEvent().(*types.Event_MultisigOwnershipTransferred).MultisigOwnershipTransferred.PreOwners,
			PrevThreshold: event[0].GetEvent().(*types.Event_MultisigOwnershipTransferred).MultisigOwnershipTransferred.PrevThreshold,
			NewOwners:     newOwners,
			NewThreshold:  sdk.NewUint(uint64(threshold)),
		}
		event[0].Event = &types.Event_MultisigOwnershipTransferred{
			MultisigOwnershipTransferred: &ownershipTransferred,
		}

		eventsAny, _ := types.PackEvents(event)
		result.Results = eventsAny

		err := handler(ctx, &result)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeTransferKeyConfirmation
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

func randTransferKeyEvent(keyRole tss.KeyRole) []types.Event {
	var events []types.Event
	events = make([]types.Event, 1)

	event := types.Event{
		Chain: evmChain,
		TxId:  types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
		Index: uint64(rand.I64Between(1, 50)),
	}

	preAddresses := make([]types.Address, rand.I64Between(10, 50))
	for i := 0; i < len(preAddresses); i++ {
		preAddresses[i] = types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
	}
	newAddresses := make([]types.Address, rand.I64Between(10, 50))
	for i := 0; i < len(newAddresses); i++ {
		newAddresses[i] = types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
	}
	switch keyRole {
	case tss.MasterKey:
		ownershipTransferred := types.EventMultisigOwnershipTransferred{
			PreOwners:     preAddresses,
			PrevThreshold: sdk.NewUint(uint64(rand.I64Between(10, 50))),
			NewOwners:     newAddresses,
			NewThreshold:  sdk.NewUint(uint64(rand.I64Between(10, 50))),
		}
		event.Event = &types.Event_MultisigOwnershipTransferred{
			MultisigOwnershipTransferred: &ownershipTransferred,
		}
	case tss.SecondaryKey:
		operatorshipTransferred := types.EventMultisigOperatorshipTransferred{
			PreOperators:  preAddresses,
			PrevThreshold: sdk.NewUint(uint64(rand.I64Between(10, 50))),
			NewOperators:  newAddresses,
			NewThreshold:  sdk.NewUint(uint64(rand.I64Between(10, 50))),
		}
		event.Event = &types.Event_MultisigOperatorshipTransferred{
			MultisigOperatorshipTransferred: &operatorshipTransferred,
		}
	}
	events[0] = event
	return events
}

func randomMultisigKey(keyRole tss.KeyRole) tss.Key {
	keyNum := rand.I64Between(5, 15)
	var pks [][]byte
	for i := int64(0); i <= keyNum; i++ {
		sk, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		pks = append(pks, sk.PubKey().SerializeCompressed())
	}

	key := tss.Key{
		ID: tssTestUtils.RandKeyID(),
		PublicKey: &tss.Key_MultisigKey_{
			MultisigKey: &tss.Key_MultisigKey{Values: pks, Threshold: keyNum / 2},
		},
		Role: keyRole,
	}

	return key
}
