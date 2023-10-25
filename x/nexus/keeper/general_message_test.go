package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
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

func TestSetNewMessage_(t *testing.T) {
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
			keeper.SetNewMessage_(ctx, msg)
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, keeper.SetNewMessage_(ctx, msg), "already exists")
		}).
		Run(t)

	givenKeeper.
		When("the message has invalid status", func() {
			msg = randMsg(exported.Processing)
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, keeper.SetNewMessage_(ctx, msg), "new general message has to be approved")
		}).
		Run(t)

	givenKeeper.
		When("the message is valid", func() {
			msg = randMsg(exported.Approved)
		}).
		Then("should store the message", func(t *testing.T) {
			assert.NoError(t, keeper.SetNewMessage_(ctx, msg))

			actual, ok := keeper.GetMessage(ctx, msg.ID)
			assert.True(t, ok)
			assert.Equal(t, msg, actual)
		}).
		Run(t)
}

func TestSetMessageProcessing_(t *testing.T) {
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
		When("the message doesn't exist", func() {}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, rand.NormalizedStr(10)), "not found")
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

			keeper.SetNewMessage_(ctx, msg)
			keeper.SetMessageProcessing_(ctx, msg.ID)
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "general message has to be approved or failed")
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

				keeper.SetNewMessage_(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "is not registered")
				}),

			When("the destination chain is not activated", func() {
				msg.Recipient = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}

				keeper.DeactivateChain(ctx, msg.Recipient.Chain)
				keeper.SetNewMessage_(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "is not activated")
				}),

			When("the destination address is invalid", func() {
				msg.Recipient = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: rand.NormalizedStr(42),
				}

				keeper.SetNewMessage_(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "not an hex address")
				}),

			When("the destination chain does't support the asset", func() {
				msg.Recipient = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}
				asset := rand.Coin()
				msg.Asset = &asset

				keeper.SetNewMessage_(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "does not support foreign asset")
				}),

			When("asset is set", func() {
				msg.Recipient = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}
				msg.Asset = &sdk.Coin{Denom: "external-erc-20", Amount: sdk.NewInt(100)}

				keeper.SetNewMessage_(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "asset transfer is not supported for wasm messages")
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

				keeper.SetNewMessage_(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "is not registered")
				}),

			When("the sender chain is not activated", func() {
				msg.Sender = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}

				keeper.DeactivateChain(ctx, msg.Sender.Chain)
				keeper.SetNewMessage_(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "is not activated")
				}),

			When("the sender address is invalid", func() {
				msg.Sender = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: rand.NormalizedStr(42),
				}

				keeper.SetNewMessage_(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "not an hex address")
				}),

			When("the sender chain does't support the asset", func() {
				msg.Sender = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}
				asset := rand.Coin()
				msg.Asset = &asset

				keeper.SetNewMessage_(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "does not support foreign asset")
				}),

			When("asset is set", func() {
				msg.Sender = exported.CrossChainAddress{
					Chain:   evm.Ethereum,
					Address: evmtestutils.RandomAddress().Hex(),
				}
				msg.Asset = &sdk.Coin{Denom: "external-erc-20", Amount: sdk.NewInt(100)}

				keeper.SetNewMessage_(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetMessageProcessing_(ctx, msg.ID), "asset transfer is not supported for wasm messages")
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

			keeper.SetNewMessage_(ctx, msg)
		}).
		Then("should set the message status to processing", func(t *testing.T) {
			assert.NoError(t, keeper.SetMessageProcessing_(ctx, msg.ID))

			actual, ok := keeper.GetMessage(ctx, msg.ID)
			assert.True(t, ok)
			assert.Equal(t, exported.Processing, actual.Status)
		}).
		Run(t)
}

func randWasmMsg(status exported.GeneralMessage_Status) exported.GeneralMessage {
	return exported.GeneralMessage{
		ID: rand.NormalizedStr(10),
		Sender: exported.CrossChainAddress{
			Chain:   nexustestutils.RandomChain(),
			Address: rand.NormalizedStr(42),
		},
		Recipient: exported.CrossChainAddress{
			Chain:   evm.Ethereum,
			Address: evmtestutils.RandomAddress().Hex(),
		},
		PayloadHash:   evmtestutils.RandomHash().Bytes(),
		Status:        status,
		Asset:         nil,
		SourceTxID:    evmtestutils.RandomHash().Bytes(),
		SourceTxIndex: uint64(rand.I64Between(0, 100)),
	}
}

func TestSetNewWasmMessage(t *testing.T) {
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
		When("the message is valid", func() {
			msg = randWasmMsg(exported.Approved)
		}).
		Branch(
			When("the message contains token transfer", func() {
				coin := rand.Coin()
				msg.Asset = &coin
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetNewWasmMessage(ctx, msg), "asset transfer is not supported")
				}),

			When("the destination chain is not registered", func() {
				msg.Recipient.Chain = nexustestutils.RandomChain()
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetNewWasmMessage(ctx, msg), "is not a registered chain")
				}),

			When("the destination chain is not activated", func() {
				keeper.DeactivateChain(ctx, msg.Recipient.Chain)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetNewWasmMessage(ctx, msg), "is not activated")
				}),

			When("the recipient address is invalid", func() {
				msg.Recipient.Address = rand.Str(20)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetNewWasmMessage(ctx, msg), "invalid recipient address")
				}),

			When("the message already exists", func() {
				keeper.SetNewWasmMessage(ctx, msg)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetNewWasmMessage(ctx, msg), "already exists")
				}),

			When("the message is invalid", func() {
				msg.Sender.Address = ""
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetNewWasmMessage(ctx, msg), "invalid source chain: invalid address: string is empty")
				}),

			When("the message status is invalid", func() {
				msg.Status = exported.Failed
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, keeper.SetNewWasmMessage(ctx, msg), "invalid message status")
				}),
		).
		Run(t)

	givenKeeper.
		Branch(
			When("the message status is approved", func() {
				msg = randWasmMsg(exported.Approved)
			}).
				Then("should be stored as approved and emit MessageReceived event", func(t *testing.T) {
					assert.NoError(t, keeper.SetNewWasmMessage(ctx, msg))

					actual, ok := keeper.GetMessage(ctx, msg.ID)
					assert.True(t, ok)
					assert.Equal(t, msg, actual)
					assert.Equal(t, "axelar.nexus.v1beta1.MessageReceived", ctx.EventManager().Events()[len(ctx.EventManager().Events())-1].Type)
				}),

			When("the message status is processing", func() {
				msg = randWasmMsg(exported.Processing)
			}).
				Then("should be stored as processing and emit MessageProcessing event", func(t *testing.T) {
					assert.NoError(t, keeper.SetNewWasmMessage(ctx, msg))

					actual, ok := keeper.GetMessage(ctx, msg.ID)
					assert.True(t, ok)
					assert.Equal(t, msg, actual)
					assert.Equal(t, "axelar.nexus.v1beta1.MessageProcessing", ctx.EventManager().Events()[len(ctx.EventManager().Events())-1].Type)
					assert.Equal(t, msg, keeper.GetProcessingMessages(ctx, msg.GetDestinationChain(), 1)[0])
				}),
		).
		Run(t)
}

func TestSetNewGeneralMessage(t *testing.T) {
	var (
		generalMessage exported.GeneralMessage
		ctx            sdk.Context
		k              nexus.Keeper
	)
	cfg := app.MakeEncodingConfig()
	sourceChain := nexustestutils.RandomChain()
	sourceChain.Module = evmtypes.ModuleName
	destinationChain := nexustestutils.RandomChain()
	asset := rand.Coin()

	givenContractCallEvent := Given("a general message with token", func() {
		k, ctx = setup(cfg)
		generalMessage = exported.GeneralMessage{
			ID: fmt.Sprintf("%s-%d", evmtestutils.RandomHash().Hex(), rand.PosI64()),

			Sender: exported.CrossChainAddress{
				Chain:   sourceChain,
				Address: evmtestutils.RandomAddress().Hex(),
			},
			Recipient: exported.CrossChainAddress{
				Chain:   destinationChain,
				Address: genCosmosAddr(destinationChain.Name.String()),
			},
			Status:      exported.Approved,
			PayloadHash: crypto.Keccak256Hash(rand.Bytes(int(rand.I64Between(1, 100)))).Bytes(),
			Asset:       &asset,
		}
	})

	whenChainsAreRegistered := givenContractCallEvent.
		When("the source and destination chains are registered", func() {
			k.SetChain(ctx, sourceChain)
			k.SetChain(ctx, destinationChain)
		})

	errorWith := func(msg string) func(t *testing.T) {
		return func(t *testing.T) {
			assert.ErrorContains(t, k.SetNewMessage(ctx, generalMessage), msg)
		}
	}

	isCosmosChain := func(isCosmosChain bool) func() {
		return func() {
			if isCosmosChain {
				destChain := funcs.MustOk(k.GetChain(ctx, destinationChain.Name))
				destChain.Module = axelarnet.ModuleName
				k.SetChain(ctx, destChain)

				generalMessage.Recipient.Chain.Module = axelarnet.ModuleName
			}
		}
	}

	isAssetRegistered := func(isRegistered bool) func() {
		return func() {
			if isRegistered {
				funcs.MustNoErr(k.RegisterAsset(ctx, sourceChain, exported.Asset{Denom: asset.Denom, IsNativeAsset: false}, utils.MaxUint, time.Hour))
				funcs.MustNoErr(k.RegisterAsset(ctx, destinationChain, exported.Asset{Denom: asset.Denom, IsNativeAsset: false}, utils.MaxUint, time.Hour))
			}
		}
	}

	givenContractCallEvent.
		When("the source chain is not registered", func() {}).
		Then("should return error", errorWith(fmt.Sprintf("source chain %s is not a registered chain", sourceChain.Name))).
		Run(t)

	givenContractCallEvent.
		When("the destination chain is not registered", func() {
			k.SetChain(ctx, sourceChain)
		}).
		Then("should return error", errorWith(fmt.Sprintf("destination chain %s is not a registered chain", destinationChain.Name))).
		Run(t)

	whenChainsAreRegistered.
		When("address validator for destination chain is set", isCosmosChain(true)).
		When("destination address is invalid", func() {
			generalMessage.Recipient.Address = rand.Str(20)
		}).
		Then("should return error", errorWith("decoding bech32 failed")).
		Run(t)

	whenChainsAreRegistered.
		When("address validator for destination chain is set", isCosmosChain(true)).
		When("asset is not registered", isAssetRegistered(false)).
		Then("should return error", errorWith("does not support foreign asset")).
		Run(t)

	whenChainsAreRegistered.
		When("address validator for destination chain is set", isCosmosChain(true)).
		When("asset is registered", isAssetRegistered(true)).
		Then("should succeed", func(t *testing.T) {
			assert.NoError(t, k.SetNewMessage(ctx, generalMessage))
		}).
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

	// Message doesn't exist, can't set any status
	err := k.SetMessageFailed(ctx, msg.ID)
	assert.Error(t, err, fmt.Sprintf("general message %s not found", msg.ID))

	err = k.SetMessageProcessing(ctx, msg.ID)
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

	err = k.SetMessageProcessing(ctx, msg.ID)
	assert.NoError(t, err)

	err = k.SetMessageProcessing(ctx, msg.ID)
	assert.Error(t, err, "general message is not approved or failed")

	err = k.SetMessageFailed(ctx, msg.ID)
	assert.NoError(t, err)

	err = k.SetMessageExecuted(ctx, msg.ID)
	assert.Error(t, err, "general message is not processed")

	err = k.SetMessageProcessing(ctx, msg.ID)
	assert.NoError(t, err)

	err = k.SetMessageExecuted(ctx, msg.ID)
	assert.NoError(t, err)

	err = k.SetMessageFailed(ctx, msg.ID)
	assert.Error(t, err, "general message is not processed")

	err = k.SetMessageProcessing(ctx, msg.ID)
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
	k.SetChain(ctx, sourceChain)
	k.SetChain(ctx, destinationChain)

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
			err := k.SetNewMessage(ctx, msg)
			assert.NoError(t, err)
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
	err = k.SetMessageProcessing(ctx, msg.ID)
	assert.NoError(t, err)
	sent = consumeSent(destinationChainName, 1)
	assert.Equal(t, len(sent), 1)
	ret, found := k.GetMessage(ctx, msg.ID)
	assert.True(t, found)
	msg.Status = exported.Executed
	assert.Equal(t, msg, ret)

	// add multiple destinations, make sure routing works
	dest2 := exported.ChainName(rand.Str(5))
	k.SetChain(ctx, exported.Chain{
		Name:                  dest2,
		SupportsForeignAssets: true,
		KeyType:               0,
		Module:                "evm",
	})
	dest3 := exported.ChainName(rand.Str(5))
	k.SetChain(ctx, exported.Chain{
		Name:                  dest3,
		SupportsForeignAssets: true,
		KeyType:               0,
		Module:                "evm",
	})
	dest4 := exported.ChainName(rand.Str(5))
	k.SetChain(ctx, exported.Chain{
		Name:                  dest4,
		SupportsForeignAssets: true,
		KeyType:               0,
		Module:                "evm",
	})

	dest2Msgs := makeSentMessages(10, dest2)
	dest3Msgs := makeSentMessages(10, dest3)
	dest4Msgs := makeSentMessages(10, dest4)

	enqueueMsgs(dest2Msgs)
	enqueueMsgs(dest3Msgs)
	enqueueMsgs(dest4Msgs)
	checkForExistence(dest2Msgs)
	checkForExistence(dest3Msgs)
	checkForExistence(dest4Msgs)
	assert.Equal(t, dest2Msgs, toMap(consumeSent(dest2, 100)))
	assert.Equal(t, dest3Msgs, toMap(consumeSent(dest3, 100)))
	assert.Equal(t, dest4Msgs, toMap(consumeSent(dest4, 100)))
}
