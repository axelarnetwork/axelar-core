package broadcast

import (
	"fmt"
	rand2 "math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/core/types"

	"github.com/axelarnetwork/axelar-core/cmd/vald/broadcast/types/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

func TestBroadcaster_Broadcast(t *testing.T) {
	t.Run("called sequentially", func(t *testing.T) {
		b, rpc := setup()

		expectedSigs := 0
		var senders []sdk.AccAddress
		iterations := int(rand.I64Between(20, 100))
		for i := 0; i < iterations; i++ {
			msgs := createMsgsWithRandomSigners()
			sender := msgs[0].GetSigners()[0]

			err := b.Broadcast(msgs...)
			assert.NoError(t, err)

			for _, msg := range msgs {
				expectedSigs += len(msg.GetSigners())
			}
			senders = append(senders, sender)
		}

		// sign called for every signer
		sigCount := 0
		for _, calls := range rpc.BroadcastTxSyncCalls() {
			sigCount += len(calls.Tx.Signatures)
		}
		assert.Equal(t, expectedSigs, sigCount)

		// broadcast from correct account
		assert.Len(t, rpc.GetAccountNumberSequenceCalls(), iterations)
		for i, sender := range senders {
			assert.Equal(t, sender, rpc.GetAccountNumberSequenceCalls()[i].Addr)
			assert.Equal(t, sender, rpc.BroadcastTxSyncCalls()[i].Tx.FeePayer())
		}
	})

	t.Run("called concurrently", func(t *testing.T) {
		b, rpc := setup()

		iterations := int(rand.I64Between(20, 100))

		// if the call to broadcast is not correctly sequenced the callCounter should be lower than the actual call count (data race)
		callCounter := 0
		rpc.BroadcastTxSyncFunc = func(tx authtypes.StdTx) (*coretypes.ResultBroadcastTx, error) {
			c := callCounter

			// simulate blocking
			timeout := time.Duration(rand.I64Between(0, 20)) * time.Millisecond
			time.Sleep(timeout)

			c++
			callCounter = c
			return &coretypes.ResultBroadcastTx{Code: abci.CodeTypeOK}, nil
		}
		wg := &sync.WaitGroup{}
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func() {
				defer wg.Done()
				msgs := createMsgsWithRandomSigners()
				err := b.Broadcast(msgs...)
				assert.NoError(t, err)
			}()
		}
		wg.Wait()
		assert.Equal(t, len(rpc.BroadcastTxSyncCalls()), callCounter)
	})

	t.Run("sequence number updated correctly", func(t *testing.T) {
		accNo := rand2.Uint64()
		seqNo := uint64(1)
		prevSeqNo := uint64(0)
		rpc := &mock.ClientMock{
			GetAccountNumberSequenceFunc: func(sdk.AccAddress) (uint64, uint64, error) {
				return accNo, seqNo, nil
			},
			BroadcastTxSyncFunc: func(tx auth.StdTx) (*coretypes.ResultBroadcastTx, error) {
				seqNo++
				return &coretypes.ResultBroadcastTx{Code: abci.CodeTypeOK}, nil
			}}
		config := types.ClientConfig{
			ChainID:         rand.StrBetween(5, 20),
			BroadcastConfig: types.BroadcastConfig{},
		}

		seen := map[string]bool{}
		s := func(from sdk.AccAddress, msg auth.StdSignMsg) (authtypes.StdSignature, error) {
			bz := string(msg.Bytes())
			if !seen[bz] {
				assert.Equal(t, prevSeqNo+1, msg.Sequence)
				atomic.StoreUint64(&prevSeqNo, msg.Sequence)
				seen[bz] = true
			}

			return authtypes.StdSignature{Signature: rand.Bytes(int(rand.I64Between(5, 100)))}, nil
		}

		b, err := NewBroadcaster(s, rpc, config, log.TestingLogger())
		if err != nil {
			panic(err)
		}

		iterations := int(rand.I64Between(200, 1000))
		wg := &sync.WaitGroup{}
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func(broadcaster *Broadcaster) {
				defer wg.Done()
				msgs := createMsgsWithRandomSigners()
				err := broadcaster.Broadcast(msgs...)
				assert.NoError(t, err)
			}(b)
		}
		wg.Wait()
	})

	t.Run("sequence number on blockchain trailing behind", func(t *testing.T) {
		accNo := rand2.Uint64()
		seqNo := uint64(1)
		prevSeqNo := uint64(0)
		rpc := &mock.ClientMock{
			GetAccountNumberSequenceFunc: func(sdk.AccAddress) (uint64, uint64, error) {
				return accNo, seqNo, nil
			},
			BroadcastTxSyncFunc: func(tx auth.StdTx) (*coretypes.ResultBroadcastTx, error) {
				return &coretypes.ResultBroadcastTx{Code: abci.CodeTypeOK}, nil
			}}
		config := types.ClientConfig{
			ChainID:         rand.StrBetween(5, 20),
			BroadcastConfig: types.BroadcastConfig{},
		}

		seen := map[string]bool{}
		s := func(from sdk.AccAddress, msg auth.StdSignMsg) (authtypes.StdSignature, error) {
			bz := string(msg.Bytes())
			if !seen[bz] {
				assert.Equal(t, prevSeqNo+1, msg.Sequence)
				atomic.StoreUint64(&prevSeqNo, msg.Sequence)
				seen[bz] = true
			}

			return authtypes.StdSignature{Signature: rand.Bytes(int(rand.I64Between(5, 100)))}, nil
		}

		b, err := NewBroadcaster(s, rpc, config, log.TestingLogger())
		if err != nil {
			panic(err)
		}

		iterations := int(rand.I64Between(200, 1000))
		wg := &sync.WaitGroup{}
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func(broadcaster *Broadcaster) {
				defer wg.Done()
				msgs := createMsgsWithRandomSigners()
				err := broadcaster.Broadcast(msgs...)
				assert.NoError(t, err)
			}(b)
		}
		wg.Wait()
	})
}

func TestXBOBroadcaster_Broadcast(t *testing.T) {
	t.Run("failed broadcast with exponential backoff", func(t *testing.T) {
		b, rpc := setup()
		retries := int(rand.I64Between(1, 20))
		xbo := WithExponentialBackoff(b, 20*time.Microsecond, retries)

		rpc.BroadcastTxSyncFunc = func(authtypes.StdTx) (*coretypes.ResultBroadcastTx, error) {
			return nil, fmt.Errorf("some error")
		}

		iterations := int(rand.I64Between(20, 100))

		wg := &sync.WaitGroup{}
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func() {
				defer wg.Done()
				msgs := createMsgsWithRandomSigners()
				err := xbo.Broadcast(msgs...)
				assert.Error(t, err)
				t.Log(err)
			}()
		}
		wg.Wait()

		assert.Len(t, rpc.BroadcastTxSyncCalls(), iterations*(retries+1))
	})

	t.Run("sequence number updated correctly", func(t *testing.T) {
		accNo := rand2.Uint64()
		seqNo := uint64(1)
		prevSeqNo := uint64(0)
		rpc := &mock.ClientMock{
			GetAccountNumberSequenceFunc: func(sdk.AccAddress) (uint64, uint64, error) {
				return accNo, seqNo, nil
			},
			BroadcastTxSyncFunc: func(tx auth.StdTx) (*coretypes.ResultBroadcastTx, error) {
				seqNo++
				return &coretypes.ResultBroadcastTx{Code: abci.CodeTypeOK}, nil
			}}
		config := types.ClientConfig{
			ChainID:         rand.StrBetween(5, 20),
			BroadcastConfig: types.BroadcastConfig{},
		}

		seen := map[string]bool{}
		s := func(from sdk.AccAddress, msg auth.StdSignMsg) (authtypes.StdSignature, error) {
			bz := string(msg.Bytes())
			if !seen[bz] {
				assert.Equal(t, prevSeqNo+1, msg.Sequence)
				atomic.StoreUint64(&prevSeqNo, msg.Sequence)
				seen[bz] = true
			}

			return authtypes.StdSignature{Signature: rand.Bytes(int(rand.I64Between(5, 100)))}, nil
		}

		b, err := NewBroadcaster(s, rpc, config, log.TestingLogger())
		if err != nil {
			panic(err)
		}
		retries := int(rand.I64Between(1, 20))
		xbo := WithExponentialBackoff(b, 20*time.Microsecond, retries)

		iterations := int(rand.I64Between(200, 1000))
		wg := &sync.WaitGroup{}
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func(broadcaster *XBOBroadcaster) {
				defer wg.Done()
				msgs := createMsgsWithRandomSigners()
				assert.NoError(t, broadcaster.Broadcast(msgs...))
			}(xbo)
		}
		wg.Wait()
	})
}

func setup() (*Broadcaster, *mock.ClientMock) {
	rpc := &mock.ClientMock{
		GetAccountNumberSequenceFunc: func(sdk.AccAddress) (uint64, uint64, error) {
			return rand2.Uint64(), rand2.Uint64(), nil
		},
		BroadcastTxSyncFunc: func(tx auth.StdTx) (*coretypes.ResultBroadcastTx, error) {
			return &coretypes.ResultBroadcastTx{Code: abci.CodeTypeOK}, nil
		}}
	config := types.ClientConfig{
		ChainID:         rand.StrBetween(5, 20),
		BroadcastConfig: types.BroadcastConfig{},
	}
	s := func(from sdk.AccAddress, msg auth.StdSignMsg) (authtypes.StdSignature, error) {
		return authtypes.StdSignature{Signature: rand.Bytes(int(rand.I64Between(5, 100)))}, nil
	}

	b, err := NewBroadcaster(s, rpc, config, log.TestingLogger())
	if err != nil {
		panic(err)
	}
	return b, rpc
}

func createMsgsWithRandomSigners() []sdk.Msg {
	var msgs []sdk.Msg
	for i := int64(0); i < rand.I64Between(1, 20); i++ {
		var signers []sdk.AccAddress
		for j := int64(0); j < rand.I64Between(1, 5); j++ {
			signers = append(signers, sdk.AccAddress(rand.Str(sdk.AddrLen)))
		}
		msg := &mock.MsgMock{GetSignersFunc: func() []sdk.AccAddress { return signers }}
		msgs = append(msgs, msg)
	}
	return msgs
}
