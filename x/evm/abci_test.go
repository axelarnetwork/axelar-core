package evm

import (
	"errors"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilsMock "github.com/axelarnetwork/axelar-core/utils/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmTestUtils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigMock "github.com/axelarnetwork/axelar-core/x/multisig/exported/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// Default test setup with common mock implementations
type routingTestSetup struct {
	ctx               sdk.Context
	baseKeeper        *mock.BaseKeeperMock
	nexus             *mock.NexusMock
	multisig          *mock.MultisigKeeperMock
	sourceChainKeeper *mock.ChainKeeperMock
	destChainKeeper   *mock.ChainKeeperMock

	sourceChain nexus.ChainName
	destChain   nexus.ChainName
}

func newRoutingTestSetup(t *testing.T) *routingTestSetup {
	sourceChain := nexus.ChainName("source-evm")
	destChain := nexus.ChainName("dest-chain")

	sourceCk := &mock.ChainKeeperMock{
		LoggerFunc:                 func(ctx sdk.Context) log.Logger { return ctx.Logger() },
		GetParamsFunc:              func(ctx sdk.Context) types.Params { return types.Params{EndBlockerLimit: 50} },
		SetEventCompletedFunc:      func(ctx sdk.Context, eventID types.EventID) error { return nil },
		SetEventFailedFunc:         func(ctx sdk.Context, eventID types.EventID) error { return nil },
		GetConfirmedEventQueueFunc: createEventQueueFunc(),                                                                    // empty queue by default
		GetGatewayAddressFunc:      func(ctx sdk.Context) (types.Address, bool) { return evmTestUtils.RandomAddress(), true }, // gateway exists
	}
	destCk := &mock.ChainKeeperMock{
		LoggerFunc:                 func(ctx sdk.Context) log.Logger { return ctx.Logger() },
		GetParamsFunc:              func(ctx sdk.Context) types.Params { return types.Params{EndBlockerLimit: 50} },
		SetEventCompletedFunc:      func(ctx sdk.Context, eventID types.EventID) error { return nil },
		SetEventFailedFunc:         func(ctx sdk.Context, eventID types.EventID) error { return nil },
		GetConfirmedEventQueueFunc: createEventQueueFunc(),                                                                    // empty queue by default
		GetGatewayAddressFunc:      func(ctx sdk.Context) (types.Address, bool) { return evmTestUtils.RandomAddress(), true }, // gateway exists
	}

	bk := &mock.BaseKeeperMock{
		LoggerFunc: func(ctx sdk.Context) log.Logger { return ctx.Logger() },
		ForChainFunc: func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			if chain == sourceChain {
				return sourceCk, nil
			}
			return destCk, nil
		},
	}

	n := &mock.NexusMock{
		GetChainsFunc: func(ctx sdk.Context) []nexus.Chain {
			return []nexus.Chain{
				{Name: sourceChain, Module: types.ModuleName},
				{Name: destChain, Module: types.ModuleName},
			}
		},
		GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			// source-evm and dest-chain are EVM chains, others are non-EVM
			if chain == sourceChain || chain == destChain {
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			}
			return nexus.Chain{Name: chain, Module: "other"}, true
		},
		IsChainActivatedFunc:      func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		SetNewMessageFunc:         func(ctx sdk.Context, msg nexus.GeneralMessage) error { return nil },
		EnqueueRouteMessageFunc:   func(ctx sdk.Context, id string) error { return nil },
		GetProcessingMessagesFunc: func(ctx sdk.Context, chain nexus.ChainName, limit int64) []nexus.GeneralMessage { return nil },
	}

	m := &mock.MultisigKeeperMock{}

	return &routingTestSetup{
		ctx:               sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.NewTestLogger(t)),
		baseKeeper:        bk,
		nexus:             n,
		multisig:          m,
		sourceChainKeeper: sourceCk,
		destChainKeeper:   destCk,
		sourceChain:       sourceChain,
		destChain:         destChain,
	}
}

func (s *routingTestSetup) createContractCallEvent() types.Event {
	return types.Event{
		Chain: s.sourceChain,
		TxID:  evmTestUtils.RandomHash(),
		Index: uint64(rand.PosI64()),
		Event: &types.Event_ContractCall{
			ContractCall: &types.EventContractCall{
				Sender:           evmTestUtils.RandomAddress(),
				DestinationChain: s.destChain,
				ContractAddress:  evmTestUtils.RandomAddress().Hex(),
				PayloadHash:      types.Hash(evmCrypto.Keccak256Hash(rand.Bytes(100))),
			},
		},
	}
}

func (s *routingTestSetup) createContractCallWithTokenEvent() types.Event {
	return types.Event{
		Chain: s.sourceChain,
		TxID:  evmTestUtils.RandomHash(),
		Index: uint64(rand.PosI64()),
		Event: &types.Event_ContractCallWithToken{
			ContractCallWithToken: &types.EventContractCallWithToken{
				Sender:           evmTestUtils.RandomAddress(),
				DestinationChain: s.destChain,
				ContractAddress:  evmTestUtils.RandomAddress().Hex(),
				PayloadHash:      types.Hash(evmCrypto.Keccak256Hash(rand.Bytes(100))),
				Symbol:           "AXL",
				Amount:           math.NewUint(1000),
			},
		},
	}
}

func (s *routingTestSetup) createTokenDeployedEvent(symbol string, tokenAddress types.Address) types.Event {
	return types.Event{
		Chain: s.sourceChain,
		TxID:  evmTestUtils.RandomHash(),
		Index: uint64(rand.PosI64()),
		Event: &types.Event_TokenDeployed{
			TokenDeployed: &types.EventTokenDeployed{
				Symbol:       symbol,
				TokenAddress: tokenAddress,
			},
		},
	}
}

func (s *routingTestSetup) setupConfirmedToken() {
	s.sourceChainKeeper.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
		return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
			Asset:  "uaxl",
			Status: types.Confirmed,
		})
	}
}

// createEventQueueFunc returns a GetConfirmedEventQueueFunc that dequeues the given events
func createEventQueueFunc(events ...types.Event) func(ctx sdk.Context) utils.KVQueue {
	eventIndex := 0
	return func(ctx sdk.Context) utils.KVQueue {
		return &utilsMock.KVQueueMock{
			DequeueFunc: func(value codec.ProtoMarshaler) bool {
				if eventIndex >= len(events) {
					return false
				}
				bz, _ := events[eventIndex].Marshal()
				funcs.MustNoErr(value.Unmarshal(bz))
				eventIndex++
				return true
			},
		}
	}
}

func (s *routingTestSetup) queueEvents(events ...types.Event) {
	s.sourceChainKeeper.GetConfirmedEventQueueFunc = createEventQueueFunc(events...)
}

func (s *routingTestSetup) queueEvent(event types.Event) {
	s.queueEvents(event)
}

func (s *routingTestSetup) createKeyRotationEvent(addresses []types.Address, weights []math.Uint, threshold math.Uint) types.Event {
	return types.Event{
		Chain: s.sourceChain,
		TxID:  evmTestUtils.RandomHash(),
		Index: uint64(rand.PosI64()),
		Event: &types.Event_MultisigOperatorshipTransferred{
			MultisigOperatorshipTransferred: &types.EventMultisigOperatorshipTransferred{
				NewOperators: addresses,
				NewWeights:   weights,
				NewThreshold: threshold,
			},
		},
	}
}

// keyRotationTestData holds matching key and event data for key rotation tests
type keyRotationTestData struct {
	participant sdk.ValAddress
	pubkey      multisig.PublicKey
	address     types.Address
	weight      math.Uint
	threshold   math.Uint
}

// newKeyRotationTestData generates consistent key/event data for key rotation tests
func newKeyRotationTestData() keyRotationTestData {
	pk, _ := evmCrypto.GenerateKey()
	pubkey := evmCrypto.CompressPubkey(&pk.PublicKey)
	address := types.Address(evmCrypto.PubkeyToAddress(pk.PublicKey))

	return keyRotationTestData{
		participant: rand.ValAddr(),
		pubkey:      pubkey,
		address:     address,
		weight:      math.NewUint(100),
		threshold:   math.NewUint(100),
	}
}

func (d keyRotationTestData) createMockKey() *multisigMock.KeyMock {
	return &multisigMock.KeyMock{
		GetParticipantsFunc: func() []sdk.ValAddress {
			return []sdk.ValAddress{d.participant}
		},
		GetPubKeyFunc: func(valAddress sdk.ValAddress) (multisig.PublicKey, bool) {
			return d.pubkey, true
		},
		GetWeightFunc: func(valAddress sdk.ValAddress) math.Uint {
			return d.weight
		},
		GetMinPassingWeightFunc: func() math.Uint {
			return d.threshold
		},
	}
}

func TestProcessConfirmedEvents(t *testing.T) {
	t.Run("ContractCall", func(t *testing.T) {
		t.Run("event to 'Axelarnet' is marked failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			event := s.createContractCallEvent()
			event.GetContractCall().DestinationChain = axelarnet.Axelarnet.Name
			s.queueEvent(event)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
			assert.Len(t, s.nexus.SetNewMessageCalls(), 0)
		})

		t.Run("event to inactive chain is marked failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.queueEvent(s.createContractCallEvent())
			s.nexus.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return false }

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
			assert.Len(t, s.nexus.SetNewMessageCalls(), 0)
		})

		t.Run("event to unregistered chain is marked failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.queueEvent(s.createContractCallEvent())
			s.nexus.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain == s.sourceChain {
					return nexus.Chain{Name: chain, Module: types.ModuleName}, true
				}
				return nexus.Chain{}, false
			}

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
			assert.Len(t, s.nexus.SetNewMessageCalls(), 0, "no message should be created")
		})

		t.Run("SetNewMessage error marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.queueEvent(s.createContractCallEvent())
			s.nexus.SetNewMessageFunc = func(ctx sdk.Context, msg nexus.GeneralMessage) error {
				return errors.New("failed to set message")
			}

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("EnqueueRouteMessage error marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.queueEvent(s.createContractCallEvent())
			s.nexus.EnqueueRouteMessageFunc = func(ctx sdk.Context, id string) error {
				return errors.New("failed to enqueue")
			}

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("event to valid chain creates message and enqueues for routing", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			event := s.createContractCallEvent()
			s.queueEvent(event)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 1, "event should be marked completed")
			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 0, "event should not be marked failed")
			assert.Len(t, s.nexus.SetNewMessageCalls(), 1, "message should be created")
			assert.Len(t, s.nexus.EnqueueRouteMessageCalls(), 1, "message should be enqueued for routing")

			msg := s.nexus.SetNewMessageCalls()[0].M
			e := event.GetContractCall()

			// Verify message content matches event
			assert.Equal(t, string(event.GetID()), msg.ID)
			assert.Equal(t, s.sourceChain, msg.Sender.Chain.Name)
			assert.Equal(t, e.Sender.Hex(), msg.Sender.Address)
			assert.Equal(t, s.destChain, msg.Recipient.Chain.Name)
			assert.Equal(t, e.ContractAddress, msg.Recipient.Address)
			assert.Equal(t, e.PayloadHash.Bytes(), msg.PayloadHash)
			assert.Equal(t, event.TxID.Bytes(), msg.SourceTxID)
			assert.Equal(t, event.Index, msg.SourceTxIndex)
			assert.Nil(t, msg.Asset, "ContractCall should have no asset")

			// Verify EnqueueRouteMessage was called with the message ID
			assert.Equal(t, msg.ID, s.nexus.EnqueueRouteMessageCalls()[0].ID)
		})

		t.Run("routing does not create command", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.queueEvent(s.createContractCallEvent())
			s.destChainKeeper.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error {
				t.Fatal("EnqueueCommand should not be called during routing")
				return nil
			}

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)
		})

		t.Run("routing does not emit ContractCallApproved", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.queueEvent(s.createContractCallEvent())

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			for _, event := range s.ctx.EventManager().Events() {
				assert.NotEqual(t, "axelar.evm.v1beta1.ContractCallApproved", event.Type,
					"ContractCallApproved should not be emitted during routing")
			}
		})
	})

	t.Run("ContractCallWithToken", func(t *testing.T) {
		t.Run("event to 'Axelarnet' is marked failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.setupConfirmedToken()
			event := s.createContractCallWithTokenEvent()
			event.GetContractCallWithToken().DestinationChain = axelarnet.Axelarnet.Name
			s.queueEvent(event)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
			assert.Len(t, s.nexus.SetNewMessageCalls(), 0)
		})

		t.Run("event to inactive chain is marked failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.setupConfirmedToken()
			s.queueEvent(s.createContractCallWithTokenEvent())
			s.nexus.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return false }

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
			assert.Len(t, s.nexus.SetNewMessageCalls(), 0)
		})

		t.Run("event to unregistered chain is marked failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.setupConfirmedToken()
			s.queueEvent(s.createContractCallWithTokenEvent())
			s.nexus.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain == s.sourceChain {
					return nexus.Chain{Name: chain, Module: types.ModuleName}, true
				}
				return nexus.Chain{}, false
			}

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
			assert.Len(t, s.nexus.SetNewMessageCalls(), 0)
		})

		t.Run("SetNewMessage error marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.setupConfirmedToken()
			s.queueEvent(s.createContractCallWithTokenEvent())
			s.nexus.SetNewMessageFunc = func(ctx sdk.Context, msg nexus.GeneralMessage) error {
				return errors.New("failed to set message")
			}

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("EnqueueRouteMessage error marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.setupConfirmedToken()
			s.queueEvent(s.createContractCallWithTokenEvent())
			s.nexus.EnqueueRouteMessageFunc = func(ctx sdk.Context, id string) error {
				return errors.New("failed to enqueue")
			}

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("source token not confirmed marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.sourceChainKeeper.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
					Asset:  "uaxl",
					Status: types.Pending, // Not confirmed
				})
			}
			s.queueEvent(s.createContractCallWithTokenEvent())

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
			assert.Len(t, s.nexus.SetNewMessageCalls(), 0)
		})

		t.Run("event to wasm chain is marked failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.setupConfirmedToken()
			event := s.createContractCallWithTokenEvent()
			event.GetContractCallWithToken().DestinationChain = nexus.ChainName("wasm-chain")
			s.queueEvent(event)

			// Configure wasm chain
			s.nexus.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain == nexus.ChainName("wasm-chain") {
					return nexus.Chain{Name: chain, Module: wasm.ModuleName}, true
				}
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			}

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
			assert.Len(t, s.nexus.SetNewMessageCalls(), 0, "no message should be created for wasm with token")
		})

		t.Run("event to valid chain creates message with asset and enqueues for routing", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.setupConfirmedToken()
			event := s.createContractCallWithTokenEvent()
			s.queueEvent(event)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 1, "event should be marked completed")
			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 0, "event should not be marked failed")
			assert.Len(t, s.nexus.SetNewMessageCalls(), 1, "message should be created")
			assert.Len(t, s.nexus.EnqueueRouteMessageCalls(), 1, "message should be enqueued for routing")

			msg := s.nexus.SetNewMessageCalls()[0].M
			e := event.GetContractCallWithToken()

			// Verify message content matches event
			assert.Equal(t, string(event.GetID()), msg.ID)
			assert.Equal(t, s.sourceChain, msg.Sender.Chain.Name)
			assert.Equal(t, e.Sender.Hex(), msg.Sender.Address)
			assert.Equal(t, s.destChain, msg.Recipient.Chain.Name)
			assert.Equal(t, e.ContractAddress, msg.Recipient.Address)
			assert.Equal(t, e.PayloadHash.Bytes(), msg.PayloadHash)
			assert.Equal(t, event.TxID.Bytes(), msg.SourceTxID)
			assert.Equal(t, event.Index, msg.SourceTxIndex)

			// Verify asset is set correctly
			assert.NotNil(t, msg.Asset, "ContractCallWithToken should have asset")
			assert.Equal(t, "uaxl", msg.Asset.Denom)
			assert.Equal(t, math.NewInt(1000), msg.Asset.Amount)

			// Verify EnqueueRouteMessage was called with the message ID
			assert.Equal(t, msg.ID, s.nexus.EnqueueRouteMessageCalls()[0].ID)
		})
	})

	t.Run("TokenDeployed", func(t *testing.T) {
		t.Run("token does not exist marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.sourceChainKeeper.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
				return types.NilToken
			}
			s.queueEvent(s.createTokenDeployedEvent("AXL", evmTestUtils.RandomAddress()))

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("token address mismatch marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			expectedAddress := evmTestUtils.RandomAddress()
			differentAddress := evmTestUtils.RandomAddress()
			s.sourceChainKeeper.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
					Asset:        "uaxl",
					Status:       types.Pending,
					TokenAddress: expectedAddress,
				})
			}
			s.queueEvent(s.createTokenDeployedEvent("AXL", differentAddress))

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("valid token deployment confirms token and emits event", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			tokenAddress := evmTestUtils.RandomAddress()
			s.sourceChainKeeper.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
					Asset:        "uaxl",
					Status:       types.Pending,
					TokenAddress: tokenAddress,
					Details:      types.TokenDetails{Symbol: "AXL"},
				})
			}
			s.queueEvent(s.createTokenDeployedEvent("AXL", tokenAddress))

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 1, "event should be marked completed")
			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 0)

			// Verify TokenConfirmation SDK event was emitted
			found := false
			for _, event := range s.ctx.EventManager().Events() {
				if event.Type == types.EventTypeTokenConfirmation {
					found = true
					break
				}
			}
			assert.True(t, found, "TokenConfirmation event should be emitted")
		})
	})

	t.Run("KeyRotation", func(t *testing.T) {
		t.Run("next key ID not found marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.multisig.GetNextKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
				return "", false
			}
			s.queueEvent(s.createKeyRotationEvent(nil, nil, math.ZeroUint()))

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("key not found marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.multisig.GetNextKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
				return "next-key-id", true
			}
			s.multisig.GetKeyFunc = func(ctx sdk.Context, keyID multisig.KeyID) (multisig.Key, bool) {
				return nil, false
			}
			s.queueEvent(s.createKeyRotationEvent(nil, nil, math.ZeroUint()))

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("operator count mismatch marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			s.multisig.GetNextKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
				return "next-key-id", true
			}
			// Key expects 2 participants
			s.multisig.GetKeyFunc = func(ctx sdk.Context, keyID multisig.KeyID) (multisig.Key, bool) {
				return &multisigMock.KeyMock{
					GetParticipantsFunc: func() []sdk.ValAddress {
						return []sdk.ValAddress{rand.ValAddr(), rand.ValAddr()}
					},
					GetPubKeyFunc: func(valAddress sdk.ValAddress) (multisig.PublicKey, bool) {
						pk, _ := evmCrypto.GenerateKey()
						return evmCrypto.CompressPubkey(&pk.PublicKey), true
					},
					GetWeightFunc: func(valAddress sdk.ValAddress) math.Uint {
						return math.NewUint(1)
					},
					GetMinPassingWeightFunc: func() math.Uint {
						return math.NewUint(1)
					},
				}, true
			}
			// Event has only 1 operator - count mismatch
			s.queueEvent(s.createKeyRotationEvent(
				[]types.Address{evmTestUtils.RandomAddress()},
				[]math.Uint{math.NewUint(1)},
				math.NewUint(1),
			))

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("RotateKey error marks event failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			keyData := newKeyRotationTestData()

			s.multisig.GetNextKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
				return "next-key-id", true
			}
			s.multisig.GetKeyFunc = func(ctx sdk.Context, keyID multisig.KeyID) (multisig.Key, bool) {
				return keyData.createMockKey(), true
			}
			s.multisig.RotateKeyFunc = func(ctx sdk.Context, chainName nexus.ChainName) error {
				return errors.New("rotate key failed")
			}
			s.queueEvent(s.createKeyRotationEvent(
				[]types.Address{keyData.address},
				[]math.Uint{keyData.weight},
				keyData.threshold,
			))

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("valid key rotation calls RotateKey, emits event, and marks completed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			keyData := newKeyRotationTestData()

			s.multisig.GetNextKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
				return "next-key-id", true
			}
			s.multisig.GetKeyFunc = func(ctx sdk.Context, keyID multisig.KeyID) (multisig.Key, bool) {
				return keyData.createMockKey(), true
			}
			s.multisig.RotateKeyFunc = func(ctx sdk.Context, chainName nexus.ChainName) error {
				return nil
			}
			s.queueEvent(s.createKeyRotationEvent(
				[]types.Address{keyData.address},
				[]math.Uint{keyData.weight},
				keyData.threshold,
			))

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			// Verify event marked completed
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 1, "event should be marked completed")
			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 0)

			// Verify RotateKey was called
			assert.Len(t, s.multisig.RotateKeyCalls(), 1, "RotateKey should be called")
			assert.Equal(t, s.sourceChain, s.multisig.RotateKeyCalls()[0].ChainName)

			// Verify TransferKeyConfirmation SDK event was emitted
			found := false
			for _, event := range s.ctx.EventManager().Events() {
				if event.Type == types.EventTypeTransferKeyConfirmation {
					found = true
					break
				}
			}
			assert.True(t, found, "TransferKeyConfirmation event should be emitted")
		})
	})

	t.Run("Resilience", func(t *testing.T) {
		t.Run("first event fails but second event still succeeds", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			failingEvent := s.createContractCallEvent()
			failingEvent.GetContractCall().DestinationChain = axelarnet.Axelarnet.Name // Will fail
			succeedingEvent := s.createContractCallEvent()                             // Will succeed

			s.queueEvents(failingEvent, succeedingEvent)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			// Both events should be processed
			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "first event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 1, "second event should be marked completed")
			assert.Len(t, s.nexus.SetNewMessageCalls(), 1, "one message should be created")
		})

		t.Run("multiple chains process independently when one has failures", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			chain1 := nexus.ChainName("chain1")
			chain2 := nexus.ChainName("chain2")

			ck1 := &mock.ChainKeeperMock{
				LoggerFunc:            func(ctx sdk.Context) log.Logger { return ctx.Logger() },
				GetParamsFunc:         func(ctx sdk.Context) types.Params { return types.Params{EndBlockerLimit: 50} },
				SetEventCompletedFunc: func(ctx sdk.Context, eventID types.EventID) error { return nil },
				SetEventFailedFunc:    func(ctx sdk.Context, eventID types.EventID) error { return nil },
			}
			ck2 := &mock.ChainKeeperMock{
				LoggerFunc:            func(ctx sdk.Context) log.Logger { return ctx.Logger() },
				GetParamsFunc:         func(ctx sdk.Context) types.Params { return types.Params{EndBlockerLimit: 50} },
				SetEventCompletedFunc: func(ctx sdk.Context, eventID types.EventID) error { return nil },
				SetEventFailedFunc:    func(ctx sdk.Context, eventID types.EventID) error { return nil },
			}

			// Chain1 has failing event, Chain2 has succeeding event
			failingEvent := types.Event{
				Chain: chain1,
				TxID:  evmTestUtils.RandomHash(),
				Index: 0,
				Event: &types.Event_ContractCall{
					ContractCall: &types.EventContractCall{
						Sender:           evmTestUtils.RandomAddress(),
						DestinationChain: axelarnet.Axelarnet.Name, // Will fail
						ContractAddress:  evmTestUtils.RandomAddress().Hex(),
						PayloadHash:      types.Hash(evmCrypto.Keccak256Hash(rand.Bytes(100))),
					},
				},
			}
			succeedingEvent := types.Event{
				Chain: chain2,
				TxID:  evmTestUtils.RandomHash(),
				Index: 0,
				Event: &types.Event_ContractCall{
					ContractCall: &types.EventContractCall{
						Sender:           evmTestUtils.RandomAddress(),
						DestinationChain: s.destChain,
						ContractAddress:  evmTestUtils.RandomAddress().Hex(),
						PayloadHash:      types.Hash(evmCrypto.Keccak256Hash(rand.Bytes(100))),
					},
				},
			}

			ck1.GetConfirmedEventQueueFunc = createEventQueueFunc(failingEvent)
			ck2.GetConfirmedEventQueueFunc = createEventQueueFunc(succeedingEvent)

			s.baseKeeper.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
				switch chain {
				case chain1:
					return ck1, nil
				case chain2:
					return ck2, nil
				default:
					return s.destChainKeeper, nil
				}
			}
			s.nexus.GetChainsFunc = func(ctx sdk.Context) []nexus.Chain {
				return []nexus.Chain{
					{Name: chain1, Module: types.ModuleName},
					{Name: chain2, Module: types.ModuleName},
				}
			}

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			// Chain1's event should fail, Chain2's should succeed
			assert.Len(t, ck1.SetEventFailedCalls(), 1, "chain1 event should be marked failed")
			assert.Len(t, ck1.SetEventCompletedCalls(), 0)
			assert.Len(t, ck2.SetEventFailedCalls(), 0)
			assert.Len(t, ck2.SetEventCompletedCalls(), 1, "chain2 event should be marked completed")
		})
	})

	t.Run("BoundedComputation", func(t *testing.T) {
		t.Run("only processes up to EndBlockerLimit events and leaves rest for next block", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			limit := int64(3)
			totalEvents := 5
			s.sourceChainKeeper.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: limit}
			}

			// Create more events than the limit
			events := make([]types.Event, totalEvents)
			for i := 0; i < totalEvents; i++ {
				events[i] = s.createContractCallEvent()
			}

			// Track how many events have been dequeued across blocks
			eventIndex := 0
			s.sourceChainKeeper.GetConfirmedEventQueueFunc = func(ctx sdk.Context) utils.KVQueue {
				return &utilsMock.KVQueueMock{
					DequeueFunc: func(value codec.ProtoMarshaler) bool {
						if eventIndex >= len(events) {
							return false
						}
						bz, _ := events[eventIndex].Marshal()
						funcs.MustNoErr(value.Unmarshal(bz))
						eventIndex++
						return true
					},
				}
			}

			// First block: process up to limit
			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			firstBlockProcessed := len(s.sourceChainKeeper.SetEventCompletedCalls()) + len(s.sourceChainKeeper.SetEventFailedCalls())
			assert.Equal(t, int(limit), firstBlockProcessed, "first block should process up to EndBlockerLimit events")
			assert.Equal(t, int(limit), eventIndex, "should only dequeue up to limit events")

			// Second block: process remaining events
			_, err = EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			secondBlockProcessed := len(s.sourceChainKeeper.SetEventCompletedCalls()) + len(s.sourceChainKeeper.SetEventFailedCalls()) - firstBlockProcessed
			assert.Equal(t, totalEvents-int(limit), secondBlockProcessed, "second block should process remaining events")
			assert.Equal(t, totalEvents, eventIndex, "all events should be dequeued after two blocks")
		})
	})

	t.Run("UnsupportedEventTypes", func(t *testing.T) {
		t.Run("Event_TokenSent is marked failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			event := types.Event{
				Chain: s.sourceChain,
				TxID:  evmTestUtils.RandomHash(),
				Index: 0,
				Event: &types.Event_TokenSent{
					TokenSent: &types.EventTokenSent{
						Sender:             evmTestUtils.RandomAddress(),
						DestinationChain:   s.destChain,
						DestinationAddress: "destination",
						Symbol:             "AXL",
						Amount:             math.NewUint(1000),
					},
				},
			}
			s.queueEvent(event)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "unsupported event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})

		t.Run("Event_Transfer is marked failed", func(t *testing.T) {
			s := newRoutingTestSetup(t)
			event := types.Event{
				Chain: s.sourceChain,
				TxID:  evmTestUtils.RandomHash(),
				Index: 0,
				Event: &types.Event_Transfer{
					Transfer: &types.EventTransfer{
						To:     evmTestUtils.RandomAddress(),
						Amount: math.NewUint(1000),
					},
				},
			}
			s.queueEvent(event)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.sourceChainKeeper.SetEventFailedCalls(), 1, "unsupported event should be marked failed")
			assert.Len(t, s.sourceChainKeeper.SetEventCompletedCalls(), 0)
		})
	})
}

// Delivery test setup - extends routing setup for message delivery tests
type deliveryTestSetup struct {
	*routingTestSetup
}

func newDeliveryTestSetup(t *testing.T) *deliveryTestSetup {
	s := newRoutingTestSetup(t)

	// Configure destChainKeeper for delivery-specific operations
	s.destChainKeeper.GetGatewayAddressFunc = func(ctx sdk.Context) (types.Address, bool) {
		return evmTestUtils.RandomAddress(), true
	}
	s.destChainKeeper.GetChainIDFunc = func(ctx sdk.Context) (math.Int, bool) {
		return math.NewInt(1), true
	}
	s.destChainKeeper.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error {
		return nil
	}
	// Default: token confirmed on destination chain
	s.destChainKeeper.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
		return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
			Asset:  asset,
			Status: types.Confirmed,
			Details: types.TokenDetails{
				Symbol: "AXL",
			},
		})
	}

	// Configure nexus for delivery
	s.nexus.SetMessageExecutedFunc = func(ctx sdk.Context, id string) error { return nil }
	s.nexus.SetMessageFailedFunc = func(ctx sdk.Context, id string) error { return nil }

	// Configure multisig for delivery
	s.multisig.GetCurrentKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
		return "current-key-id", true
	}

	return &deliveryTestSetup{routingTestSetup: s}
}

func (s *deliveryTestSetup) createGeneralMessage() nexus.GeneralMessage {
	return nexus.GeneralMessage{
		ID: "msg-" + rand.HexStr(8),
		Sender: nexus.CrossChainAddress{
			Chain:   nexus.Chain{Name: s.sourceChain, Module: types.ModuleName},
			Address: evmTestUtils.RandomAddress().Hex(),
		},
		Recipient: nexus.CrossChainAddress{
			Chain:   nexus.Chain{Name: s.destChain, Module: types.ModuleName},
			Address: evmTestUtils.RandomAddress().Hex(),
		},
		PayloadHash:   rand.Bytes(32),
		SourceTxID:    rand.Bytes(32),
		SourceTxIndex: uint64(rand.PosI64()),
		Status:        nexus.Processing,
	}
}

func (s *deliveryTestSetup) createGeneralMessageWithToken() nexus.GeneralMessage {
	msg := s.createGeneralMessage()
	msg.Asset = &sdk.Coin{Denom: "uaxl", Amount: math.NewInt(1000)}
	return msg
}

func (s *deliveryTestSetup) queueMessages(msgs ...nexus.GeneralMessage) {
	s.nexus.GetProcessingMessagesFunc = func(ctx sdk.Context, chain nexus.ChainName, limit int64) []nexus.GeneralMessage {
		// Only return messages destined for this chain
		var result []nexus.GeneralMessage
		for _, msg := range msgs {
			if msg.Recipient.Chain.Name == chain {
				result = append(result, msg)
			}
		}
		return result
	}
}

func TestDeliverPendingMessages(t *testing.T) {
	t.Run("DeliverMessage", func(t *testing.T) {
		t.Run("gateway not set marks message failed", func(t *testing.T) {
			s := newDeliveryTestSetup(t)
			s.destChainKeeper.GetGatewayAddressFunc = func(ctx sdk.Context) (types.Address, bool) {
				return types.Address{}, false
			}
			msg := s.createGeneralMessage()
			s.queueMessages(msg)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.nexus.SetMessageFailedCalls(), 1, "message should be marked failed")
			assert.Equal(t, msg.ID, s.nexus.SetMessageFailedCalls()[0].ID)
			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 0)
		})

		t.Run("current key not found marks message failed", func(t *testing.T) {
			s := newDeliveryTestSetup(t)
			s.multisig.GetCurrentKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
				return "", false
			}
			msg := s.createGeneralMessage()
			s.queueMessages(msg)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.nexus.SetMessageFailedCalls(), 1, "message should be marked failed")
			assert.Equal(t, msg.ID, s.nexus.SetMessageFailedCalls()[0].ID)
			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 0)
		})

		t.Run("destination chain deactivated marks message failed", func(t *testing.T) {
			s := newDeliveryTestSetup(t)
			s.nexus.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool {
				return chain.Name != s.destChain // dest-chain is deactivated
			}
			msg := s.createGeneralMessage()
			s.queueMessages(msg)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.nexus.SetMessageFailedCalls(), 1, "message should be marked failed")
			assert.Equal(t, msg.ID, s.nexus.SetMessageFailedCalls()[0].ID)
			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 0)
		})

		t.Run("invalid contract address marks message failed", func(t *testing.T) {
			s := newDeliveryTestSetup(t)
			msg := s.createGeneralMessage()
			msg.Recipient.Address = "not-a-valid-hex-address"
			s.queueMessages(msg)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.nexus.SetMessageFailedCalls(), 1, "message should be marked failed")
			assert.Equal(t, msg.ID, s.nexus.SetMessageFailedCalls()[0].ID)
			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 0)

			// Verify ContractCallFailed event emitted (consistent with EVMEventFailed for routing)
			foundEvent := false
			for _, event := range s.ctx.EventManager().Events() {
				if event.Type == "axelar.evm.v1beta1.ContractCallFailed" {
					foundEvent = true
					break
				}
			}
			assert.True(t, foundEvent, "ContractCallFailed event should be emitted on delivery failure")
		})

		t.Run("valid message creates command and marks executed", func(t *testing.T) {
			s := newDeliveryTestSetup(t)

			// Use specific values for verification
			keyID := multisig.KeyID("test-key-id")
			s.multisig.GetCurrentKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
				return keyID, true
			}

			msg := s.createGeneralMessage()
			s.queueMessages(msg)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			// Command should be enqueued on destination chain
			assert.Len(t, s.destChainKeeper.EnqueueCommandCalls(), 1, "command should be enqueued")
			cmd := s.destChainKeeper.EnqueueCommandCalls()[0].Cmd

			// Verify command type and key
			assert.Equal(t, types.COMMAND_TYPE_APPROVE_CONTRACT_CALL, cmd.Type)
			assert.Equal(t, keyID, cmd.KeyID)

			// Verify command params contain message data
			sourceChain, sourceAddress, contractAddress, payloadHash, sourceTxID, sourceEventIndex := types.DecodeApproveContractCallParams(cmd.Params)
			assert.Equal(t, string(msg.GetSourceChain()), sourceChain)
			assert.Equal(t, msg.GetSourceAddress(), sourceAddress)
			assert.Equal(t, msg.GetDestinationAddress(), contractAddress.Hex())
			assert.Equal(t, msg.PayloadHash, payloadHash.Bytes())
			assert.Equal(t, msg.SourceTxID, sourceTxID.Bytes())
			assert.Equal(t, msg.SourceTxIndex, sourceEventIndex.Uint64())

			// Verify ContractCallApproved event emitted
			foundEvent := false
			for _, event := range s.ctx.EventManager().Events() {
				if event.Type == "axelar.evm.v1beta1.ContractCallApproved" {
					foundEvent = true
					break
				}
			}
			assert.True(t, foundEvent, "ContractCallApproved event should be emitted")

			// Message should be marked executed
			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 1)
			assert.Equal(t, msg.ID, s.nexus.SetMessageExecutedCalls()[0].ID)
			assert.Len(t, s.nexus.SetMessageFailedCalls(), 0)
		})
	})

	t.Run("DeliverMessageWithToken", func(t *testing.T) {
		t.Run("destination token not confirmed marks message failed", func(t *testing.T) {
			s := newDeliveryTestSetup(t)
			s.destChainKeeper.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
					Asset:  asset,
					Status: types.Pending, // not confirmed
				})
			}
			msg := s.createGeneralMessageWithToken()
			s.queueMessages(msg)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			assert.Len(t, s.nexus.SetMessageFailedCalls(), 1, "message should be marked failed")
			assert.Equal(t, msg.ID, s.nexus.SetMessageFailedCalls()[0].ID)
			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 0)
		})

		t.Run("valid message with token creates command and marks executed", func(t *testing.T) {
			s := newDeliveryTestSetup(t)

			keyID := multisig.KeyID("test-key-id")
			s.multisig.GetCurrentKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
				return keyID, true
			}

			msg := s.createGeneralMessageWithToken()
			s.queueMessages(msg)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			// Command should be enqueued on destination chain
			assert.Len(t, s.destChainKeeper.EnqueueCommandCalls(), 1, "command should be enqueued")
			cmd := s.destChainKeeper.EnqueueCommandCalls()[0].Cmd

			// Verify command type and key
			assert.Equal(t, types.COMMAND_TYPE_APPROVE_CONTRACT_CALL_WITH_MINT, cmd.Type)
			assert.Equal(t, keyID, cmd.KeyID)

			// Verify command params contain message data including symbol
			sourceChain, sourceAddress, contractAddress, payloadHash, symbol, amount, sourceTxID, sourceEventIndex := types.DecodeApproveContractCallWithMintParams(cmd.Params)
			assert.Equal(t, string(msg.GetSourceChain()), sourceChain)
			assert.Equal(t, msg.GetSourceAddress(), sourceAddress)
			assert.Equal(t, msg.GetDestinationAddress(), contractAddress.Hex())
			assert.Equal(t, msg.PayloadHash, payloadHash.Bytes())
			assert.Equal(t, "AXL", symbol) // Symbol from destination token
			assert.Equal(t, msg.Asset.Amount.BigInt(), amount)
			assert.Equal(t, msg.SourceTxID, sourceTxID.Bytes())
			assert.Equal(t, msg.SourceTxIndex, sourceEventIndex.Uint64())

			// Verify ContractCallWithMintApproved event emitted
			foundEvent := false
			for _, event := range s.ctx.EventManager().Events() {
				if event.Type == "axelar.evm.v1beta1.ContractCallWithMintApproved" {
					foundEvent = true
					break
				}
			}
			assert.True(t, foundEvent, "ContractCallWithMintApproved event should be emitted")

			// Message should be marked executed
			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 1)
			assert.Equal(t, msg.ID, s.nexus.SetMessageExecutedCalls()[0].ID)
			assert.Len(t, s.nexus.SetMessageFailedCalls(), 0)
		})
	})

	t.Run("Resilience", func(t *testing.T) {
		t.Run("first message fails, second succeeds", func(t *testing.T) {
			s := newDeliveryTestSetup(t)

			failingMsg := s.createGeneralMessage()
			failingMsg.Recipient.Address = "invalid-address" // will fail validation

			successMsg := s.createGeneralMessage()
			s.queueMessages(failingMsg, successMsg)

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			// First message should fail
			assert.Len(t, s.nexus.SetMessageFailedCalls(), 1)
			assert.Equal(t, failingMsg.ID, s.nexus.SetMessageFailedCalls()[0].ID)

			// Second message should succeed
			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 1)
			assert.Equal(t, successMsg.ID, s.nexus.SetMessageExecutedCalls()[0].ID)
		})

		t.Run("multiple chains process messages independently when one has failures", func(t *testing.T) {
			s := newDeliveryTestSetup(t)
			chain1 := nexus.ChainName("chain1")
			chain2 := nexus.ChainName("chain2")

			ck1 := &mock.ChainKeeperMock{
				LoggerFunc:                   func(ctx sdk.Context) log.Logger { return ctx.Logger() },
				GetParamsFunc:                func(ctx sdk.Context) types.Params { return types.Params{EndBlockerLimit: 50} },
				GetConfirmedEventQueueFunc:   createEventQueueFunc(), // empty queue for routing phase
				// Chain1 has no gateway - will fail
				GetGatewayAddressFunc: func(ctx sdk.Context) (types.Address, bool) {
					return types.Address{}, false
				},
			}
			ck2 := &mock.ChainKeeperMock{
				LoggerFunc:                   func(ctx sdk.Context) log.Logger { return ctx.Logger() },
				GetParamsFunc:                func(ctx sdk.Context) types.Params { return types.Params{EndBlockerLimit: 50} },
				GetConfirmedEventQueueFunc:   createEventQueueFunc(), // empty queue for routing phase
				// Chain2 has gateway and can enqueue commands
				GetGatewayAddressFunc: func(ctx sdk.Context) (types.Address, bool) {
					return evmTestUtils.RandomAddress(), true
				},
				GetChainIDFunc: func(ctx sdk.Context) (math.Int, bool) {
					return math.NewInt(2), true
				},
				EnqueueCommandFunc: func(ctx sdk.Context, cmd types.Command) error {
					return nil
				},
			}

			// Create messages for each chain
			chain1Msg := nexus.GeneralMessage{
				ID: "msg-chain1",
				Sender: nexus.CrossChainAddress{
					Chain:   nexus.Chain{Name: s.sourceChain, Module: types.ModuleName},
					Address: evmTestUtils.RandomAddress().Hex(),
				},
				Recipient: nexus.CrossChainAddress{
					Chain:   nexus.Chain{Name: chain1, Module: types.ModuleName},
					Address: evmTestUtils.RandomAddress().Hex(),
				},
				PayloadHash:   rand.Bytes(32),
				SourceTxID:    rand.Bytes(32),
				SourceTxIndex: 0,
				Status:        nexus.Processing,
			}
			chain2Msg := nexus.GeneralMessage{
				ID: "msg-chain2",
				Sender: nexus.CrossChainAddress{
					Chain:   nexus.Chain{Name: s.sourceChain, Module: types.ModuleName},
					Address: evmTestUtils.RandomAddress().Hex(),
				},
				Recipient: nexus.CrossChainAddress{
					Chain:   nexus.Chain{Name: chain2, Module: types.ModuleName},
					Address: evmTestUtils.RandomAddress().Hex(),
				},
				PayloadHash:   rand.Bytes(32),
				SourceTxID:    rand.Bytes(32),
				SourceTxIndex: 0,
				Status:        nexus.Processing,
			}

			s.baseKeeper.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
				switch chain {
				case chain1:
					return ck1, nil
				case chain2:
					return ck2, nil
				default:
					return nil, errors.New("unknown chain")
				}
			}
			s.nexus.GetChainsFunc = func(ctx sdk.Context) []nexus.Chain {
				return []nexus.Chain{
					{Name: chain1, Module: types.ModuleName},
					{Name: chain2, Module: types.ModuleName},
				}
			}
			s.nexus.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool {
				return true
			}
			s.nexus.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			}
			s.nexus.GetProcessingMessagesFunc = func(ctx sdk.Context, chain nexus.ChainName, limit int64) []nexus.GeneralMessage {
				switch chain {
				case chain1:
					return []nexus.GeneralMessage{chain1Msg}
				case chain2:
					return []nexus.GeneralMessage{chain2Msg}
				default:
					return nil
				}
			}

			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			// Chain1's message should fail (no gateway), Chain2's should succeed
			assert.Len(t, s.nexus.SetMessageFailedCalls(), 1, "chain1 message should be marked failed")
			assert.Equal(t, chain1Msg.ID, s.nexus.SetMessageFailedCalls()[0].ID)
			assert.Len(t, s.nexus.SetMessageExecutedCalls(), 1, "chain2 message should be marked executed")
			assert.Equal(t, chain2Msg.ID, s.nexus.SetMessageExecutedCalls()[0].ID)
		})
	})

	t.Run("BoundedComputation", func(t *testing.T) {
		t.Run("only processes up to EndBlockerLimit messages and leaves rest for next block", func(t *testing.T) {
			s := newDeliveryTestSetup(t)

			limit := int64(3)
			totalMessages := 5
			s.destChainKeeper.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: limit}
			}

			// Create more messages than the limit
			var messages []nexus.GeneralMessage
			for i := 0; i < totalMessages; i++ {
				messages = append(messages, s.createGeneralMessage())
			}

			// Track which messages have been processed
			messageIndex := 0
			s.nexus.GetProcessingMessagesFunc = func(ctx sdk.Context, chain nexus.ChainName, requestedLimit int64) []nexus.GeneralMessage {
				if chain != s.destChain {
					return nil
				}
				// Return up to requestedLimit messages starting from messageIndex
				var result []nexus.GeneralMessage
				for i := 0; i < int(requestedLimit) && messageIndex < len(messages); i++ {
					result = append(result, messages[messageIndex])
					messageIndex++
				}
				return result
			}

			// First block: process up to limit
			_, err := EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			firstBlockProcessed := len(s.nexus.SetMessageExecutedCalls()) + len(s.nexus.SetMessageFailedCalls())
			assert.Equal(t, int(limit), firstBlockProcessed, "first block should process up to EndBlockerLimit messages")
			assert.Equal(t, int(limit), messageIndex, "should only fetch up to limit messages")

			// Second block: process remaining messages
			_, err = EndBlocker(s.ctx, s.baseKeeper, s.nexus, s.multisig)
			assert.NoError(t, err)

			secondBlockProcessed := len(s.nexus.SetMessageExecutedCalls()) + len(s.nexus.SetMessageFailedCalls()) - firstBlockProcessed
			assert.Equal(t, totalMessages-int(limit), secondBlockProcessed, "second block should process remaining messages")
			assert.Equal(t, totalMessages, messageIndex, "all messages should be fetched after two blocks")
		})
	})
}
