package keeper_test

import (
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
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
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
	sourceChainName := exported.ChainName(rand.Str(5))
	destinationChainName := exported.ChainName(rand.Str(5))
	asset := rand.Coin()

	givenContractCallEvent := Given("a general message with token", func() {
		k, ctx = setup(cfg)
		generalMessage = exported.GeneralMessage{
			ID: exported.MessageID{
				ID:    fmt.Sprintf("%s-%d", evmtestutils.RandomHash().Hex(), rand.PosI64()),
				Chain: destinationChainName,
			},
			SourceChain: sourceChainName,
			Sender:      evmtestutils.RandomAddress().Hex(),
			Receiver:    genCosmosAddr(destinationChainName.String()),
			Status:      exported.Approved,
			PayloadHash: crypto.Keccak256Hash(rand.Bytes(int(rand.I64Between(1, 100)))).Bytes(),
			Asset:       &asset,
		}

	})

	whenChainsAreRegistered := givenContractCallEvent.
		When("the source and destination chains are registered", func() {
			k.SetChain(ctx, exported.Chain{Name: sourceChainName, SupportsForeignAssets: true})
			k.SetChain(ctx, exported.Chain{Name: destinationChainName, SupportsForeignAssets: true})
		})

	errorWith := func(msg string) func(t *testing.T) {
		return func(t *testing.T) {
			assert.ErrorContains(t, k.SetNewMessage(ctx, generalMessage), msg)
		}
	}

	isCosmosChain := func(isCosmosChain bool) func() {
		return func() {
			if isCosmosChain {
				destChain := funcs.MustOk(k.GetChain(ctx, destinationChainName))
				destChain.Module = axelarnet.ModuleName
				k.SetChain(ctx, destChain)
			}
		}
	}

	isAssetRegistered := func(isRegistered bool) func() {
		return func() {
			if isRegistered {
				srcChain := funcs.MustOk(k.GetChain(ctx, sourceChainName))
				destChain := funcs.MustOk(k.GetChain(ctx, destinationChainName))
				funcs.MustNoErr(k.RegisterAsset(ctx, srcChain, exported.Asset{Denom: asset.Denom, IsNativeAsset: false}, utils.MaxUint, time.Hour))
				funcs.MustNoErr(k.RegisterAsset(ctx, destChain, exported.Asset{Denom: asset.Denom, IsNativeAsset: false}, utils.MaxUint, time.Hour))
			}
		}
	}

	givenContractCallEvent.
		When("the source chain is not registered", func() {}).
		Then("should return error", errorWith(fmt.Sprintf("source chain %s is not a registered chain", sourceChainName))).
		Run(t)

	givenContractCallEvent.
		When("the destination chain is not registered", func() {
			k.SetChain(ctx, exported.Chain{Name: sourceChainName})
		}).
		Then("should return error", errorWith(fmt.Sprintf("destination chain %s is not a registered chain", destinationChainName))).
		Run(t)

	whenChainsAreRegistered.
		When("address validator for destination chain is set", isCosmosChain(true)).
		When("destination address is invalid", func() {
			generalMessage.Receiver = rand.Str(20)
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

func TestGetGeneralMessageID(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	k, ctx := setup(cfg)

	hash := evmtestutils.RandomHash().Hex()
	chain := exported.ChainName(rand.Str(5))
	// use the same hash and source chain, still shouldn't collide
	id := k.GetGeneralMessageID(ctx, hash, chain)
	id2 := k.GetGeneralMessageID(ctx, hash, chain)
	assert.NotEqual(t, id, id2)
}

func TestSetMessageFailed(t *testing.T) {

	cfg := app.MakeEncodingConfig()
	k, ctx := setup(cfg)
	srcChain := rand.Str(5)
	msg := exported.GeneralMessage{
		ID:          exported.MessageID{ID: k.GetGeneralMessageID(ctx, evmtestutils.RandomHash().Hex(), exported.ChainName(srcChain)), Chain: exported.ChainName(rand.Str(5))},
		SourceChain: exported.ChainName(srcChain),
		Sender:      genCosmosAddr(srcChain),
		Receiver:    evmtestutils.RandomAddress().Hex(),
		Status:      exported.Failed,
		PayloadHash: crypto.Keccak256Hash(rand.Bytes(int(rand.I64Between(1, 100)))).Bytes(),
		Asset:       nil,
	}
	k.SetChain(ctx, exported.Chain{Name: msg.ID.Chain, SupportsForeignAssets: true, Module: "evm"})
	k.SetChain(ctx, exported.Chain{Name: msg.SourceChain, SupportsForeignAssets: true, Module: "axelarnet"})
	err := k.SetMessageFailed(ctx, msg.ID)
	assert.Error(t, err, fmt.Sprintf("general message %s not found", msg.ID.String()))

	err = k.SetNewMessage(ctx, msg)
	assert.NoError(t, err)
	err = k.SetMessageFailed(ctx, msg.ID)
	assert.Error(t, err, "general message is not sent or approved")
	err = k.SetMessageSent(ctx, msg.ID)
	assert.NoError(t, err)
	err = k.SetMessageFailed(ctx, msg.ID)
	assert.NoError(t, err)
}

func TestGetMessage(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	k, ctx := setup(cfg)

	srcChain := rand.Str(5)
	msg := exported.GeneralMessage{
		ID:          exported.MessageID{ID: k.GetGeneralMessageID(ctx, evmtestutils.RandomHash().Hex(), exported.ChainName(srcChain)), Chain: exported.ChainName(rand.Str(5))},
		SourceChain: exported.ChainName(srcChain),
		Sender:      genCosmosAddr(srcChain),
		Receiver:    evmtestutils.RandomAddress().Hex(),
		Status:      exported.Approved,
		PayloadHash: crypto.Keccak256Hash(rand.Bytes(int(rand.I64Between(1, 100)))).Bytes(),
		Asset:       nil,
	}
	k.SetChain(ctx, exported.Chain{Name: msg.ID.Chain, SupportsForeignAssets: true, Module: "evm"})
	k.SetChain(ctx, exported.Chain{Name: msg.SourceChain, SupportsForeignAssets: true, Module: "axelarnet"})

	err := k.SetNewMessage(ctx, msg)
	assert.NoError(t, err)

	exp, found := k.GetMessageWithStatus(ctx, msg.ID, []exported.GeneralMessage_Status{exported.Approved})
	assert.True(t, found)
	assert.Equal(t, exp, msg)

	exp, found = k.GetMessageWithStatus(ctx, msg.ID, []exported.GeneralMessage_Status{exported.Failed, exported.Approved})
	assert.True(t, found)
	assert.Equal(t, exp, msg)

	exp, found = k.GetMessageWithStatus(ctx, msg.ID, []exported.GeneralMessage_Status{exported.Approved, exported.Failed})
	assert.True(t, found)
	assert.Equal(t, exp, msg)

	_, found = k.GetMessageWithStatus(ctx, msg.ID, []exported.GeneralMessage_Status{exported.Failed})
	assert.False(t, found)
	_, found = k.GetMessageWithStatus(ctx, msg.ID, []exported.GeneralMessage_Status{exported.Sent})
	assert.False(t, found)
	_, found = k.GetMessageWithStatus(ctx, msg.ID, []exported.GeneralMessage_Status{exported.Executed})
	assert.False(t, found)

	exp, found = k.GetMessageAnyStatus(ctx, msg.ID)
	assert.True(t, found)
	assert.Equal(t, exp, msg)
}

func TestConsumeMessages(t *testing.T) {

	cfg := app.MakeEncodingConfig()
	k, ctx := setup(cfg)
	sourceChainName := exported.ChainName(rand.Str(5))
	destinationChainName := exported.ChainName(rand.Str(5))
	k.SetChain(ctx, exported.Chain{
		Name:                  sourceChainName,
		SupportsForeignAssets: true,
		KeyType:               0,
		Module:                axelarnet.ModuleName,
	})
	k.SetChain(ctx, exported.Chain{
		Name:                  destinationChainName,
		SupportsForeignAssets: true,
		KeyType:               0,
		Module:                "evm",
	})
	makeMessages := func(numMsgs int, destChain exported.ChainName) map[string]exported.GeneralMessage {

		msgs := make(map[string]exported.GeneralMessage)

		for i := 0; i < numMsgs; i++ {

			msg := exported.GeneralMessage{
				ID:          exported.MessageID{ID: k.GetGeneralMessageID(ctx, evmtestutils.RandomHash().Hex(), sourceChainName), Chain: destChain},
				SourceChain: sourceChainName,
				Sender:      genCosmosAddr(destinationChainName.String()),
				Receiver:    evmtestutils.RandomAddress().Hex(),
				Status:      exported.Approved,
				PayloadHash: crypto.Keccak256Hash(rand.Bytes(int(rand.I64Between(1, 100)))).Bytes(),
				Asset:       nil,
			}
			msgs[msg.ID.ID] = msg
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
			retMsgs[msg.ID.ID] = msg
		}
		return retMsgs
	}
	checkForExistence := func(msgs map[string]exported.GeneralMessage) {
		for _, msg := range msgs {
			retMsg, found := k.GetMessageAnyStatus(ctx, msg.ID)
			assert.True(t, found)
			assert.Equal(t, retMsg, msg)
		}
	}
	msgs := makeMessages(10, destinationChainName)
	enqueueMsgs(msgs)
	// check msgs can be fetched directly
	checkForExistence(msgs)

	consumed := k.ConsumeApprovedMessages(ctx, destinationChainName, 100)
	retMsgs := toMap(consumed)
	assert.Equal(t, msgs, retMsgs)

	// check msgs are deleted
	consumed = k.ConsumeApprovedMessages(ctx, destinationChainName, 100)
	assert.Empty(t, consumed)
	for _, msg := range msgs {
		_, found := k.GetMessageAnyStatus(ctx, msg.ID)
		assert.False(t, found)
	}

	// make sure limit works
	msgs = makeMessages(100, destinationChainName)
	enqueueMsgs(msgs)
	consumed = k.ConsumeApprovedMessages(ctx, destinationChainName, 50)
	assert.Equal(t, len(consumed), 50)
	consumed = append(consumed, k.ConsumeApprovedMessages(ctx, destinationChainName, 50)...)
	retMsgs = toMap(consumed)
	assert.Equal(t, msgs, retMsgs)
	consumed = k.ConsumeApprovedMessages(ctx, destinationChainName, 10)
	assert.Empty(t, consumed)

	// make sure SetFailed does not enqueue
	msgs = makeMessages(1, destinationChainName)
	for id, m := range msgs {
		m.Status = exported.Sent
		msgs[id] = m
		err := k.SetNewMessage(ctx, m)
		assert.NoError(t, err)
		err = k.SetMessageFailed(ctx, m.ID)
		assert.NoError(t, err)
	}
	checkForExistence(msgs)
	assert.Empty(t, k.ConsumeApprovedMessages(ctx, destinationChainName, 100))
	checkForExistence(msgs)
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

	dest2Msgs := makeMessages(10, dest2)
	dest3Msgs := makeMessages(10, dest3)
	dest4Msgs := makeMessages(10, dest4)

	enqueueMsgs(dest2Msgs)
	enqueueMsgs(dest3Msgs)
	enqueueMsgs(dest4Msgs)
	checkForExistence(dest2Msgs)
	checkForExistence(dest3Msgs)
	checkForExistence(dest4Msgs)
	assert.Equal(t, dest2Msgs, toMap(k.ConsumeApprovedMessages(ctx, dest2, 100)))
	assert.Equal(t, dest3Msgs, toMap(k.ConsumeApprovedMessages(ctx, dest3, 100)))
	assert.Equal(t, dest4Msgs, toMap(k.ConsumeApprovedMessages(ctx, dest4, 100)))

}
