package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

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
