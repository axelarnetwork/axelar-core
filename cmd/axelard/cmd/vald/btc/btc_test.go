package btc

import (
	"fmt"
	"strconv"
	"testing"

	tmEvents "github.com/axelarnetwork/tm-events/events"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/app"
	mock3 "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types/mock"
	mock2 "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/btc/rpc/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestMgr_ProcessConfirmation(t *testing.T) {
	var (
		mgr         *Mgr
		rpc         *mock2.ClientMock
		broadcaster *mock3.BroadcasterMock
		attributes  map[string]string
		info        btc.OutPointInfo
		confHeight  int64
	)

	setup := func() {
		cdc := app.MakeEncodingConfig().Amino
		rpc = &mock2.ClientMock{}
		broadcaster = &mock3.BroadcasterMock{}
		ctx := client.Context{}

		mgr = NewMgr(rpc, ctx, broadcaster, log.TestingLogger(), cdc)

		confHeight = rand.PosI64()
		pollKey := exported.NewPollKey(btc.ModuleName, rand.StrBetween(1, 100))

		info = randomOutpointInfo()
		attributes = map[string]string{
			btc.AttributeKeyConfHeight:   strconv.FormatInt(confHeight, 10),
			btc.AttributeKeyOutPointInfo: string(mgr.cdc.MustMarshalJSON(info)),
			btc.AttributeKeyPoll:         string(mgr.cdc.MustMarshalJSON(pollKey)),
		}
	}

	// Test cases

	repetitionCount := 20
	t.Run("missing attributes", testutils.Func(func(t *testing.T) {
		setup()
		for key := range attributes {
			delete(attributes, key)
			err := mgr.ProcessConfirmation(tmEvents.Event{Attributes: attributes})
			assert.Error(t, err)
			assert.Len(t, broadcaster.BroadcastCalls(), 0)
		}
	}).Repeat(repetitionCount))

	t.Run("RPC unavailable", testutils.Func(func(t *testing.T) {
		setup()
		rpc.GetTxOutFunc = func(*chainhash.Hash, uint32, bool) (*btcjson.GetTxOutResult, error) {
			return nil, fmt.Errorf("some error")
		}

		err := mgr.ProcessConfirmation(tmEvents.Event{Attributes: attributes})
		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		msg := unwrapRefundMsg(broadcaster.BroadcastCalls()[0].Msgs[0])
		assert.False(t, msg.(*btc.VoteConfirmOutpointRequest).Confirmed)
	}).Repeat(repetitionCount))

	t.Run("tx out not found", testutils.Func(func(t *testing.T) {
		setup()
		rpc.GetTxOutFunc = func(*chainhash.Hash, uint32, bool) (*btcjson.GetTxOutResult, error) {
			return nil, nil
		}

		err := mgr.ProcessConfirmation(tmEvents.Event{Attributes: attributes})
		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		msg := unwrapRefundMsg(broadcaster.BroadcastCalls()[0].Msgs[0])
		assert.False(t, msg.(*btc.VoteConfirmOutpointRequest).Confirmed)
	}).Repeat(repetitionCount))

	t.Run("not enough confirmations", testutils.Func(func(t *testing.T) {
		setup()
		rpc.GetTxOutFunc = func(*chainhash.Hash, uint32, bool) (*btcjson.GetTxOutResult, error) {
			return &btcjson.GetTxOutResult{
				Confirmations: rand.I64Between(0, confHeight),
				Value:         info.Amount.ToBTC(),
				ScriptPubKey:  btcjson.ScriptPubKeyResult{Addresses: []string{info.Address}},
			}, nil
		}

		err := mgr.ProcessConfirmation(tmEvents.Event{Attributes: attributes})
		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		msg := unwrapRefundMsg(broadcaster.BroadcastCalls()[0].Msgs[0])
		assert.False(t, msg.(*btc.VoteConfirmOutpointRequest).Confirmed)
	}).Repeat(repetitionCount))

	t.Run("wrong expected data", func(t *testing.T) {
		setup()
		rpc.GetTxOutFunc = func(*chainhash.Hash, uint32, bool) (*btcjson.GetTxOutResult, error) {
			info := randomOutpointInfo()
			return &btcjson.GetTxOutResult{
				Confirmations: rand.PInt64Gen().Where(func(h int64) bool { return h >= confHeight }).Next(),
				Value:         info.Amount.ToBTC(),
				ScriptPubKey:  btcjson.ScriptPubKeyResult{Addresses: []string{info.Address}},
			}, nil
		}

		err := mgr.ProcessConfirmation(tmEvents.Event{Attributes: attributes})
		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		msg := unwrapRefundMsg(broadcaster.BroadcastCalls()[0].Msgs[0])
		assert.False(t, msg.(*btc.VoteConfirmOutpointRequest).Confirmed)
	})

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		rpc.GetTxOutFunc = func(*chainhash.Hash, uint32, bool) (*btcjson.GetTxOutResult, error) {
			return &btcjson.GetTxOutResult{
				Confirmations: rand.PInt64Gen().Where(func(h int64) bool { return h >= confHeight }).Next(),
				Value:         info.Amount.ToBTC(),
				ScriptPubKey:  btcjson.ScriptPubKeyResult{Addresses: []string{info.Address}},
			}, nil
		}

		err := mgr.ProcessConfirmation(tmEvents.Event{Attributes: attributes})
		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		msg := unwrapRefundMsg(broadcaster.BroadcastCalls()[0].Msgs[0])
		assert.True(t, msg.(*btc.VoteConfirmOutpointRequest).Confirmed)
	}).Repeat(repetitionCount))
}

func randomOutpointInfo() btc.OutPointInfo {
	txHash, err := chainhash.NewHash(rand.Bytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}

	voutIdx := uint32(rand.I64Between(0, 100))
	return btc.OutPointInfo{
		OutPoint: wire.NewOutPoint(txHash, voutIdx).String(),
		Amount:   btcutil.Amount(rand.I64Between(1, 10000000)),
		Address:  rand.StrBetween(1, 100),
	}
}

func unwrapRefundMsg(msg sdk.Msg) sdk.Msg {
	return msg.(*axelarnet.RefundMsgRequest).GetInnerMessage()
}
