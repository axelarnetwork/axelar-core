package keeper_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func randMsg(status exported.GeneralMessage_Status, withAsset ...bool) exported.GeneralMessage {
	var asset *sdk.Coin
	if len(withAsset) > 0 && withAsset[0] {
		coin := rand.Coin()
		asset = &coin
	}

	return exported.GeneralMessage{
		ID: rand.NormalizedStr(10),
		Sender: exported.CrossChainAddress{
			Chain:   nexustestutils.RandomChain(),
			Address: rand.NormalizedStr(42),
		},
		Recipient: exported.CrossChainAddress{
			Chain:   nexustestutils.RandomChain(),
			Address: rand.NormalizedStr(42),
		},
		PayloadHash:   evmtestutils.RandomHash().Bytes(),
		Status:        status,
		Asset:         asset,
		SourceTxID:    evmtestutils.RandomHash().Bytes(),
		SourceTxIndex: uint64(rand.I64Between(0, 100)),
	}
}

func TestRouteMessageQueue(t *testing.T) {
	var (
		msg    exported.GeneralMessage
		ctx    sdk.Context
		keeper nexus.Keeper
	)

	cfg := app.MakeEncodingConfig()
	givenKeeper := Given("the keeper", func() {
		keeper, ctx = setup(cfg)
		keeper.SetMessageRouter(
			types.NewMessageRouter().
				AddRoute(evm.Ethereum.Module, func(_ sdk.Context, rCtx exported.RoutingContext, _ exported.GeneralMessage) error { return nil }),
		)
	})

	givenKeeper.
		Branch(
			When("the general message does not exist", func() {}).
				Then("should fail to enqueue for routing", func(t *testing.T) {
					assert.ErrorContains(t, keeper.EnqueueRouteMessage(ctx, rand.NormalizedStr(10)), "not found")
				}),

			When("the general message is not at the expected status", func() {
				msg = randMsg(exported.Approved)
				msg.Sender = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}
				msg.Recipient = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}

				funcs.MustNoErr(keeper.SetNewMessage(ctx, msg))
				funcs.MustNoErr(keeper.RouteMessage(ctx, msg.ID))
			}).
				Then("should fail to enqueue for routing", func(t *testing.T) {
					assert.ErrorContains(t, keeper.EnqueueRouteMessage(ctx, msg.ID), "general message has to be approved or failed")
				}),

			When("the general message is approved", func() {
				msg = randMsg(exported.Approved)
				msg.Sender = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}
				msg.Recipient = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}

				funcs.MustNoErr(keeper.SetNewMessage(ctx, msg))
			}).
				Then("should enqueue for routing", func(t *testing.T) {
					assert.NoError(t, keeper.EnqueueRouteMessage(ctx, msg.ID))

					actual, ok := keeper.DequeueRouteMessage(ctx)
					assert.True(t, ok)
					assert.Equal(t, msg, actual)

					_, ok = keeper.DequeueRouteMessage(ctx)
					assert.False(t, ok)

					actual, ok = keeper.GetMessage(ctx, msg.ID)
					assert.True(t, ok)
					assert.Equal(t, msg, actual)
				}),
		).
		Run(t)
}

func TestSetNewMessage(t *testing.T) {
	var (
		msg    exported.GeneralMessage
		ctx    sdk.Context
		keeper nexus.Keeper
	)

	cfg := app.MakeEncodingConfig()
	givenKeeper := Given("the keeper", func() {
		keeper, ctx = setup(cfg)
	})

	givenKeeper.
		When("the message already exists", func() {
			msg = randMsg(exported.Approved, true)
			keeper.SetNewMessage(ctx, msg)
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, keeper.SetNewMessage(ctx, msg), "already exists")
		}).
		Run(t)

	givenKeeper.
		When("the message has invalid status", func() {
			msg = randMsg(exported.Processing)
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, keeper.SetNewMessage(ctx, msg), "new general message has to be approved")
		}).
		Run(t)

	givenKeeper.
		When("the message is valid", func() {
			msg = randMsg(exported.Approved)
		}).
		Then("should store the message", func(t *testing.T) {
			assert.NoError(t, keeper.SetNewMessage(ctx, msg))

			actual, ok := keeper.GetMessage(ctx, msg.ID)
			assert.True(t, ok)
			assert.Equal(t, msg, actual)
		}).
		Run(t)
}

func TestRouteMessage(t *testing.T) {
	var (
		msg        exported.GeneralMessage
		ctx        sdk.Context
		keeper     nexus.Keeper
		routeCount uint
		routingCtx exported.RoutingContext
		payload    []byte
	)

	cfg := app.MakeEncodingConfig()
	givenKeeper := Given("the keeper", func() {
		keeper, ctx = setup(cfg)
	})

	givenKeeper.
		When("the message doesn't exist", func() {}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, keeper.RouteMessage(ctx, rand.NormalizedStr(10)), "not found")
		}).
		Run(t)

	givenKeeper.
		When("the message is being processed", func() {
			msg = randMsg(exported.Approved)
			msg.Sender = exported.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			}
			msg.Recipient = exported.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			}

			keeper.SetNewMessage(ctx, msg)
			keeper.RouteMessage(ctx, msg.ID)
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "general message has to be approved or failed")
		}).
		Run(t)

	givenKeeper.
		When("the message is from wasm", func() {
			msg = randMsg(exported.Approved)
			msg.Sender = exported.CrossChainAddress{
				Chain:   nexustestutils.RandomChain(),
				Address: rand.NormalizedStr(42),
			}
			msg.Sender.Chain.Module = wasm.ModuleName
		}).
		Branch(
			When("the destination chain is not registered", func() {
				msg.Recipient.Chain = nexustestutils.RandomChain()

				keeper.SetNewMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "is not registered")
				}),

			When("the destination chain is not activated", func() {
				msg.Recipient = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}

				keeper.DeactivateChain(ctx, msg.Recipient.Chain)
				keeper.SetNewMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "is not activated")
				}),

			When("the destination address is invalid", func() {
				msg.Recipient = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: rand.NormalizedStr(42),
				}

				keeper.SetNewMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "not an hex address")
				}),

			When("the destination chain doesn't support the asset", func() {
				msg.Recipient = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}
				asset := rand.Coin()
				msg.Asset = &asset

				keeper.SetNewMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "does not support foreign asset")
				}),

			When("asset is set", func() {
				msg.Recipient = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}
				msg.Asset = &sdk.Coin{Denom: "external-erc-20", Amount: sdk.NewInt(100)}

				keeper.SetNewMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "asset transfer is not supported for wasm messages")
				}),
		).
		Run(t)

	givenKeeper.
		When("the message is to wasm", func() {
			msg = randMsg(exported.Approved)
			msg.Recipient = exported.CrossChainAddress{
				Chain:   nexustestutils.RandomChain(),
				Address: rand.NormalizedStr(42),
			}
			msg.Recipient.Chain.Module = wasm.ModuleName
		}).
		Branch(
			When("the sender chain is not registered", func() {
				msg.Sender.Chain = nexustestutils.RandomChain()

				keeper.SetNewMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "is not registered")
				}),

			When("the sender chain is not activated", func() {
				msg.Sender = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}

				keeper.DeactivateChain(ctx, msg.Sender.Chain)
				keeper.SetNewMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "is not activated")
				}),

			When("the sender address is invalid", func() {
				msg.Sender = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: rand.NormalizedStr(42),
				}

				keeper.SetNewMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "not an hex address")
				}),

			When("the sender chain doesn't support the asset", func() {
				msg.Sender = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}
				asset := rand.Coin()
				msg.Asset = &asset

				keeper.SetNewMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "does not support foreign asset")
				}),

			When("asset is set", func() {
				msg.Sender = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}
				msg.Asset = &sdk.Coin{Denom: "external-erc-20", Amount: sdk.NewInt(100)}

				keeper.SetNewMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "asset transfer is not supported for wasm messages")
				}),
		).
		Run(t)

	givenKeeper.
		When("the message is valid", func() {
			msg = randMsg(exported.Approved)
			msg.Sender.Chain.Module = wasm.ModuleName
			msg.Recipient = exported.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			}
			payload = rand.Bytes(100)
			msg.PayloadHash = crypto.Keccak256Hash(payload).Bytes()

			keeper.SetNewMessage(ctx, msg)
		}).
		Branch(
			When("no route is registered for the destination chain", func() {}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.RouteMessage(ctx, msg.ID), "no route found")
				}),

			When("the route is registed for the destination chain and routing context is not provided", func() {
				routeCount = 0

				keeper.SetMessageRouter(types.NewMessageRouter().AddRoute(evm.Ethereum.Module, func(_ sdk.Context, routingCtx exported.RoutingContext, _ exported.GeneralMessage) error {
					routeCount++

					if !routingCtx.FeeGranter.Empty() || !routingCtx.Sender.Empty() || routingCtx.Payload != nil {
						return fmt.Errorf("unexpected routing context")
					}

					return nil
				}))
			}).
				Then("should route the message with the default routing context when it is not provided", func(t *testing.T) {
					assert.NoError(t, keeper.RouteMessage(ctx, msg.ID))
					assert.EqualValues(t, 1, routeCount)
				}),

			When("the route is registed for the destination chain and routing context is provided", func() {
				routeCount = 0
				routingCtx = exported.RoutingContext{
					Sender:     rand.AccAddr(),
					FeeGranter: rand.AccAddr(),
					Payload:    payload,
				}

				keeper.SetMessageRouter(types.NewMessageRouter().AddRoute(evm.Ethereum.Module, func(_ sdk.Context, rCtx exported.RoutingContext, _ exported.GeneralMessage) error {
					routeCount++

					if !rCtx.Sender.Equals(routingCtx.Sender) || !rCtx.FeeGranter.Equals(routingCtx.FeeGranter) || !bytes.Equal(rCtx.Payload, routingCtx.Payload) {
						return fmt.Errorf("unexpected routing context")
					}

					return nil
				}))
			}).
				Then("should route the message with the default routing context when it is not provided", func(t *testing.T) {
					assert.NoError(t, keeper.RouteMessage(ctx, msg.ID, routingCtx))
					assert.EqualValues(t, 1, routeCount)
				}),
		).
		Run(t)
}

func TestGenerateMessageID(t *testing.T) {
	var (
		ctx    sdk.Context
		k      nexus.Keeper
		txhash [32]byte
	)

	Given("a keeper", func() {
		cfg := app.MakeEncodingConfig()
		k, ctx = setup(cfg)
	}).
		When("tx bytes are set", func() {
			tx := rand.Bytes(int(rand.I64Between(1, 100)))
			txhash = sha256.Sum256(tx)
			ctx = ctx.WithTxBytes(tx)
		}).
		Then("should return message id with counter 0", func(t *testing.T) {
			for i := range [10]int{} {
				id, txId, txIndex := k.GenerateMessageID(ctx)
				assert.Equal(t, txhash[:], txId)
				assert.Equal(t, uint64(i), txIndex)
				assert.Equal(t, fmt.Sprintf("0x%s-%d", hex.EncodeToString(txhash[:]), i), id)
			}
		}).
		Run(t)
}

func TestStatusTransitions(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	k, ctx := setup(cfg)
	sourceChain := nexustestutils.RandomChain()
	sourceChain.Module = axelarnet.ModuleName
	destinationChain := nexustestutils.RandomChain()
	destinationChain.Module = evmtypes.ModuleName
	id, txID, nonce := k.GenerateMessageID(ctx)
	msg := exported.GeneralMessage{
		ID:            id,
		Sender:        exported.CrossChainAddress{Chain: sourceChain, Address: genCosmosAddr(sourceChain.Name.String())},
		Recipient:     exported.CrossChainAddress{Chain: destinationChain, Address: evmtestutils.RandomAddress().Hex()},
		Status:        exported.Approved,
		PayloadHash:   crypto.Keccak256Hash(rand.Bytes(int(rand.I64Between(1, 100)))).Bytes(),
		Asset:         nil,
		SourceTxID:    txID,
		SourceTxIndex: nonce,
	}
	k.SetChain(ctx, sourceChain)
	k.SetChain(ctx, destinationChain)
	k.ActivateChain(ctx, sourceChain)
	k.ActivateChain(ctx, destinationChain)
	k.SetMessageRouter(types.NewMessageRouter().AddRoute(destinationChain.Module, func(_ sdk.Context, _ exported.RoutingContext, _ exported.GeneralMessage) error {
		return nil
	}))

	// Message doesn't exist, can't set any status
	err := k.SetMessageFailed(ctx, msg.ID)
	assert.Error(t, err, fmt.Sprintf("general message %s not found", msg.ID))

	err = k.RouteMessage(ctx, msg.ID)
	assert.Error(t, err, fmt.Sprintf("general message %s not found", msg.ID))

	err = k.SetMessageExecuted(ctx, msg.ID)
	assert.Error(t, err, fmt.Sprintf("general message %s not found", msg.ID))

	// Now store the message with approved status
	err = k.SetNewMessage(ctx, msg)
	assert.NoError(t, err)

	err = k.SetMessageFailed(ctx, msg.ID)
	assert.Error(t, err, "general message is not processed")

	err = k.SetMessageExecuted(ctx, msg.ID)
	assert.Error(t, err, "general message is not processed")

	err = k.RouteMessage(ctx, msg.ID)
	assert.NoError(t, err)

	err = k.RouteMessage(ctx, msg.ID)
	assert.Error(t, err, "general message is not approved or failed")

	err = k.SetMessageFailed(ctx, msg.ID)
	assert.NoError(t, err)

	err = k.SetMessageExecuted(ctx, msg.ID)
	assert.Error(t, err, "general message is not processed")

	err = k.RouteMessage(ctx, msg.ID)
	assert.NoError(t, err)

	err = k.SetMessageExecuted(ctx, msg.ID)
	assert.NoError(t, err)

	err = k.SetMessageFailed(ctx, msg.ID)
	assert.Error(t, err, "general message is not processed")

	err = k.RouteMessage(ctx, msg.ID)
	assert.Error(t, err, "general message is not approved or failed")
}

func TestGetMessage(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	k, ctx := setup(cfg)
	sourceChain := nexustestutils.RandomChain()
	sourceChain.Module = axelarnet.ModuleName
	destinationChain := nexustestutils.RandomChain()
	destinationChain.Module = evmtypes.ModuleName
	id, txID, nonce := k.GenerateMessageID(ctx)
	msg := exported.GeneralMessage{
		ID:            id,
		Sender:        exported.CrossChainAddress{Chain: sourceChain, Address: genCosmosAddr(sourceChain.Name.String())},
		Recipient:     exported.CrossChainAddress{Chain: destinationChain, Address: evmtestutils.RandomAddress().Hex()},
		Status:        exported.Approved,
		PayloadHash:   crypto.Keccak256Hash(rand.Bytes(int(rand.I64Between(1, 100)))).Bytes(),
		Asset:         nil,
		SourceTxID:    txID,
		SourceTxIndex: nonce,
	}

	err := k.SetNewMessage(ctx, msg)
	assert.NoError(t, err)

	exp, found := k.GetMessage(ctx, msg.ID)
	assert.True(t, found)
	assert.Equal(t, exp, msg)
}

func TestGetSentMessages(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	k, ctx := setup(cfg)
	sourceChain := nexustestutils.RandomChain()
	sourceChain.Module = axelarnet.ModuleName
	destinationChain := nexustestutils.RandomChain()
	destinationChain.Module = evmtypes.ModuleName
	k.SetChain(ctx, sourceChain)
	k.SetChain(ctx, destinationChain)
	k.ActivateChain(ctx, sourceChain)
	k.ActivateChain(ctx, destinationChain)
	k.SetMessageRouter(types.NewMessageRouter().AddRoute(destinationChain.Module, func(_ sdk.Context, _ exported.RoutingContext, _ exported.GeneralMessage) error {
		return nil
	}))

	makeSentMessages := func(numMsgs int, destChainName exported.ChainName) map[string]exported.GeneralMessage {
		msgs := make(map[string]exported.GeneralMessage)

		for i := 0; i < numMsgs; i++ {
			destChain := destinationChain
			destChain.Name = destChainName
			id, txID, nonce := k.GenerateMessageID(ctx)
			msg := exported.GeneralMessage{
				ID:            id,
				Sender:        exported.CrossChainAddress{Chain: sourceChain, Address: genCosmosAddr(sourceChain.Name.String())},
				Recipient:     exported.CrossChainAddress{Chain: destChain, Address: evmtestutils.RandomAddress().Hex()},
				Status:        exported.Processing,
				PayloadHash:   crypto.Keccak256Hash(rand.Bytes(int(rand.I64Between(1, 100)))).Bytes(),
				Asset:         nil,
				SourceTxID:    txID,
				SourceTxIndex: nonce,
			}

			msgs[msg.ID] = msg
		}
		return msgs
	}
	enqueueMsgs := func(msgs map[string]exported.GeneralMessage) {
		for _, msg := range msgs {
			status := msg.Status

			msg.Status = exported.Approved
			assert.NoError(t, k.SetNewMessage(ctx, msg))

			switch status {
			case exported.Processing:
				assert.NoError(t, k.RouteMessage(ctx, msg.ID))
			case exported.Executed:
				assert.NoError(t, k.RouteMessage(ctx, msg.ID))
				assert.NoError(t, k.SetMessageExecuted(ctx, msg.ID))
			case exported.Failed:
				assert.NoError(t, k.RouteMessage(ctx, msg.ID))
				assert.NoError(t, k.SetMessageFailed(ctx, msg.ID))
			default:
			}
		}
	}

	toMap := func(msgs []exported.GeneralMessage) map[string]exported.GeneralMessage {

		retMsgs := make(map[string]exported.GeneralMessage)
		for _, msg := range msgs {
			retMsgs[msg.ID] = msg
		}
		return retMsgs
	}
	checkForExistence := func(msgs map[string]exported.GeneralMessage) {
		for _, msg := range msgs {
			retMsg, found := k.GetMessage(ctx, msg.ID)
			assert.True(t, found)
			assert.Equal(t, retMsg, msg)
		}
	}
	consumeSent := func(dest exported.ChainName, limit int64) []exported.GeneralMessage {
		sent := k.GetProcessingMessages(ctx, dest, limit)
		for _, msg := range sent {
			err := k.SetMessageExecuted(ctx, msg.ID)
			assert.NoError(t, err)
		}
		return sent
	}
	destinationChainName := destinationChain.Name
	msgs := makeSentMessages(10, destinationChainName)
	enqueueMsgs(msgs)
	// check msgs can be fetched directly
	checkForExistence(msgs)

	sent := consumeSent(destinationChainName, 100)
	retMsgs := toMap(sent)
	assert.Equal(t, msgs, retMsgs)

	// make sure executed messages are not returned
	sent = k.GetProcessingMessages(ctx, destinationChainName, 100)
	assert.Empty(t, sent)
	for _, msg := range msgs {
		m, found := k.GetMessage(ctx, msg.ID)
		assert.True(t, found)
		msg.Status = exported.Executed
		assert.Equal(t, m, msg)
	}

	// make sure limit works
	msgs = makeSentMessages(100, destinationChainName)
	enqueueMsgs(msgs)
	sent = consumeSent(destinationChainName, 50)
	assert.Equal(t, len(sent), 50)
	sent = append(sent, consumeSent(destinationChainName, 50)...)
	retMsgs = toMap(sent)
	assert.Equal(t, msgs, retMsgs)
	sent = consumeSent(destinationChainName, 10)
	assert.Empty(t, sent)

	// make sure failed messages are not returned
	msgs = makeSentMessages(1, destinationChainName)
	enqueueMsgs(msgs)
	sent = k.GetProcessingMessages(ctx, destinationChainName, 1)
	assert.Equal(t, len(msgs), len(sent))
	err := k.SetMessageFailed(ctx, sent[0].ID)
	assert.NoError(t, err)
	msg := msgs[sent[0].ID]
	msg.Status = exported.Failed
	msgs[msg.ID] = msg
	checkForExistence(msgs)
	assert.Empty(t, consumeSent(destinationChainName, 100))
	checkForExistence(msgs)

	//resend the failed message
	err = k.RouteMessage(ctx, msg.ID)
	assert.NoError(t, err)
	sent = consumeSent(destinationChainName, 1)
	assert.Equal(t, len(sent), 1)
	ret, found := k.GetMessage(ctx, msg.ID)
	assert.True(t, found)
	msg.Status = exported.Executed
	assert.Equal(t, msg, ret)

	// add multiple destinations, make sure routing works
	chain2 := exported.Chain{
		Name:                  exported.ChainName(rand.Str(5)),
		SupportsForeignAssets: true,
		KeyType:               0,
		Module:                "evm",
	}
	k.SetChain(ctx, chain2)
	k.ActivateChain(ctx, chain2)

	chain3 := exported.Chain{
		Name:                  exported.ChainName(rand.Str(5)),
		SupportsForeignAssets: true,
		KeyType:               0,
		Module:                "evm",
	}
	k.SetChain(ctx, chain3)
	k.ActivateChain(ctx, chain3)

	chain4 := exported.Chain{
		Name:                  exported.ChainName(rand.Str(5)),
		SupportsForeignAssets: true,
		KeyType:               0,
		Module:                "evm",
	}
	k.SetChain(ctx, chain4)
	k.ActivateChain(ctx, chain4)

	dest2Msgs := makeSentMessages(10, chain2.Name)
	dest3Msgs := makeSentMessages(10, chain3.Name)
	dest4Msgs := makeSentMessages(10, chain4.Name)

	enqueueMsgs(dest2Msgs)
	enqueueMsgs(dest3Msgs)
	enqueueMsgs(dest4Msgs)
	checkForExistence(dest2Msgs)
	checkForExistence(dest3Msgs)
	checkForExistence(dest4Msgs)
	assert.Equal(t, dest2Msgs, toMap(consumeSent(chain2.Name, 100)))
	assert.Equal(t, dest3Msgs, toMap(consumeSent(chain3.Name, 100)))
	assert.Equal(t, dest4Msgs, toMap(consumeSent(chain4.Name, 100)))
}
