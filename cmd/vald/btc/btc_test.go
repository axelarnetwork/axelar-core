package btc

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	mock2 "github.com/axelarnetwork/axelar-core/cmd/vald/broadcast/types/mock"
	"github.com/axelarnetwork/axelar-core/cmd/vald/btc/rpc/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestMgr_ProcessConfirmation(t *testing.T) {
	var (
		mgr         *Mgr
		rpc         *mock.ClientMock
		broadcaster *mock2.BroadcasterMock
		attributes  []sdk.Attribute
		info        btc.OutPointInfo
		confHeight  int64
	)

	setup := func() {
		rpc = &mock.ClientMock{}
		broadcaster = &mock2.BroadcasterMock{}
		mgr = NewMgr(rpc, rand.StrBetween(5, 20), broadcaster, nil, log.TestingLogger())

		confHeight = rand.PosI64()
		poll := exported.NewPollMetaWithNonce(
			btc.ModuleName,
			btc.EventTypeOutpointConfirmation,
			rand.StrBetween(1, 100),
			rand.PosI64(),
			rand.PosI64(),
		)

		info = randomOutpointInfo()
		attributes = []sdk.Attribute{
			sdk.NewAttribute(btc.AttributeKeyConfHeight, strconv.FormatInt(confHeight, 10)),
			sdk.NewAttribute(btc.AttributeKeyOutPointInfo, string(testutils.Codec().MustMarshalJSON(info))),
			sdk.NewAttribute(btc.AttributeKeyPoll, string(testutils.Codec().MustMarshalJSON(poll))),
		}
	}

	// Test cases

	repetitionCount := 20
	t.Run("missing attributes", testutils.Func(func(t *testing.T) {
		setup()
		for i := 0; i < len(attributes); i++ {
			// remove one attribute at a time
			wrongAttributes := make([]sdk.Attribute, len(attributes))
			copy(wrongAttributes, attributes)
			wrongAttributes = append(wrongAttributes[:i], wrongAttributes[(i+1):]...)

			err := mgr.ProcessConfirmation(wrongAttributes)
			assert.Error(t, err)
		}
	}).Repeat(repetitionCount))

	t.Run("RPC unavailable", testutils.Func(func(t *testing.T) {
		setup()
		rpc.GetTxOutFunc = func(*chainhash.Hash, uint32, bool) (*btcjson.GetTxOutResult, error) {
			return nil, fmt.Errorf("some error")
		}

		err := mgr.ProcessConfirmation(attributes)
		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(btc.MsgVoteConfirmOutpoint).Confirmed)
	}).Repeat(repetitionCount))

	t.Run("tx out not found", testutils.Func(func(t *testing.T) {
		setup()
		rpc.GetTxOutFunc = func(*chainhash.Hash, uint32, bool) (*btcjson.GetTxOutResult, error) {
			return nil, nil
		}

		err := mgr.ProcessConfirmation(attributes)
		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(btc.MsgVoteConfirmOutpoint).Confirmed)
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

		err := mgr.ProcessConfirmation(attributes)
		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(btc.MsgVoteConfirmOutpoint).Confirmed)
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

		err := mgr.ProcessConfirmation(attributes)
		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(btc.MsgVoteConfirmOutpoint).Confirmed)
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

		err := mgr.ProcessConfirmation(attributes)
		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.True(t, broadcaster.BroadcastCalls()[0].Msgs[0].(btc.MsgVoteConfirmOutpoint).Confirmed)
	}).Repeat(repetitionCount))
}

func randomOutpointInfo() btc.OutPointInfo {
	txHash, err := chainhash.NewHash(rand.Bytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}

	voutIdx := uint32(rand.I64Between(0, 100))
	return btc.OutPointInfo{
		OutPoint: wire.NewOutPoint(txHash, voutIdx),
		Amount:   btcutil.Amount(rand.I64Between(1, 10000000)),
		Address:  rand.StrBetween(1, 100),
	}
}
