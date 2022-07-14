package evm

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmTestUtils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (sdk.Context, *mock.BaseKeeperMock, *mock.NexusMock, *mock.SignerMock, *mock.ChainKeeperMock, *mock.ChainKeeperMock) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

	bk := &mock.BaseKeeperMock{}
	n := &mock.NexusMock{}
	s := &mock.SignerMock{}
	sourceCk := &mock.ChainKeeperMock{}
	destinationCk := &mock.ChainKeeperMock{}

	bk.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
	sourceCk.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
	return ctx, bk, n, s, sourceCk, destinationCk
}

func TestHandleContractCall(t *testing.T) {
	var (
		event types.Event

		ctx           sdk.Context
		bk            *mock.BaseKeeperMock
		n             *mock.NexusMock
		s             *mock.SignerMock
		sourceCk      *mock.ChainKeeperMock
		destinationCk *mock.ChainKeeperMock
	)

	sourceChainName := nexus.ChainName(rand.Str(5))
	destinationChainName := nexus.ChainName(rand.Str(5))
	payload := rand.Bytes(100)

	givenContractCallEvent := Given("a ContractCall event", func() {
		event = types.Event{
			Chain: sourceChainName,
			TxId:  evmTestUtils.RandomHash(),
			Index: uint64(rand.PosI64()),
			Event: &types.Event_ContractCall{
				ContractCall: &types.EventContractCall{
					Sender:           evmTestUtils.RandomAddress(),
					DestinationChain: destinationChainName,
					ContractAddress:  evmTestUtils.RandomAddress().Hex(),
					PayloadHash:      types.Hash(evmCrypto.Keccak256Hash(payload)),
				},
			},
		}
		ctx, bk, n, s, sourceCk, destinationCk = setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
	})

	whenChainsAreRegistered := givenContractCallEvent.
		When("the source chain is registered", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				switch chain {
				case sourceChainName, destinationChainName:
					return nexus.Chain{Name: chain}, true
				default:
					return nexus.Chain{}, false
				}
			}
		})

	panicWith := func(msg string) func(t *testing.T) {
		return func(t *testing.T) {
			assert.PanicsWithError(t, msg, func() {
				handleContractCall(ctx, event, bk, n, s)
			})
		}
	}

	isDestinationChainEvm := func(isEvm bool) func() {
		return func() {
			bk.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return chain == destinationChainName && isEvm }
		}
	}

	isDestinationChainIDSet := func(isSet bool) func() {
		return func() {
			destinationCk.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.ZeroInt(), isSet }
		}
	}

	isCurrentKeySet := func(isSet bool) func() {
		return func() {
			s.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				if !isSet {
					return "", false
				}

				return tssTestUtils.RandKeyID(), true
			}
		}
	}

	enqueueCommandSucceed := func(isSuccessful bool) func() {
		return func() {
			destinationCk.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error {
				if !isSuccessful {
					return fmt.Errorf("enqueue error")
				}

				return nil
			}
		}
	}

	givenContractCallEvent.
		When("the source chain is not registered", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{}, chain != sourceChainName
			}
		}).
		Then("should panic", panicWith(fmt.Sprintf("%s is not a registered chain", sourceChainName))).
		Run(t)

	givenContractCallEvent.
		When("the destination chain is not registered", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{}, chain != destinationChainName
			}
		}).
		Then("should return false", func(t *testing.T) {
			ok := handleContractCall(ctx, event, bk, n, s)
			assert.False(t, ok)
		}).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is not an evm chain", isDestinationChainEvm(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleContractCall(ctx, event, bk, n, s)
			assert.False(t, ok)
		}).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		When("destination chain ID is not set", isDestinationChainIDSet(false)).
		Then("should panic", panicWith(fmt.Sprintf("could not find chain ID for '%s'", destinationChainName))).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		When("destination chain ID is set", isDestinationChainIDSet(true)).
		When("current key is not set", isCurrentKeySet(false)).
		Then("should panic", panicWith(fmt.Sprintf("no key for chain %s found", destinationChainName))).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		When("destination chain ID is set", isDestinationChainIDSet(true)).
		When("current key is set", isCurrentKeySet(true)).
		When("enqueue command fails", enqueueCommandSucceed(false)).
		Then("should panic", panicWith("enqueue error")).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		When("destination chain ID is set", isDestinationChainIDSet(true)).
		When("current key is set", isCurrentKeySet(true)).
		When("enqueue command succeeds", enqueueCommandSucceed(true)).
		Then("should return true", func(t *testing.T) {
			ok := handleContractCall(ctx, event, bk, n, s)
			assert.True(t, ok)
			assert.Len(t, destinationCk.EnqueueCommandCalls(), 1)
		}).
		Run(t)
}

func TestHandleTokenSent(t *testing.T) {
	var (
		event types.Event

		ctx           sdk.Context
		bk            *mock.BaseKeeperMock
		n             *mock.NexusMock
		sourceCk      *mock.ChainKeeperMock
		destinationCk *mock.ChainKeeperMock
	)

	sourceChainName := nexus.ChainName(rand.Str(5))
	destinationChainName := nexus.ChainName(rand.Str(5))

	givenTokenSentEvent := Given("a TokenSent event", func() {
		event = types.Event{
			Chain: sourceChainName,
			TxId:  evmTestUtils.RandomHash(),
			Index: uint64(rand.PosI64()),
			Event: &types.Event_TokenSent{
				TokenSent: &types.EventTokenSent{
					Sender:             evmTestUtils.RandomAddress(),
					DestinationChain:   destinationChainName,
					DestinationAddress: evmTestUtils.RandomAddress().Hex(),
					Symbol:             rand.Denom(3, 5),
					Amount:             sdk.NewUint(uint64(rand.I64Between(1, 10000))),
				},
			},
		}
		ctx, bk, n, _, sourceCk, destinationCk = setup()
	})

	whenChainsAreRegistered := givenTokenSentEvent.
		When("the source chain is registered", func() {
			bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
				switch chain {
				case sourceChainName:
					return sourceCk
				case destinationChainName:
					return destinationCk
				default:
					return nil
				}
			}

			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				switch chain {
				case sourceChainName, destinationChainName:
					return nexus.Chain{Name: chain}, true
				default:
					return nexus.Chain{}, false
				}
			}

			bk.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return true }
			n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		})

	panicWith := func(msg string) func(t *testing.T) {
		return func(t *testing.T) {
			assert.PanicsWithError(t, msg, func() {
				handleTokenSent(ctx, event, bk, n)
			})
		}
	}

	assertFalse := func(t *testing.T) {
		ok := handleTokenSent(ctx, event, bk, n)
		assert.False(t, ok)
	}

	tokenRegisteredOnChain := func(chain **mock.ChainKeeperMock, confirmed bool) func() {
		return func() {
			(*chain).GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
				if confirmed && symbol == event.GetTokenSent().Symbol {
					return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
				}
				return types.NilToken
			}

			(*chain).GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
				if confirmed && asset == event.GetTokenSent().Symbol {
					return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: asset})
				}
				return types.NilToken
			}
		}
	}

	isDestinationChainEvm := func(isEvm bool) func() {
		return func() {
			bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
				switch chain {
				case sourceChainName:
					return sourceCk
				case destinationChainName:
					if isEvm {
						return destinationCk
					}

					return nil
				default:
					return nil
				}
			}
			bk.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return chain == destinationChainName && isEvm }

			if isEvm {
				tokenRegisteredOnChain(&destinationCk, true)()
			}
		}
	}

	whenTokensAreConfirmed := whenChainsAreRegistered.
		When("token is confirmed on the source chain", tokenRegisteredOnChain(&sourceCk, true)).
		When("token is confirmed on the destination chain", tokenRegisteredOnChain(&destinationCk, true))

	givenTokenSentEvent.
		When("source chain is not registered", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{}, chain != sourceChainName
			}
		}).
		Then("should panic", panicWith(fmt.Sprintf("%s is not a registered chain", sourceChainName))).
		Run(t)

	givenTokenSentEvent.
		When("destination chain is not registered", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{}, chain != destinationChainName
			}
		}).
		Then("should return false", assertFalse).
		Run(t)

	whenChainsAreRegistered.
		When("token is not confirmed on the source chain", tokenRegisteredOnChain(&sourceCk, false)).
		Then("should return false", assertFalse).
		Run(t)

	whenChainsAreRegistered.
		When("token is confirmed on the source chain", tokenRegisteredOnChain(&sourceCk, true)).
		When("token is not confirmed on the destination chain", tokenRegisteredOnChain(&destinationCk, false)).
		Then("should return false", assertFalse).
		Run(t)

	whenTokensAreConfirmed.
		When("failed to enqueue the transfer", func() {
			n.EnqueueTransferFunc = func(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error) {
				return 0, fmt.Errorf("err")
			}
		})
	Then("should return false", assertFalse).
		Run(t)

	whenTokensAreConfirmed.
		When("enqueue the transfer", func() {
			n.EnqueueTransferFunc = func(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error) {
				return 0, nil
			}
		}).
		Then("should return true", func(t *testing.T) {
			ok := handleTokenSent(ctx, event, bk, n)
			assert.True(t, ok)
			assert.Len(t, n.EnqueueTransferCalls(), 1)
		}).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is not an evm chain", isDestinationChainEvm(false)).
		When("token is confirmed on the source chain", tokenRegisteredOnChain(&sourceCk, true)).
		When("enqueue the transfer", func() {
			n.EnqueueTransferFunc = func(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error) {
				return 0, nil
			}
		}).
		Then("should return true", func(t *testing.T) {
			ok := handleTokenSent(ctx, event, bk, n)
			assert.True(t, ok)
			assert.Len(t, n.EnqueueTransferCalls(), 1)
		}).
		Run(t)
}

func TestHandleContractCallWithToken(t *testing.T) {
	payload := rand.Bytes(100)
	sourceChainName := nexus.ChainName(rand.Str(5))
	destinationChainName := nexus.ChainName(rand.Str(5))
	event := types.Event{
		Chain: sourceChainName,
		TxId:  evmTestUtils.RandomHash(),
		Index: uint64(rand.PosI64()),
		Event: &types.Event_ContractCallWithToken{
			ContractCallWithToken: &types.EventContractCallWithToken{
				Sender:           evmTestUtils.RandomAddress(),
				DestinationChain: destinationChainName,
				ContractAddress:  evmTestUtils.RandomAddress().Hex(),
				PayloadHash:      types.Hash(evmCrypto.Keccak256Hash(payload)),
				Symbol:           rand.Denom(3, 5),
				Amount:           sdk.NewUint(uint64(rand.I64Between(1, 10000))),
			},
		},
	}

	t.Run("should panic if the source chain is not registered", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{}, chain != sourceChainName
		}

		assert.PanicsWithError(t, fmt.Sprintf("%s is not a registered chain", sourceChainName), func() {
			handleContractCallWithToken(ctx, event, bk, n, s)
		})
	}))

	t.Run("should panic if the destination chain is not registered", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{}, chain != destinationChainName
		}

		ok := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.False(t, ok)
	}))

	t.Run("should return false if the token is not confirmed on the source chain", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			return types.NilToken
		}

		ok := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.False(t, ok)
	}))

	t.Run("should return false if the token is not confirmed on the destination chain", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			if symbol == event.GetContractCallWithToken().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
			}
			return types.NilToken
		}
		destinationCk.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			return types.NilToken
		}

		ok := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.False(t, ok)
	}))

	t.Run("should return false if the contract address is invalid", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			if symbol == event.GetContractCallWithToken().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
			}
			return types.NilToken
		}
		destinationCk.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			if asset == event.GetContractCallWithToken().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: asset})
			}
			return types.NilToken
		}
		contractAddress := event.GetContractCallWithToken().ContractAddress
		event.GetContractCallWithToken().ContractAddress = rand.Str(42)

		ok := handleContractCallWithToken(ctx, event, bk, n, s)
		event.GetContractCallWithToken().ContractAddress = contractAddress
		assert.False(t, ok)
	}))

	t.Run("should panic if the destination chain ID is not found", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()
		fee := sdk.NewCoin(event.GetContractCallWithToken().Symbol, sdk.NewInt(rand.I64Between(1, event.GetContractCallWithToken().Amount.BigInt().Int64())))

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			if symbol == event.GetContractCallWithToken().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
			}
			return types.NilToken
		}
		destinationCk.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			if asset == event.GetContractCallWithToken().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: asset})
			}
			return types.NilToken
		}
		n.ComputeTransferFeeFunc = func(ctx sdk.Context, sourceChain, destinationChain nexus.Chain, asset sdk.Coin) (sdk.Coin, error) {
			return fee, nil
		}
		destinationCk.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.ZeroInt(), false }

		assert.PanicsWithError(t, fmt.Sprintf("could not find chain ID for '%s'", destinationChainName), func() {
			handleContractCallWithToken(ctx, event, bk, n, s)
		})
	}))

	t.Run("should panic if the destination chain does not have the key set", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()
		fee := sdk.NewCoin(event.GetContractCallWithToken().Symbol, sdk.NewInt(rand.I64Between(1, event.GetContractCallWithToken().Amount.BigInt().Int64())))

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			if symbol == event.GetContractCallWithToken().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
			}
			return types.NilToken
		}
		destinationCk.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			if asset == event.GetContractCallWithToken().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: asset})
			}
			return types.NilToken
		}
		n.ComputeTransferFeeFunc = func(ctx sdk.Context, sourceChain, destinationChain nexus.Chain, asset sdk.Coin) (sdk.Coin, error) {
			return fee, nil
		}
		destinationCk.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.NewInt(1), true }
		s.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), false
		}

		assert.PanicsWithError(t, fmt.Sprintf("no key for chain %s found", destinationChainName), func() {
			handleContractCallWithToken(ctx, event, bk, n, s)
		})
	}))

	t.Run("should return true if successfully created the command", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			if symbol == event.GetContractCallWithToken().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
			}
			return types.NilToken
		}
		destinationCk.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			if asset == event.GetContractCallWithToken().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: asset})
			}
			return types.NilToken
		}
		destinationCk.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.NewInt(1), true }
		s.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		}
		destinationCk.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error { return nil }

		ok := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.True(t, ok)
		assert.Len(t, destinationCk.EnqueueCommandCalls(), 1)
	}))
}

func TestHandleConfirmDeposit(t *testing.T) {
	var (
		event types.Event

		ctx      sdk.Context
		bk       *mock.BaseKeeperMock
		n        *mock.NexusMock
		sourceCk *mock.ChainKeeperMock
	)

	sourceChainName := nexus.ChainName(rand.Str(5))

	givenTransferEvent := Given("a Transfer event", func() {
		event = types.Event{
			Chain: sourceChainName,
			TxId:  evmTestUtils.RandomHash(),
			Index: uint64(rand.PosI64()),
			Event: &types.Event_Transfer{
				Transfer: &types.EventTransfer{
					To:     evmTestUtils.RandomAddress(),
					Amount: sdk.NewUint(uint64(rand.I64Between(1, 10000))),
				},
			},
		}
		ctx, bk, n, _, sourceCk, _ = setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			return sourceCk
		}

		sourceCk.SetDepositFunc = func(sdk.Context, types.ERC20Deposit, types.DepositStatus) {}
	})

	burnerInfoFound := func(found bool) func() {
		return func() {
			sourceCk.GetBurnerInfoFunc = func(sdk.Context, types.Address) *types.BurnerInfo {
				if !found {
					return nil
				}
				return &types.BurnerInfo{
					TokenAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
					Symbol:       rand.StrBetween(5, 10),
					Asset:        rand.Denom(5, 10),
					Salt:         types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
				}
			}
		}
	}

	recipientFound := func(found bool) func() {
		return func() {
			n.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
				return nexus.CrossChainAddress{}, found
			}
		}
	}

	enqueueTransferSucceed := func(succeed bool) func() {
		return func() {
			n.EnqueueForTransferFunc = func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) (nexus.TransferID, error) {
				if !succeed {
					return 0, fmt.Errorf("err")
				}
				return 0, nil
			}
		}
	}

	depositFound := func(found bool) func() {
		return func() {
			sourceCk.GetDepositFunc = func(ctx sdk.Context, txID common.Hash, burnerAddr common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
				return types.ERC20Deposit{}, types.DepositStatus_Confirmed, found
			}
		}
	}

	givenTransferEvent.
		When("burner info not found", burnerInfoFound(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleConfirmDeposit(ctx, event, sourceCk, n, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient not found", recipientFound(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleConfirmDeposit(ctx, event, sourceCk, n, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient found", recipientFound(true)).
		When("failed to enqueue the transfer", enqueueTransferSucceed(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleConfirmDeposit(ctx, event, sourceCk, n, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient found", recipientFound(true)).
		When("enqueue the transfer", enqueueTransferSucceed(true)).
		When("deposit does not exist", depositFound(false)).
		Then("should return true", func(t *testing.T) {
			ok := handleConfirmDeposit(ctx, event, sourceCk, n, exported.Ethereum)
			assert.True(t, ok)
			assert.Len(t, n.EnqueueForTransferCalls(), 1)
			assert.Len(t, sourceCk.SetDepositCalls(), 1)
		}).
		Run(t)
}

func TestHandleConfirmToken(t *testing.T) {
	var (
		event types.Event

		ctx      sdk.Context
		bk       *mock.BaseKeeperMock
		sourceCk *mock.ChainKeeperMock
	)

	sourceChainName := nexus.ChainName(rand.Str(5))

	givenTokenDeployedEvent := Given("a TokenDeployed event", func() {
		event = types.Event{
			Chain: sourceChainName,
			TxId:  evmTestUtils.RandomHash(),
			Index: uint64(rand.PosI64()),
			Event: &types.Event_TokenDeployed{
				TokenDeployed: &types.EventTokenDeployed{
					TokenAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
					Symbol:       rand.Denom(5, 20),
				},
			},
		}
		ctx, bk, _, _, sourceCk, _ = setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			return sourceCk
		}

	})

	canGetERC20TokenBySymbol := func(found bool) func() {
		return func() {
			sourceCk.GetERC20TokenBySymbolFunc = func(sdk.Context, string) types.ERC20Token {
				if !found {
					return types.NilToken
				}
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
					Status:       types.Pending,
					TokenAddress: event.GetEvent().(*types.Event_TokenDeployed).TokenDeployed.TokenAddress,
				})
			}
		}
	}

	givenTokenDeployedEvent.
		When("can not find token by symbol", canGetERC20TokenBySymbol(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleTokenDeployed(ctx, event, sourceCk, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenTokenDeployedEvent.
		When("token address in event does not match expected address", func() {
			sourceCk.GetERC20TokenBySymbolFunc = func(sdk.Context, string) types.ERC20Token {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
					Status:       types.Pending,
					TokenAddress: evmTestUtils.RandomAddress(),
				})
			}
		}).
		Then("should return false", func(t *testing.T) {
			ok := handleTokenDeployed(ctx, event, sourceCk, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenTokenDeployedEvent.
		When("token address in event matches expected address", canGetERC20TokenBySymbol(true)).
		Then("should return true", func(t *testing.T) {
			ok := handleTokenDeployed(ctx, event, sourceCk, exported.Ethereum)
			assert.True(t, ok)
		}).
		Run(t)
}

func TestHandleTransferKey(t *testing.T) {
	var (
		event types.Event

		ctx      sdk.Context
		bk       *mock.BaseKeeperMock
		s        *mock.SignerMock
		sourceCk *mock.ChainKeeperMock
	)

	sourceChainName := nexus.ChainName(rand.Str(5))

	givenMultisigTransferKeyEvent := Given("a MultisigTransferKey event", func() {
		event = randTransferKeyEvent(sourceChainName)
		ctx, bk, _, s, sourceCk, _ = setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			return sourceCk
		}

	})

	isCurrentKeySet := func(isSet bool) func() {
		return func() {
			s.GetNextKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				if !isSet {
					return "", false
				}

				return tssTestUtils.RandKeyID(), true
			}
		}
	}

	KeyFound := func(found bool) func() {
		return func() {
			s.GetKeyFunc = func(sdk.Context, tss.KeyID) (tss.Key, bool) {
				if !found {
					return tss.Key{}, false
				}

				return randomMultisigKey(tss.MasterKey), true
			}
		}
	}

	keyMatches := func() {
		masterKey := randomMultisigKey(tss.MasterKey)
		s.GetKeyFunc = func(sdk.Context, tss.KeyID) (tss.Key, bool) {
			return masterKey, true
		}

		multisigPubKeys, _ := masterKey.GetMultisigPubKey()
		expectedAddresses := types.KeysToAddresses(multisigPubKeys...)
		threshold := masterKey.GetMultisigKey().Threshold

		newOwners := slices.Map(expectedAddresses, func(addr common.Address) types.Address { return types.Address(addr) })

		operatorshipTransferred := types.EventMultisigOperatorshipTransferred{
			NewOperators: newOwners,
			NewThreshold: sdk.NewUint(uint64(threshold)),
		}
		event.Event = &types.Event_MultisigOperatorshipTransferred{
			MultisigOperatorshipTransferred: &operatorshipTransferred,
		}

		s.RotateKeyFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) error { return nil }
		s.GetRotationCountFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) int64 { return 0 }

	}

	givenMultisigTransferKeyEvent.
		When("next key id not set", isCurrentKeySet(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleMultisigTransferKey(ctx, event, sourceCk, s, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenMultisigTransferKeyEvent.
		When("next key id is set", isCurrentKeySet(true)).
		When("next key not found", KeyFound(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleMultisigTransferKey(ctx, event, sourceCk, s, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenMultisigTransferKeyEvent.
		When("next key id is set", isCurrentKeySet(true)).
		When("next key is found, but does not match expected", KeyFound(true)).
		Then("should return false", func(t *testing.T) {
			ok := handleMultisigTransferKey(ctx, event, sourceCk, s, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenMultisigTransferKeyEvent.
		When("next key id is set", isCurrentKeySet(true)).
		When("next key is found, matches expected key", keyMatches).
		Then("should return true", func(t *testing.T) {
			ok := handleMultisigTransferKey(ctx, event, sourceCk, s, exported.Ethereum)
			assert.True(t, ok)
			assert.Len(t, s.RotateKeyCalls(), 1)

		}).
		Run(t)
}

func randTransferKeyEvent(chain nexus.ChainName) types.Event {
	event := types.Event{
		Chain: chain,
		TxId:  types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
		Index: uint64(rand.I64Between(1, 50)),
	}

	newAddresses := make([]types.Address, rand.I64Between(10, 50))
	for i := 0; i < len(newAddresses); i++ {
		newAddresses[i] = types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
	}

	operatorshipTransferred := types.EventMultisigOperatorshipTransferred{
		NewOperators: newAddresses,
		NewThreshold: sdk.NewUint(uint64(rand.I64Between(10, 50))),
	}
	event.Event = &types.Event_MultisigOperatorshipTransferred{
		MultisigOperatorshipTransferred: &operatorshipTransferred,
	}

	return event
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
