package evm

import (
	"errors"
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
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmTestUtils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTestUtils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	multisigTypesTestuilts "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
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
	destinationCk.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
	return ctx, bk, n, multisigKeeper, sourceCk, destinationCk
}

func TestHandleGeneralMessage(t *testing.T) {

	var (
		ctx            sdk.Context
		n              *mock.NexusMock
		multisigKeeper *mock.MultisigKeeperMock
		destinationCk  *mock.ChainKeeperMock

		msg nexus.GeneralMessage
	)

	destinationChain := nexus.Chain{Name: nexustestutils.RandomChainName(), Module: types.ModuleName}
	sourceChain := nexus.Chain{Name: nexustestutils.RandomChainName(), Module: types.ModuleName}
	sender := nexus.CrossChainAddress{Chain: sourceChain, Address: evmTestUtils.RandomAddress().Hex()}
	receiver := nexus.CrossChainAddress{Chain: destinationChain, Address: evmTestUtils.RandomAddress().Hex()}
	payload := rand.Bytes(100)
	asset := rand.Coin()

	givenMessage := Given("a message", func() {
		msg = nexus.NewGeneralMessage(evmTestUtils.RandomHash().Hex(), sender, receiver, evmCrypto.Keccak256(payload), evmTestUtils.RandomHash().Bytes()[:], uint64(rand.I64Between(0, 10000)), nil)

		ctx, _, n, multisigKeeper, _, destinationCk = setup()
		n.SetMessageFailedFunc = func(ctx sdk.Context, id string) error {
			return nil
		}

	})

	withToken := Given("with token", func() {
		msg.Asset = &asset

		destinationCk.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			if asset == msg.Asset.Denom {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed})
			}
			return types.NilToken
		}
	})

	panicWith := func(errMsg string) func(t *testing.T) {
		return func(t *testing.T) {
			assert.PanicsWithError(t, errMsg, func() {
				handleMessage(ctx, destinationCk, rand.IntBetween(sdk.ZeroInt(), sdk.NewInt(10000)), multisigTestUtils.KeyID(), msg)
			})
		}
	}

	validationFailedWith := func(errMsg string) func(t *testing.T) {
		return func(t *testing.T) {
			assert.ErrorContains(t, validateMessage(ctx, destinationCk, n, multisigKeeper, destinationChain, msg), errMsg)
		}
	}

	isChainActivated := func(isActivated bool) func() {
		return func() {
			n.IsChainActivatedFunc = func(_ sdk.Context, _ nexus.Chain) bool { return isActivated }
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

	isGatewayAddressSet := func(isSet bool) func() {
		return func() {
			destinationCk.GetGatewayAddressFunc = func(ctx sdk.Context) (types.Address, bool) {
				if !isSet {
					return types.ZeroAddress, false
				}
				return evmTestUtils.RandomAddress(), true
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

	givenMessage.
		When("current key not set", isCurrentKeySet(false)).
		Then("should fail validation", validationFailedWith("current key not set")).
		Run(t)

	givenMessage.
		When("current key is set", isCurrentKeySet(true)).
		When("chain is not activated", isChainActivated(false)).
		Then("should fail validation", validationFailedWith("destination chain de-activated")).
		Run(t)

	givenMessage.
		When("current key is set", isCurrentKeySet(true)).
		When("chain is activated", isChainActivated(true)).
		When("gateway not set", isGatewayAddressSet(false)).
		Then("should fail validation", validationFailedWith("destination chain gateway not deployed yet")).
		Run(t)

	givenMessage.
		When("current key is set", isCurrentKeySet(true)).
		When("chain is activated", isChainActivated(true)).
		When("gateway is set", isGatewayAddressSet(true)).
		When("recipient address is invalid", func() {
			msg.Recipient.Address = "0xFF"
		}).
		Then("should fail validation", validationFailedWith("invalid contract address")).
		Run(t)

	givenMessage.
		Given2(withToken).
		When("current key is set", isCurrentKeySet(true)).
		When("chain is activated", isChainActivated(true)).
		When("gateway is set", isGatewayAddressSet(true)).
		When("token not confirmed on destination chain", func() {
			destinationCk.GetERC20TokenByAssetFunc = func(sdk.Context, string) types.ERC20Token {
				return types.NilToken
			}
		}).
		Then("should fail validation", validationFailedWith(
			fmt.Sprintf("asset %s not confirmed on destination chain", asset.GetDenom())),
		).
		Run(t)

	givenMessage.
		Given2(withToken).
		When("current key is set", isCurrentKeySet(true)).
		When("chain is activated", isChainActivated(true)).
		When("gateway is set", isGatewayAddressSet(true)).
		When("token not confirmed on destination chain", func() {
			destinationCk.GetERC20TokenByAssetFunc = func(sdk.Context, string) types.ERC20Token {
				return types.NilToken
			}
		}).
		Then("should fail validation", validationFailedWith(
			fmt.Sprintf("asset %s not confirmed on destination chain", asset.GetDenom())),
		).
		Run(t)

	givenValidMessage := givenMessage.Given("message is valid", func() {
		isCurrentKeySet(true)()
		isChainActivated(true)()
		isGatewayAddressSet(true)()
	})

	var err error
	givenValidMessage.
		Given2(withToken).
		When("validate message", func() {
			err = validateMessage(ctx, destinationCk, n, multisigKeeper, destinationChain, msg)
		}).
		Then("should pass validation", func(t *testing.T) {
			assert.NoError(t, err)
		}).
		Run(t)

	givenValidMessage.
		When("enqueue command fails", enqueueCommandSucceed(false)).
		Then("should panic", panicWith("call should not have failed: enqueue error")).
		Run(t)

	givenValidMessage.
		Given2(withToken).
		When("enqueue command succeeds", enqueueCommandSucceed(true)).
		Then("should succeed", func(t *testing.T) {
			handleMessage(ctx, destinationCk, rand.IntBetween(sdk.ZeroInt(), sdk.NewInt(10000)), multisigTestUtils.KeyID(), msg)
			assert.Len(t, destinationCk.EnqueueCommandCalls(), 1)
		}).
		Run(t)
}

func TestHandleGeneralMessages(t *testing.T) {
	var (
		ctx            sdk.Context
		bk             *mock.BaseKeeperMock
		n              *mock.NexusMock
		multisigKeeper *mock.MultisigKeeperMock
		ck1            *mock.ChainKeeperMock
		ck2            *mock.ChainKeeperMock
		ck3            *mock.ChainKeeperMock
	)
	chain1 := nexus.ChainName(rand.Str(5))
	chain2 := nexus.ChainName(rand.Str(5))

	givenGeneralMessagesEnqueued := Given("general messages queued", func() {
		ctx, bk, n, multisigKeeper, ck1, ck2 = setup()
		n.GetChainsFunc = func(_ sdk.Context) []nexus.Chain {
			return []nexus.Chain{{Name: chain1, Module: types.ModuleName}, {Name: chain2, Module: types.ModuleName}}
		}
		multisigKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), true
		}
		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case chain1:
				return ck1, nil
			case chain2:
				return ck2, nil
			default:
				return nil, errors.New("not found")
			}
		}
		ck1.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.ZeroInt(), true }
		ck1.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error { return nil }
		ck2.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.ZeroInt(), true }
		ck2.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error { return nil }
		ck1.GetGatewayAddressFunc = func(ctx sdk.Context) (types.Address, bool) { return evmTestUtils.RandomAddress(), true }
		ck2.GetGatewayAddressFunc = func(ctx sdk.Context) (types.Address, bool) { return evmTestUtils.RandomAddress(), true }
		n.SetMessageExecutedFunc = func(ctx sdk.Context, id string) error { return nil }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
	})

	withGeneralMessages := func(numPerChain map[nexus.ChainName]int) WhenStatement {
		return When("having general messages", func() {
			n.GetProcessingMessagesFunc = func(_ sdk.Context, chain nexus.ChainName, limit int64) []nexus.GeneralMessage {
				msgs := []nexus.GeneralMessage{}

				for i := 0; i < int(limit) && i < numPerChain[chain]; i++ {
					srcChain := nexustestutils.RandomChain()
					srcChain.Module = types.ModuleName
					destChain := nexus.Chain{Name: chain, Module: types.ModuleName}
					sender := nexus.CrossChainAddress{Chain: srcChain, Address: evmTestUtils.RandomAddress().Hex()}
					receiver := nexus.CrossChainAddress{Chain: destChain, Address: evmTestUtils.RandomAddress().Hex()}

					msg := nexus.NewGeneralMessage(evmTestUtils.RandomHash().Hex(), sender, receiver, evmTestUtils.RandomHash().Bytes(), evmTestUtils.RandomHash().Bytes()[:], uint64(rand.I64Between(0, 10000)), nil)
					msg.Status = nexus.Processing

					msgs = append(msgs, msg)
				}

				return msgs
			}
		})
	}

	panicWith := func(msg string) func(t *testing.T) {
		return func(t *testing.T) {
			assert.PanicsWithError(t, msg, func() {
				handleMessages(ctx, bk, n, multisigKeeper)
			})
		}
	}

	givenGeneralMessagesEnqueued.
		When2(withGeneralMessages(map[nexus.ChainName]int{})).
		When("ForChainFails", func() {
			bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
				return nil, errors.New("not found")
			}
		}).
		Then("should panic", panicWith("call should not have failed: not found")).
		Run(t)

	givenGeneralMessagesEnqueued.
		When2(withGeneralMessages(map[nexus.ChainName]int{})).
		When("no messages", func() {
			ck1.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: 50}
			}
			ck2.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: 50}
			}
		}).
		Then("should handle", func(t *testing.T) {
			handleMessages(ctx, bk, n, multisigKeeper)
			assert.Len(t, ck1.EnqueueCommandCalls(), 0)
			assert.Len(t, ck2.EnqueueCommandCalls(), 0)
		}).
		Run(t)

	givenGeneralMessagesEnqueued.
		When2(withGeneralMessages(map[nexus.ChainName]int{chain1: 10})).
		When("end blocker limit is greater than num messages", func() {
			ck1.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: 50}
			}
			ck2.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: 50}
			}
		}).
		Then("should handle", func(t *testing.T) {
			handleMessages(ctx, bk, n, multisigKeeper)
			assert.Len(t, ck1.EnqueueCommandCalls(), 10)
			assert.Len(t, ck2.EnqueueCommandCalls(), 0)
		}).
		Run(t)

	givenGeneralMessagesEnqueued.
		When2(withGeneralMessages(map[nexus.ChainName]int{chain2: 100})).
		When("end blocker limit is less than num messages", func() {
			ck1.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: 50}
			}
			ck2.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: 50}
			}
		}).
		Then("should handle", func(t *testing.T) {
			handleMessages(ctx, bk, n, multisigKeeper)
			assert.Len(t, ck1.EnqueueCommandCalls(), 0)
			assert.Len(t, ck2.EnqueueCommandCalls(), int(ck2.GetParams(ctx).EndBlockerLimit))
		}).
		Run(t)

	givenGeneralMessagesEnqueued.
		When2(withGeneralMessages(map[nexus.ChainName]int{chain1: 100, chain2: 65})).
		When("multiple chains", func() {
			ck1.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: 45}
			}
			ck2.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: 40}
			}
		}).
		Then("should handle", func(t *testing.T) {
			handleMessages(ctx, bk, n, multisigKeeper)
			assert.Len(t, ck1.EnqueueCommandCalls(), int(ck1.GetParams(ctx).EndBlockerLimit))
			assert.Len(t, ck2.EnqueueCommandCalls(), int(ck2.GetParams(ctx).EndBlockerLimit))
		}).
		Run(t)

	givenGeneralMessagesEnqueued.
		When("some associated EVM event exists", func() {
			srcChain := nexustestutils.RandomChain()
			srcChain.Module = types.ModuleName

			ck1.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: 45}
			}
			ck2.GetParamsFunc = func(ctx sdk.Context) types.Params {
				return types.Params{EndBlockerLimit: 40}
			}
			ck3 = &mock.ChainKeeperMock{}

			bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
				switch chain {
				case chain1:
					return ck1, nil
				case chain2:
					return ck2, nil
				case srcChain.Name:
					return ck3, nil
				default:
					return nil, errors.New("not found")
				}
			}

			n.GetProcessingMessagesFunc = func(_ sdk.Context, chain nexus.ChainName, limit int64) []nexus.GeneralMessage {
				destChain := nexus.Chain{Name: chain, Module: types.ModuleName}
				sender := nexus.CrossChainAddress{Chain: srcChain, Address: evmTestUtils.RandomAddress().Hex()}
				receiver := nexus.CrossChainAddress{Chain: destChain, Address: evmTestUtils.RandomAddress().Hex()}

				msg := nexus.NewGeneralMessage(evmTestUtils.RandomHash().Hex(), sender, receiver, evmTestUtils.RandomHash().Bytes(), evmTestUtils.RandomHash().Bytes()[:], uint64(rand.I64Between(0, 10000)), nil)

				return []nexus.GeneralMessage{msg}
			}
		}).
		Then("should handle", func(t *testing.T) {
			ck3.SetEventCompletedFunc = func(ctx sdk.Context, eventID types.EventID) error { return nil }

			handleMessages(ctx, bk, n, multisigKeeper)

			assert.Len(t, ck1.EnqueueCommandCalls(), 1)
			assert.Len(t, ck2.EnqueueCommandCalls(), 1)
			assert.Len(t, ck3.SetEventCompletedCalls(), 2)
		}).
		Run(t)
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
		isEvm          bool
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

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				if isEvm {
					return destinationCk, nil
				}
				return nil, errors.New("not an EVM chain")
			default:
				return nil, errors.New("not found")
			}
		}
	})

	whenChainsAreRegistered := givenContractCallEvent.
		When("the source and destination chains are registered", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				switch chain {
				case sourceChainName, destinationChainName:
					return nexus.Chain{Name: chain, Module: types.ModuleName}, true
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

	errorWith := func(msg string) func(t *testing.T) {
		return func(t *testing.T) {
			assert.ErrorContains(t, handleContractCall(ctx, event, bk, n, multisigKeeper), msg)
		}
	}

	isDestinationChainEvm := func(_isEvm bool) func() {
		return func() {
			isEvm = _isEvm
		}
	}

	destinationChainIsCosmos := func() func() {
		return func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				switch chain {
				case sourceChainName:
					return nexus.Chain{Name: chain, Module: types.ModuleName}, true
				case destinationChainName:
					return nexus.Chain{Name: chain, Module: axelarnet.ModuleName}, true
				default:
					return nexus.Chain{}, false
				}
			}
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

	setGeneralMessageSucceed := func(isSuccessful bool) func() {
		return func() {
			n.SetNewMessageFunc = func(sdk.Context, nexus.GeneralMessage) error {
				if !isSuccessful {
					return fmt.Errorf("set general message error")
				}

				return nil
			}
		}
	}

	givenContractCallEvent.
		When("the destination chain is not registered", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{}, chain != destinationChainName
			}
		}).
		Then("should panic", panicWith("result is not found")).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is not an evm chain", isDestinationChainEvm(false)).
		Then("should panic", panicWith("call should not have failed: not an EVM chain")).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		When("destination chain ID is not set", isDestinationChainIDSet(false)).
		Then("should panic", panicWith("result is not found")).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		When("destination chain ID is set", isDestinationChainIDSet(true)).
		When("current key is not set", isCurrentKeySet(false)).
		Then("should panic", panicWith("result is not found")).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		When("destination chain ID is set", isDestinationChainIDSet(true)).
		When("current key is set", isCurrentKeySet(true)).
		When("enqueue command fails", enqueueCommandSucceed(false)).
		Then("should panic", panicWith("call should not have failed: enqueue error")).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is an evm chain", isDestinationChainEvm(true)).
		When("destination chain ID is set", isDestinationChainIDSet(true)).
		When("current key is set", isCurrentKeySet(true)).
		When("enqueue command succeeds", enqueueCommandSucceed(true)).
		Then("should succeed", func(t *testing.T) {
			err := handleContractCall(ctx, event, bk, n, multisigKeeper)
			assert.NoError(t, err)
			assert.Len(t, destinationCk.EnqueueCommandCalls(), 1)
		}).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is a cosmos chain", destinationChainIsCosmos()).
		When("set general message fails", setGeneralMessageSucceed(false)).
		Then("should fail", errorWith("set general message error")).
		Run(t)

	whenChainsAreRegistered.
		When("destination chain is a cosmos chain", destinationChainIsCosmos()).
		When("set general message succeeds", setGeneralMessageSucceed(true)).
		Then("should succeed", func(t *testing.T) {
			err := handleContractCall(ctx, event, bk, n, multisigKeeper)
			assert.NoError(t, err)
			assert.Len(t, n.SetNewMessageCalls(), 1)
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
		When("the source and destination chains are registered", func() {
			bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
				switch chain {
				case sourceChainName:
					return sourceCk, nil
				case destinationChainName:
					return destinationCk, nil
				default:
					return nil, errors.New("not found")
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

			n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		})

	panicWith := func(msg string) func(t *testing.T) {
		return func(t *testing.T) {
			assert.PanicsWithError(t, msg, func() {
				handleTokenSent(ctx, event, bk, n)
			})
		}
	}

	assertFail := func(t *testing.T) {
		err := handleTokenSent(ctx, event, bk, n)
		assert.Error(t, err)
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
			bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
				switch chain {
				case sourceChainName:
					return sourceCk, nil
				case destinationChainName:
					if isEvm {
						return destinationCk, nil
					}

					return nil, errors.New("not an EVM chain")
				default:
					return nil, errors.New("not found")
				}
			}

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
		Then("should panic", panicWith("result is not found")).
		Run(t)

	givenTokenSentEvent.
		When("destination chain is not registered", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{}, chain != destinationChainName
			}
		}).
		Then("should panic", panicWith("result is not found")).
		Run(t)

	whenChainsAreRegistered.
		When("token is not confirmed on the source chain", tokenRegisteredOnChain(&sourceCk, false)).
		Then("should fail", assertFail).
		Run(t)

	whenChainsAreRegistered.
		When("token is confirmed on the source chain", tokenRegisteredOnChain(&sourceCk, true)).
		When("token is not confirmed on the destination chain", tokenRegisteredOnChain(&destinationCk, false)).
		Then("should fail", assertFail).
		Run(t)

	whenTokensAreConfirmed.
		When("failed to enqueue the transfer", func() {
			n.EnqueueTransferFunc = func(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error) {
				return 0, fmt.Errorf("err")
			}
		}).
		Then("should fail", assertFail).
		Run(t)

	whenTokensAreConfirmed.
		When("enqueue the transfer", func() {
			n.EnqueueTransferFunc = func(ctx sdk.Context, senderChain nexus.Chain, recipient nexus.CrossChainAddress, asset sdk.Coin) (nexus.TransferID, error) {
				return 0, nil
			}
		}).
		Then("should succeed", func(t *testing.T) {
			err := handleTokenSent(ctx, event, bk, n)
			assert.NoError(t, err)
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
		Then("should succeed", func(t *testing.T) {
			err := handleTokenSent(ctx, event, bk, n)
			assert.NoError(t, err)
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

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				return destinationCk, nil
			default:
				return nil, errors.New("not found")
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{}, chain != sourceChainName
		}

		assert.PanicsWithError(t, "result is not found", func() {
			handleContractCallWithToken(ctx, event, bk, n, s)
		})
	}))

	t.Run("should panic if the destination chain is not registered", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				return destinationCk, nil
			default:
				return nil, errors.New("not found")
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{}, chain != destinationChainName
		}

		assert.PanicsWithError(t, "result is not found", func() {
			handleContractCallWithToken(ctx, event, bk, n, s)
		})
	}))

	t.Run("should fail if the token is not confirmed on the source chain", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				return destinationCk, nil
			default:
				return nil, errors.New("not found")
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			default:
				return nexus.Chain{}, false
			}
		}

		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			return types.NilToken
		}

		err := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.Error(t, err)
	}))

	t.Run("should fail if the token is not confirmed on the destination chain", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				return destinationCk, nil
			default:
				return nil, errors.New("not found")
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			default:
				return nexus.Chain{}, false
			}
		}

		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		n.RateLimitTransferFunc = func(sdk.Context, nexus.ChainName, sdk.Coin, nexus.TransferDirection) error { return nil }
		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			if symbol == event.GetContractCallWithToken().Symbol {
				return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
			}
			return types.NilToken
		}
		destinationCk.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			return types.NilToken
		}

		err := handleContractCallWithToken(ctx, event, bk, n, s)
		assert.Error(t, err)
	}))

	t.Run("should fail if the contract address is invalid", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				return destinationCk, nil
			default:
				return nil, errors.New("not found")
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			default:
				return nexus.Chain{}, false
			}
		}

		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		n.RateLimitTransferFunc = func(sdk.Context, nexus.ChainName, sdk.Coin, nexus.TransferDirection) error { return nil }
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

		err := handleContractCallWithToken(ctx, event, bk, n, s)
		event.GetContractCallWithToken().ContractAddress = contractAddress
		assert.Error(t, err)
	}))

	t.Run("should panic if the destination chain ID is not found", testutils.Func(func(t *testing.T) {
		ctx, bk, n, s, sourceCk, destinationCk := setup()
		fee := sdk.NewCoin(event.GetContractCallWithToken().Symbol, sdk.NewInt(rand.I64Between(1, event.GetContractCallWithToken().Amount.BigInt().Int64())))

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				return destinationCk, nil
			default:
				return nil, errors.New("not found")
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			default:
				return nexus.Chain{}, false
			}
		}

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
		n.RateLimitTransferFunc = func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
			return nil
		}
		destinationCk.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.ZeroInt(), false }

		assert.PanicsWithError(t, "result is not found", func() {
			handleContractCallWithToken(ctx, event, bk, n, s)
		})
	}))

	t.Run("should panic if the destination chain does not have the key set", testutils.Func(func(t *testing.T) {
		ctx, bk, n, multisigKeeper, sourceCk, destinationCk := setup()
		fee := sdk.NewCoin(event.GetContractCallWithToken().Symbol, sdk.NewInt(rand.I64Between(1, event.GetContractCallWithToken().Amount.BigInt().Int64())))

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				return destinationCk, nil
			default:
				return nil, errors.New("not found")
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			default:
				return nexus.Chain{}, false
			}
		}

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
		n.RateLimitTransferFunc = func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
			return nil
		}
		destinationCk.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.NewInt(1), true }
		multisigKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), false
		}

		assert.PanicsWithError(t, "result is not found", func() {
			handleContractCallWithToken(ctx, event, bk, n, multisigKeeper)
		})
	}))

	t.Run("should fail if rate limit is exceeded", testutils.Func(func(t *testing.T) {
		ctx, bk, n, multisigKeeper, sourceCk, destinationCk := setup()
		fee := sdk.NewCoin(event.GetContractCallWithToken().Symbol, sdk.NewInt(rand.I64Between(1, event.GetContractCallWithToken().Amount.BigInt().Int64())))

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				return destinationCk, nil
			default:
				return nil, errors.New("not found")
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			default:
				return nexus.Chain{}, false
			}
		}

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
		n.RateLimitTransferFunc = func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
			return fmt.Errorf("rate limit exceeded")
		}

		err := handleContractCallWithToken(ctx, event, bk, n, multisigKeeper)
		assert.ErrorContains(t, err, "rate limit exceeded")
	}))

	t.Run("should succeed if successfully created the command", testutils.Func(func(t *testing.T) {
		ctx, bk, n, multisigKeeper, sourceCk, destinationCk := setup()

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				return destinationCk, nil
			default:
				return nil, errors.New("not found")
			}
		}
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			default:
				return nexus.Chain{}, false
			}
		}

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
		n.RateLimitTransferFunc = func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
			return nil
		}
		destinationCk.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.NewInt(1), true }
		multisigKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), true
		}

		destinationCk.EnqueueCommandFunc = func(ctx sdk.Context, cmd types.Command) error { return nil }

		err := handleContractCallWithToken(ctx, event, bk, n, multisigKeeper)
		assert.NoError(t, err)
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

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			return sourceCk, nil
		}

		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{Name: sourceChainName}, true
		}

		sourceCk.SetDepositFunc = func(sdk.Context, types.ERC20Deposit, types.DepositStatus) {}
		sourceCk.GetDepositFunc = func(ctx sdk.Context, txID types.Hash, logIndex uint64) (types.ERC20Deposit, types.DepositStatus, bool) {
			return types.ERC20Deposit{}, types.DepositStatus_None, false
		}
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
			sourceCk.GetLegacyDepositFunc = func(ctx sdk.Context, txID types.Hash, burnerAddr types.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
				return types.ERC20Deposit{}, types.DepositStatus_Confirmed, found
			}
		}
	}

	givenTransferEvent.
		When("burner info not found", burnerInfoFound(false)).
		Then("should fail", func(t *testing.T) {
			err := handleConfirmDeposit(ctx, event, bk, n)
			assert.Error(t, err)
			assert.Len(t, n.EnqueueForTransferCalls(), 0)
			assert.Len(t, sourceCk.SetDepositCalls(), 0)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient not found", recipientFound(false)).
		Then("should fail", func(t *testing.T) {
			err := handleConfirmDeposit(ctx, event, bk, n)
			assert.Error(t, err)
			assert.Len(t, n.EnqueueForTransferCalls(), 0)
			assert.Len(t, sourceCk.SetDepositCalls(), 0)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient found", recipientFound(true)).
		When("deposit exists", depositFound(true)).
		Then("should fail", func(t *testing.T) {
			err := handleConfirmDeposit(ctx, event, bk, n)
			assert.Error(t, err)
			assert.Len(t, n.EnqueueForTransferCalls(), 0)
			assert.Len(t, sourceCk.SetDepositCalls(), 0)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient found", recipientFound(true)).
		When("deposit does not exist", depositFound(false)).
		When("failed to enqueue the transfer", enqueueTransferSucceed(false)).
		Then("should fail", func(t *testing.T) {
			err := handleConfirmDeposit(ctx, event, bk, n)
			assert.Error(t, err)
			assert.Len(t, n.EnqueueForTransferCalls(), 1)
			assert.Len(t, sourceCk.SetDepositCalls(), 0)
		}).
		Run(t)

	givenTransferEvent.
		When("burner info found", burnerInfoFound(true)).
		When("recipient found", recipientFound(true)).
		When("deposit does not exist", depositFound(false)).
		When("enqueue the transfer", enqueueTransferSucceed(true)).
		Then("should succeed", func(t *testing.T) {
			err := handleConfirmDeposit(ctx, event, bk, n)
			assert.NoError(t, err)
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
		n        *mock.NexusMock
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
		ctx, bk, n, _, sourceCk, _ = setup()

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			return sourceCk, nil
		}

		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{Name: sourceChainName}, true
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
		Then("should fail", func(t *testing.T) {
			err := handleTokenDeployed(ctx, event, bk, n)
			assert.Error(t, err)
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
		Then("should fail", func(t *testing.T) {
			err := handleTokenDeployed(ctx, event, bk, n)
			assert.Error(t, err)
		}).
		Run(t)

	givenTokenDeployedEvent.
		When("token address in event matches expected address", canGetERC20TokenBySymbol(true)).
		Then("should succeed", func(t *testing.T) {
			err := handleTokenDeployed(ctx, event, bk, n)
			assert.NoError(t, err)
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
		n              *mock.NexusMock
	)

	sourceChainName := nexus.ChainName(rand.Str(5))

	givenMultisigTransferKeyEvent := Given("a MultisigTransferKey event", func() {
		event = randTransferKeyEvent(sourceChainName)
		ctx, bk, n, multisigKeeper, sourceCk, _ = setup()

		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			return sourceCk, nil
		}

		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{Name: sourceChainName}, true
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
		Then("should fail", func(t *testing.T) {
			err := handleMultisigTransferKey(ctx, event, bk, n, multisigKeeper)
			assert.Error(t, err)
		}).
		Run(t)

	givenMultisigTransferKeyEvent.
		When("next key id is set", isCurrentKeySet(true)).
		When("next key not found", KeyFound(false)).
		Then("should fail", func(t *testing.T) {
			err := handleMultisigTransferKey(ctx, event, bk, n, multisigKeeper)
			assert.Error(t, err)
		}).
		Run(t)

	givenMultisigTransferKeyEvent.
		When("next key id is set", isCurrentKeySet(true)).
		When("next key is found, but does not match expected", KeyFound(true)).
		Then("should fail", func(t *testing.T) {
			err := handleMultisigTransferKey(ctx, event, bk, n, multisigKeeper)
			assert.Error(t, err)
		}).
		Run(t)

	givenMultisigTransferKeyEvent.
		When("next key id is set", isCurrentKeySet(true)).
		When("next key is found, matches expected key", keyMatches).
		Then("should succeed", func(t *testing.T) {
			err := handleMultisigTransferKey(ctx, event, bk, n, multisigKeeper)
			assert.NoError(t, err)
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
		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
			switch chain {
			case sourceChainName:
				return sourceCk, nil
			case destinationChainName:
				return destinationCk, nil
			default:
				return nil, errors.New("not found")
			}
		}

		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			switch chain {
			case sourceChainName, destinationChainName:
				return nexus.Chain{Name: chain, Module: types.ModuleName}, true
			default:
				return nexus.Chain{}, true
			}
		}
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		n.GetChainsFunc = func(_ sdk.Context) []nexus.Chain {
			return []nexus.Chain{{Name: sourceChainName, Module: types.ModuleName}}
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

		sourceCk.GetERC20TokenBySymbolFunc = func(ctx sdk.Context, symbol string) types.ERC20Token {
			return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed, Asset: symbol})
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
			handleConfirmedEvents(ctx, bk, n, multisigKeeper)
			assert.Len(t, sourceCk.SetEventCompletedCalls(), int(sourceCk.GetParams(ctx).EndBlockerLimit))
			assert.Len(t, destinationCk.EnqueueCommandCalls(), int(sourceCk.GetParams(ctx).EndBlockerLimit))
		}).
		Run(t)

	withEvents := func(num int) WhenStatement {
		return When(fmt.Sprintf("having %d events", num), func() {
			count := 0
			confirmedEventQueue.DequeueFunc = func(value codec.ProtoMarshaler) bool {
				if count >= num {
					return false
				}
				count++

				event := evmTestUtils.RandomGatewayEvent(types.EventConfirmed)
				event.Chain = sourceChainName
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
		handleConfirmedEvents(ctx, bk, n, multisigKeeper)
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
