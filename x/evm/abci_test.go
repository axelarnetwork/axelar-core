package evm

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmTestUtils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
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

	sourceChainName := rand.Str(5)
	destinationChainName := rand.Str(5)
	payload := rand.Bytes(100)

	givenContractCallEvent := Given("a ContractCall event", func(t *testing.T) {
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
					Payload:          payload,
				},
			},
		}
		ctx, bk, n, s, sourceCk, destinationCk = setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
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
		When("the source chain is registered", func(t *testing.T) {
			n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
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

	isDestinationChainEvm := func(isEvm bool) func(t *testing.T) {
		return func(t *testing.T) {
			bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return chain == destinationChainName && isEvm }
		}
	}

	isDestinationChainActivated := func(isActivated bool) func(t *testing.T) {
		return func(t *testing.T) {
			n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool {
				return chain.Name == destinationChainName && isActivated
			}
		}
	}

	isDestinationChainIDSet := func(isSet bool) func(t *testing.T) {
		return func(t *testing.T) {
			destinationCk.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.ZeroInt(), isSet }
		}
	}

	isCurrentSecondaryKeySet := func(isSet bool) func(t *testing.T) {
		return func(t *testing.T) {
			s.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				if !isSet {
					return "", false
				}

				return tssTestUtils.RandKeyID(), true
			}
		}
	}

	enqueueCommandSucceed := func(isSuccessful bool) func(t *testing.T) {
		return func(t *testing.T) {
			destinationCk.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error {
				if !isSuccessful {
					return fmt.Errorf("enqueue error")
				}

				return nil
			}
		}
	}

	givenContractCallEvent.
		When("the source chain is not registered", func(t *testing.T) {
			n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{}, chain != sourceChainName
			}
		}).
		Then("should panic", panicWith(fmt.Sprintf("%s is not a registered chain", sourceChainName))).
		Run(t)

	givenContractCallEvent.
		When("the destination chain is not registered", func(t *testing.T) {
			n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{}, chain != destinationChainName
			}
		}).
		Then("should return false", func(t *testing.T) {
			ok := handleContractCall(ctx, event, bk, n, s)
			assert.False(t, ok)
		}).
		Run(t)

	whenChainsAreRegistered.
		And().
		When("destination chain is not an evm chain", isDestinationChainEvm(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleContractCall(ctx, event, bk, n, s)
			assert.False(t, ok)
		}).
		Run(t)

	whenChainsAreRegistered.
		And().
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		And().
		When("destination chain is not activated", isDestinationChainActivated(false)).
		Then("should return false", func(t *testing.T) {
			ok := handleContractCall(ctx, event, bk, n, s)
			assert.False(t, ok)
		}).
		Run(t)

	whenChainsAreRegistered.
		And().
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		And().
		When("destination chain is not activated", isDestinationChainActivated(true)).
		And().
		When("destination chain ID is not set", isDestinationChainIDSet(false)).
		Then("should panic", panicWith(fmt.Sprintf("could not find chain ID for '%s'", destinationChainName))).
		Run(t)

	whenChainsAreRegistered.
		And().
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		And().
		When("destination chain is not activated", isDestinationChainActivated(true)).
		And().
		When("destination chain ID is set", isDestinationChainIDSet(true)).
		And().
		When("current secondary key is not set", isCurrentSecondaryKeySet(false)).
		Then("should panic", panicWith(fmt.Sprintf("no secondary key for chain %s found", destinationChainName))).
		Run(t)

	whenChainsAreRegistered.
		And().
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		And().
		When("destination chain is not activated", isDestinationChainActivated(true)).
		And().
		When("destination chain ID is set", isDestinationChainIDSet(true)).
		And().
		When("current secondary key is set", isCurrentSecondaryKeySet(true)).
		And().
		When("enqueue command fails", enqueueCommandSucceed(false)).
		Then("should panic", panicWith("enqueue error")).
		Run(t)

	whenChainsAreRegistered.
		And().
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		And().
		When("destination chain is not activated", isDestinationChainActivated(true)).
		And().
		When("destination chain ID is set", isDestinationChainIDSet(true)).
		And().
		When("current secondary key is set", isCurrentSecondaryKeySet(true)).
		And().
		When("enqueue command succeeds", enqueueCommandSucceed(true)).
		Then("should return true", func(t *testing.T) {
			ok := handleContractCall(ctx, event, bk, n, s)
			assert.True(t, ok)
			assert.Len(t, destinationCk.EnqueueCommandCalls(), 1)
		}).
		Run(t)
}

func TestHandleTokenSent(t *testing.T) {
	sourceChainName := rand.Str(5)
	destinationChainName := rand.Str(5)
	event := types.Event{
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

	t.Run("should panic if the source chain is not registered", testutils.Func(func(t *testing.T) {
		ctx, bk, n, _, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, chain != sourceChainName
		}

		assert.PanicsWithError(t, fmt.Sprintf("%s is not a registered chain", sourceChainName), func() {
			handleTokenSent(ctx, event, bk, n)
		})
	}))

	t.Run("should return false if the destination chain is not registered", testutils.Func(func(t *testing.T) {
		ctx, bk, n, _, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, chain != destinationChainName
		}

		ok := handleTokenSent(ctx, event, bk, n)
		assert.False(t, ok)
	}))

	t.Run("should return false if the token is not confirmed on the source chain", testutils.Func(func(t *testing.T) {
		ctx, bk, n, _, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			return types.NilToken
		}

		ok := handleTokenSent(ctx, event, bk, n)
		assert.False(t, ok)
	}))

	t.Run("should return false if the token is not confirmed on the destination chain", testutils.Func(func(t *testing.T) {
		ctx, bk, n, _, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			if symbol == event.GetTokenSent().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
			}
			return types.NilToken
		}
		destinationCk.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			return types.NilToken
		}

		ok := handleTokenSent(ctx, event, bk, n)
		assert.False(t, ok)
	}))

	t.Run("should return false if failed to enqueue the transfer", testutils.Func(func(t *testing.T) {
		ctx, bk, n, _, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			if symbol == event.GetTokenSent().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
			}
			return types.NilToken
		}
		destinationCk.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			if asset == event.GetTokenSent().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: asset})
			}
			return types.NilToken
		}
		n.EnqueueTransferFunc = func(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error) {
			return 0, fmt.Errorf("err")
		}

		ok := handleTokenSent(ctx, event, bk, n)
		assert.False(t, ok)
	}))

	t.Run("should return true if succeeded to enqueue the transfer", testutils.Func(func(t *testing.T) {
		ctx, bk, n, _, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			if symbol == event.GetTokenSent().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
			}
			return types.NilToken
		}
		destinationCk.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			if asset == event.GetTokenSent().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: asset})
			}
			return types.NilToken
		}
		n.EnqueueTransferFunc = func(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error) {
			return 0, nil
		}

		ok := handleTokenSent(ctx, event, bk, n)
		assert.True(t, ok)
		assert.Len(t, n.EnqueueTransferCalls(), 1)
	}))
}

func TestHandleContractCallWithToken(t *testing.T) {
	payload := rand.Bytes(100)
	sourceChainName := rand.Str(5)
	destinationChainName := rand.Str(5)
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
				Payload:          payload,
				Symbol:           rand.Denom(3, 5),
				Amount:           sdk.NewUint(uint64(rand.I64Between(1, 10000))),
			},
		},
	}

	t.Run("should panic if the source chain is not registered", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, chain != sourceChainName
		}

		assert.PanicsWithError(t, fmt.Sprintf("%s is not a registered chain", sourceChainName), func() {
			handleContractCallWithToken(ctx, event, bk, n, s)
		})
	}))

	t.Run("should panic if the destination chain is not registered", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, chain != destinationChainName
		}

		ok := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.False(t, ok)
	}))

	t.Run("should return false if the token is not confirmed on the source chain", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			return types.NilToken
		}

		ok := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.False(t, ok)
	}))

	t.Run("should return false if the token is not confirmed on the destination chain", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
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

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
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

	t.Run("should return false if failed to compute transfer fee", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
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
			return sdk.Coin{}, fmt.Errorf("")
		}

		ok := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.False(t, ok)
	}))

	t.Run("should return false if the amount is not enough to cover the fee", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()
		fee := sdk.NewCoin(event.GetContractCallWithToken().Symbol, sdk.Int(event.GetContractCallWithToken().Amount.AddUint64(1)))

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
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

		ok := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.False(t, ok)
	}))

	t.Run("should panic if the destination chain ID is not found", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()
		fee := sdk.NewCoin(event.GetContractCallWithToken().Symbol, sdk.NewInt(rand.I64Between(1, event.GetContractCallWithToken().Amount.BigInt().Int64())))

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
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

	t.Run("should panic if the destination chain does not have the secondary key set", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()
		fee := sdk.NewCoin(event.GetContractCallWithToken().Symbol, sdk.NewInt(rand.I64Between(1, event.GetContractCallWithToken().Amount.BigInt().Int64())))

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
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

		assert.PanicsWithError(t, fmt.Sprintf("no secondary key for chain %s found", destinationChainName), func() {
			handleContractCallWithToken(ctx, event, bk, n, s)
		})
	}))

	t.Run("should return true if successfully created the command", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()
		fee := sdk.NewCoin(event.GetContractCallWithToken().Symbol, sdk.NewInt(rand.I64Between(1, event.GetContractCallWithToken().Amount.BigInt().Int64())))

		bk.ForChainFunc = func(chain string) types.ChainKeeper {
			switch chain {
			case sourceChainName:
				return sourceCk
			case destinationChainName:
				return destinationCk
			default:
				return nil
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain}, true
			default:
				return nexus.Chain{}, false
			}
		}
		bk.HasChainFunc = func(ctx sdk.Context, chain string) bool { return true }
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
			return tssTestUtils.RandKeyID(), true
		}
		destinationCk.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error { return nil }
		n.AddTransferFeeFunc = func(ctx sdk.Context, coin sdk.Coin) {}

		ok := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.True(t, ok)
		assert.Len(t, destinationCk.EnqueueCommandCalls(), 1)
		assert.Len(t, n.AddTransferFeeCalls(), 1)
		assert.Equal(t, n.AddTransferFeeCalls()[0].Coin, fee)
	}))
}
