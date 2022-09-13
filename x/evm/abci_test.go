package evm

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilsMock "github.com/axelarnetwork/axelar-core/utils/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmTestUtils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTestUtils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	multisigTypesTestuilts "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (sdk.Context, *mock.BaseKeeperMock, *mock.NexusMock, *mock.MultisigKeeperMock, *mock.ChainKeeperMock, *mock.ChainKeeperMock) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

	bk := &mock.BaseKeeperMock{}
	n := &mock.NexusMock{}
	multisigKeeper := &mock.MultisigKeeperMock{}
	sourceCk := &mock.ChainKeeperMock{}
	destinationCk := &mock.ChainKeeperMock{}

	bk.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
	sourceCk.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
	return ctx, bk, n, multisigKeeper, sourceCk, destinationCk
}

func TestHandleContractCall(t *testing.T) {
	var (
		event types.Event

		ctx            sdk.Context
		bk             *mock.BaseKeeperMock
		n              *mock.NexusMock
		multisigKeeper *mock.MultisigKeeperMock
		sourceCk       *mock.ChainKeeperMock
		destinationCk  *mock.ChainKeeperMock
	)

	sourceChainName := nexus.ChainName(rand.Str(5))
	destinationChainName := nexus.ChainName(rand.Str(5))
	payload := rand.Bytes(100)

	givenContractCallEvent := Given("a ContractCall event", func() {
		event = types.Event{
			Chain: sourceChainName,
			TxID:  evmTestUtils.RandomHash(),
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
		ctx, bk, n, multisigKeeper, sourceCk, destinationCk = setup()

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
				handleContractCall(ctx, event, bk, n, multisigKeeper)
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
			multisigKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
				if !isSet {
					return "", false
				}

				return multisigTestUtils.KeyID(), true
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
			ok := handleContractCall(ctx, event, bk, n, multisigKeeper)
			assert.False(t, ok)
		}).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is not an evm chain", isDestinationChainEvm(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleContractCall(ctx, event, bk, n, multisigKeeper)
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
			ok := handleContractCall(ctx, event, bk, n, multisigKeeper)
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
			TxID:  evmTestUtils.RandomHash(),
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
		}).
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
		TxID:  evmTestUtils.RandomHash(),
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
		ctx, bk, n, multisigKeeper, sourceCk, destinationCk := setup()
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
		multisigKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), false
		}

		assert.PanicsWithError(t, fmt.Sprintf("no key for chain %s found", destinationChainName), func() {
			handleContractCallWithToken(ctx, event, bk, n, multisigKeeper)
		})
	}))

	t.Run("should return true if successfully created the command", testutils.Func(func(t *testing.T) {
		ctx, bk, n, multisigKeeper, sourceCk, destinationCk := setup()

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
		multisigKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), true
		}
		destinationCk.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error { return nil }

		ok := handleContractCallWithToken(ctx, event, bk, n, multisigKeeper)
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
			TxID:  evmTestUtils.RandomHash(),
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
			sourceCk.GetDepositFunc = func(ctx sdk.Context, txID types.Hash, burnerAddr types.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
				return types.ERC20Deposit{}, types.DepositStatus_Confirmed, found
			}
		}
	}

	givenTransferEvent.
		When("burner info not found", burnerInfoFound(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleConfirmDeposit(ctx, event, sourceCk, n, exported.Ethereum)
			assert.False(t, ok)
			assert.Len(t, n.EnqueueForTransferCalls(), 0)
			assert.Len(t, sourceCk.SetDepositCalls(), 0)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient not found", recipientFound(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleConfirmDeposit(ctx, event, sourceCk, n, exported.Ethereum)
			assert.False(t, ok)
			assert.Len(t, n.EnqueueForTransferCalls(), 0)
			assert.Len(t, sourceCk.SetDepositCalls(), 0)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient found", recipientFound(true)).
		When("deposit exists", depositFound(true)).
		Then("should return false", func(t *testing.T) {
			ok := handleConfirmDeposit(ctx, event, sourceCk, n, exported.Ethereum)
			assert.False(t, ok)
			assert.Len(t, n.EnqueueForTransferCalls(), 0)
			assert.Len(t, sourceCk.SetDepositCalls(), 0)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient found", recipientFound(true)).
		When("deposit does not exist", depositFound(false)).
		When("failed to enqueue the transfer", enqueueTransferSucceed(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleConfirmDeposit(ctx, event, sourceCk, n, exported.Ethereum)
			assert.False(t, ok)
			assert.Len(t, n.EnqueueForTransferCalls(), 1)
			assert.Len(t, sourceCk.SetDepositCalls(), 0)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient found", recipientFound(true)).
		When("deposit does not exist", depositFound(false)).
		When("enqueue the transfer", enqueueTransferSucceed(true)).
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
			TxID:  evmTestUtils.RandomHash(),
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

		ctx            sdk.Context
		bk             *mock.BaseKeeperMock
		multisigKeeper *mock.MultisigKeeperMock
		sourceCk       *mock.ChainKeeperMock
	)

	sourceChainName := nexus.ChainName(rand.Str(5))

	givenMultisigTransferKeyEvent := Given("a MultisigTransferKey event", func() {
		event = randTransferKeyEvent(sourceChainName)
		ctx, bk, _, multisigKeeper, sourceCk, _ = setup()

		bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
			return sourceCk
		}

	})

	isCurrentKeySet := func(isSet bool) func() {
		return func() {
			multisigKeeper.GetNextKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
				if !isSet {
					return "", false
				}

				return multisigTestUtils.KeyID(), true
			}
		}
	}

	KeyFound := func(found bool) func() {
		return func() {
			multisigKeeper.GetKeyFunc = func(ctx sdk.Context, keyID multisig.KeyID) (multisig.Key, bool) {
				if !found {
					return nil, false
				}

				key := multisigTypesTestuilts.Key()

				return &key, true
			}
		}
	}

	keyMatches := func() {
		key := multisigTypesTestuilts.Key()
		multisigKeeper.GetKeyFunc = func(sdk.Context, multisig.KeyID) (multisig.Key, bool) {
			return &key, true
		}
		addressWeights, newThreshold := types.ParseMultisigKey(&key)
		addresses := maps.Keys(addressWeights)
		newOperators := slices.Map(addresses, func(a string) types.Address { return types.Address(common.HexToAddress(a)) })
		newWeights := slices.Map(addresses, func(a string) sdk.Uint { return addressWeights[a] })

		operatorshipTransferred := types.EventMultisigOperatorshipTransferred{
			NewOperators: newOperators,
			NewWeights:   newWeights,
			NewThreshold: newThreshold,
		}
		event.Event = &types.Event_MultisigOperatorshipTransferred{
			MultisigOperatorshipTransferred: &operatorshipTransferred,
		}

		multisigKeeper.RotateKeyFunc = func(ctx sdk.Context, chainName nexus.ChainName) error { return nil }
	}

	givenMultisigTransferKeyEvent.
		When("next key id not set", isCurrentKeySet(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleMultisigTransferKey(ctx, event, sourceCk, multisigKeeper, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenMultisigTransferKeyEvent.
		When("next key id is set", isCurrentKeySet(true)).
		When("next key not found", KeyFound(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleMultisigTransferKey(ctx, event, sourceCk, multisigKeeper, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenMultisigTransferKeyEvent.
		When("next key id is set", isCurrentKeySet(true)).
		When("next key is found, but does not match expected", KeyFound(true)).
		Then("should return false", func(t *testing.T) {
			ok := handleMultisigTransferKey(ctx, event, sourceCk, multisigKeeper, exported.Ethereum)
			assert.False(t, ok)
		}).
		Run(t)

	givenMultisigTransferKeyEvent.
		When("next key id is set", isCurrentKeySet(true)).
		When("next key is found, matches expected key", keyMatches).
		Then("should return true", func(t *testing.T) {
			ok := handleMultisigTransferKey(ctx, event, sourceCk, multisigKeeper, exported.Ethereum)
			assert.True(t, ok)
			assert.Len(t, multisigKeeper.RotateKeyCalls(), 1)

		}).
		Run(t)
}

func TestHandleConfirmedEvent(t *testing.T) {
	var (
		ctx            sdk.Context
		bk             *mock.BaseKeeperMock
		n              *mock.NexusMock
		multisigKeeper *mock.MultisigKeeperMock
		sourceCk       *mock.ChainKeeperMock
		destinationCk  *mock.ChainKeeperMock

		confirmedEventQueue *utilsMock.KVQueueMock
	)

	confirmedEventQueue = &utilsMock.KVQueueMock{}
	sourceChainName := nexus.ChainName(rand.Str(5))
	destinationChainName := nexus.ChainName(rand.Str(5))

	givenConfirmedEventQueue := Given("confirmedEvent queue", func() {
		ctx, bk, n, multisigKeeper, sourceCk, destinationCk = setup()
		confirmedEventQueue = &utilsMock.KVQueueMock{}
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
		bk.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return true }

		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, true
			}
		}
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		n.GetChainsFunc = func(_ sdk.Context) []nexus.Chain {
			return []nexus.Chain{{Name: sourceChainName}}
		}

		destinationCk.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.ZeroInt(), true }
		destinationCk.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error { return nil }
		destinationCk.GetGatewayAddressFunc = func(sdk.Context) (types.Address, bool) {
			return types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))), true
		}

		multisigKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), true
		}
		multisigKeeper.GetNextKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
			return "", false
		}

		sourceCk.SetEventCompletedFunc = func(ctx sdk.Context, eventID types.EventID) error { return nil }
		sourceCk.GetConfirmedEventQueueFunc = func(sdk.Context) utils.KVQueue {
			return confirmedEventQueue
		}
		sourceCk.SetEventFailedFunc = func(sdk.Context, types.EventID) error { return nil }
	})

	givenConfirmedEventQueue.
		When("end blocker limit is set", func() {
			sourceCk.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{
					EndBlockerLimit: 50,
				}
			}
		}).
		When("events in queue exceeds limit", func() {
			confirmedEventQueue.IsEmptyFunc = func() bool { return false }
			event := types.Event{
				Chain: sourceChainName,
				TxID:  evmTestUtils.RandomHash(),
				Index: uint64(rand.PosI64()),
				Event: &types.Event_ContractCall{
					ContractCall: &types.EventContractCall{
						Sender:           evmTestUtils.RandomAddress(),
						DestinationChain: destinationChainName,
						ContractAddress:  evmTestUtils.RandomAddress().Hex(),
						PayloadHash:      types.Hash(evmCrypto.Keccak256Hash(rand.Bytes(100))),
					},
				},
			}
			confirmedEventQueue.DequeueFunc = func(value codec.ProtoMarshaler) bool {
				bz, _ := event.Marshal()
				if err := value.Unmarshal(bz); err != nil {
					panic(err)
				}

				return true
			}
		}).
		Then("should handle limited number of events", func(t *testing.T) {
			err := handleConfirmedEvents(ctx, bk, n, multisigKeeper)
			assert.NoError(t, err)
			assert.Len(t, sourceCk.SetEventCompletedCalls(), int(sourceCk.GetParams(ctx).EndBlockerLimit))
			assert.Len(t, destinationCk.EnqueueCommandCalls(), int(sourceCk.GetParams(ctx).EndBlockerLimit))
		}).
		Run(t)

	withEvents := func(num int) WhenStatement {
		return When(fmt.Sprintf("having %d events", num), func() {
			count := 0
			confirmedEventQueue.IsEmptyFunc = func() bool { return false }
			confirmedEventQueue.DequeueFunc = func(value codec.ProtoMarshaler) bool {
				if count >= num {
					return false
				}
				count++

				event := evmTestUtils.RandomEvent(types.EventConfirmed)
				switch event.GetEvent().(type) {
				case *types.Event_ContractCall:
					e := event.GetEvent().(*types.Event_ContractCall)
					e.ContractCall.DestinationChain = destinationChainName
					event.Event = e
				case *types.Event_ContractCallWithToken:
					e := event.GetEvent().(*types.Event_ContractCallWithToken)
					e.ContractCallWithToken.DestinationChain = destinationChainName
					event.Event = e
				case *types.Event_TokenSent:
					e := event.GetEvent().(*types.Event_TokenSent)
					e.TokenSent.DestinationChain = destinationChainName
					event.Event = e
				}
				bz, _ := event.Marshal()
				if err := value.Unmarshal(bz); err != nil {
					panic(err)
				}

				return true
			}
		})
	}

	eventNums := int(rand.I64Between(1, 10))
	shouldSetEventFailed := Then("should set event failed", func(t *testing.T) {
		err := handleConfirmedEvents(ctx, bk, n, multisigKeeper)
		assert.NoError(t, err)
		assert.Len(t, sourceCk.SetEventFailedCalls(), eventNums)
	})

	givenConfirmedEventQueueWithEvents := givenConfirmedEventQueue.
		When("end blocker limit is set", func() {
			sourceCk.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{
					EndBlockerLimit: 50,
				}
			}
		}).
		When2(withEvents(eventNums))

	givenConfirmedEventQueueWithEvents.
		When("chain is not registered", func() {
			n.GetChainFunc = func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{}, false
			}
		}).
		Then2(shouldSetEventFailed).Run(t)

	givenConfirmedEventQueueWithEvents.
		When("chain is not activated", func() {
			n.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool {
				return false
			}
		}).
		Then2(shouldSetEventFailed).Run(t)

	givenConfirmedEventQueueWithEvents.
		When("gateway is not set", func() {
			destinationCk.GetGatewayAddressFunc = func(sdk.Context) (types.Address, bool) {
				return types.Address{}, false
			}
		}).
		Then2(shouldSetEventFailed).Run(t)
}

func randTransferKeyEvent(chain nexus.ChainName) types.Event {
	event := types.Event{
		Chain: chain,
		TxID:  types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
		Index: uint64(rand.I64Between(1, 50)),
	}

	newAddresses := slices.Expand(func(_ int) types.Address {
		return types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
	}, int(rand.I64Between(10, 50)))

	totalWeight := sdk.ZeroUint()
	newWeights := slices.Expand(func(_ int) sdk.Uint {
		newWeight := sdk.NewUint(uint64(rand.I64Between(1, 20)))
		totalWeight = totalWeight.Add(newWeight)

		return newWeight
	}, len(newAddresses))

	operatorshipTransferred := types.EventMultisigOperatorshipTransferred{
		NewOperators: newAddresses,
		NewWeights:   newWeights,
		NewThreshold: sdk.NewUint(uint64(rand.I64Between(1, totalWeight.BigInt().Int64()+1))),
	}
	event.Event = &types.Event_MultisigOperatorshipTransferred{
		MultisigOperatorshipTransferred: &operatorshipTransferred,
	}

	return event
}
