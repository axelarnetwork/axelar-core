package broadcaster

import (
	"context"
	"fmt"
	mathRand "math/rand"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	mock2 "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestBroadcast(t *testing.T) {
	t.Run("called sequentially", func(t *testing.T) {
		signer := rand.AccAddr()
		b, clientCtx := setup(signer)

		iterations := int(rand.I64Between(20, 100))
		for i := 0; i < iterations; i++ {
			msgs := createMsgsWithSigner(signer, rand.I64Between(1, 20))

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			_, err := b.Broadcast(ctx, msgs...)
			cancel()

			assert.NoError(t, err)
		}

		assert.Len(t, clientCtx.Client.(*mock2.ClientMock).BroadcastTxSyncCalls(), iterations)
	})

	t.Run("sequence number updated correctly", func(t *testing.T) {
		signer := rand.AccAddr()
		b, clientCtx := setup(signer)

		iterations := int(rand.I64Between(200, 1000))
		wg := &sync.WaitGroup{}
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func(broadcaster *RefundableBroadcaster) {
				defer wg.Done()
				msgs := createMsgsWithSigner(signer, rand.I64Between(1, 20))
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
				_, err := broadcaster.Broadcast(ctx, msgs...)
				cancel()

				assert.NoError(t, err)
			}(b)
		}
		wg.Wait()

		foundSeqNo := map[uint64]bool{}
		maxSeqNo := uint64(0)
		for _, call := range clientCtx.Client.(*mock2.ClientMock).BroadcastTxSyncCalls() {
			decodedTx, err := clientCtx.TxConfig.TxDecoder()(call.Tx)
			assert.NoError(t, err)
			sigs, err := decodedTx.(authsigning.SigVerifiableTx).GetSignaturesV2()
			assert.NoError(t, err)
			for _, sig := range sigs {
				foundSeqNo[sig.Sequence] = true
				maxSeqNo = sig.Sequence
			}
		}
		assert.Equal(t, maxSeqNo+1, b.broadcaster.txFactory.Sequence())
		assert.NotContains(t, foundSeqNo, false)
	})

	t.Run("sequence number on blockchain trailing behind", func(t *testing.T) {
		accNo := mathRand.Uint64()
		seqNoOnChain := uint64(0)
		signer := rand.AccAddr()
		b, ctx := setup(signer)
		ctx.AccountRetriever.(*mock2.AccountRetrieverMock).GetAccountNumberSequenceFunc =
			func(client.Context, sdk.AccAddress) (uint64, uint64, error) {
				return accNo, seqNoOnChain, nil
			}
		ctx.Client.(*mock2.ClientMock).BroadcastTxSyncFunc = func(context.Context, types.Tx) (*coretypes.ResultBroadcastTx, error) {
			return &coretypes.ResultBroadcastTx{Code: abci.CodeTypeOK}, nil
		}

		iterations := int(rand.I64Between(200, 1000))
		wg := &sync.WaitGroup{}
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func(broadcaster *RefundableBroadcaster) {
				defer wg.Done()
				msgs := createMsgsWithSigner(signer, rand.I64Between(1, 20))
				_, err := broadcaster.Broadcast(context.TODO(), msgs...)
				assert.NoError(t, err)
			}(b)
		}
		wg.Wait()

		foundSeqNo := map[uint64]bool{}
		maxSeqNo := uint64(0)
		for _, call := range ctx.Client.(*mock2.ClientMock).BroadcastTxSyncCalls() {
			decodedTx, err := ctx.TxConfig.TxDecoder()(call.Tx)
			assert.NoError(t, err)
			sigs, err := decodedTx.(authsigning.SigVerifiableTx).GetSignaturesV2()
			assert.NoError(t, err)
			for _, sig := range sigs {
				foundSeqNo[sig.Sequence] = true
				maxSeqNo = sig.Sequence
			}
		}
		assert.Equal(t, maxSeqNo+1, b.broadcaster.txFactory.Sequence())
		assert.NotContains(t, foundSeqNo, false)
	})

	var (
		broadcaster *RefundableBroadcaster
		ctx         client.Context
		msgs        []sdk.Msg
	)

	Given("a broadcaster", func() {
		signer := rand.AccAddr()
		broadcaster, ctx = setup(signer)
	}).When("a batch of multiple messages", func() {
		msgs = createMsgsWithSigner(ctx.FromAddress, rand.I64Between(2, 20))
	}).
		When("the broadcaster returns a message specific error", func() {
			attempt := 0
			ctx.Client.(*mock2.ClientMock).BroadcastTxSyncFunc = func(context.Context, types.Tx) (*coretypes.ResultBroadcastTx, error) {
				if attempt == 0 {
					attempt++
					return nil, fmt.Errorf("backing off (retry in 6.590290871s ): rpc error: code = InvalidArgument desc = failed to execute message; message index: %d: "+
						"failed to execute message: voter axelarvaloper1qy9uq03rkpqkzwsa4fz7xxetkxttdcj6tf09pg has already voted: bridge error: reward module error: invalid request", rand.I64Between(0, int64(len(msgs))))
				}
				return &coretypes.ResultBroadcastTx{}, nil
			}
		}).
		Then("exclude the specific message and try the batch again", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			_, err := broadcaster.Broadcast(ctx, msgs...)
			assert.NoError(t, err)
		}).Run(t, 20)
}

func TestRetryPipeline_Push(t *testing.T) {
	testCases := []struct {
		label    string
		strategy func(minTimeOut time.Duration) utils.BackOff
	}{
		{"exponential", utils.ExponentialBackOff},
		{"linear", utils.LinearBackOff}}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("failed broadcast with %s backoff", testCase.label), func(t *testing.T) {

			retries := int(rand.I64Between(1, 20))
			backOff := testCase.strategy(20 * time.Nanosecond)
			pipeCap := int(rand.I64Between(10, 100000))
			p := NewPipelineWithRetry(pipeCap, retries, backOff, log.TestingLogger())

			iterations := int(rand.I64Between(5, 30))

			wg := &sync.WaitGroup{}
			wg.Add(iterations)
			for i := 0; i < iterations; i++ {
				go func(i int) {
					defer wg.Done()
					retry := 0
					err := p.Push(func() error {
						retry++
						return fmt.Errorf("retry %d, iteration %d", retry, i)
					}, func(_ error) bool { return true })
					assert.Error(t, err)
				}(i)
			}
			wg.Wait()
		})
	}

	t.Run("called concurrently", func(t *testing.T) {
		retries := int(rand.I64Between(1, 20))
		backOff := utils.LinearBackOff(2 * time.Microsecond)
		p := NewPipelineWithRetry(int(rand.I64Between(10, 100000)), retries, backOff, log.TestingLogger())

		iterations := int(rand.I64Between(20, 30))

		// introducing a data race on purpose to assert that broadcast calls are serialized
		callCounter := 0
		mockFunc := func() error {
			c := callCounter

			// simulate blocking
			timeout := time.Duration(rand.I64Between(0, 20)) * time.Millisecond
			time.Sleep(timeout)

			c++
			callCounter = c
			return nil
		}
		wg := &sync.WaitGroup{}
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func() {
				defer wg.Done()
				assert.NoError(t, p.Push(mockFunc, func(_ error) bool { return true }))
			}()
		}
		wg.Wait()
		// assert the func has been called the expected amount of times and no data races occurred
		assert.Equal(t, iterations, callCounter)
	})

	t.Run("no retry if retry filter is false", func(t *testing.T) {
		retries := int(rand.I64Between(1, 20))
		backOff := utils.LinearBackOff(2 * time.Microsecond)
		p := NewPipelineWithRetry(int(rand.I64Between(10, 100000)), retries, backOff, log.TestingLogger())

		iterations := int(rand.I64Between(20, 100))

		wg := &sync.WaitGroup{}
		wg.Add(iterations)
		for i := 0; i < iterations; i++ {
			go func(i int) {
				defer wg.Done()
				retry := 0
				err := p.Push(func() error {
					retry++
					return fmt.Errorf("retry %d, iteration %d", retry, i)
				}, func(_ error) bool { return false })
				assert.NoError(t, err)
				assert.True(t, retry == 1)
			}(i)
		}
		wg.Wait()
	})

}

func setup(signer sdk.AccAddress) (*RefundableBroadcaster, client.Context) {
	pk, err := cryptocodec.FromTmPubKeyInterface(ed25519.GenPrivKey().PubKey())
	if err != nil {
		panic(err)
	}
	key := &mock2.InfoMock{
		GetPubKeyFunc: func() cryptotypes.PubKey {
			return pk
		},
	}
	ctx := client.Context{
		FromAddress:   signer,
		BroadcastMode: flags.BroadcastSync,
		Client: &mock2.ClientMock{
			BroadcastTxSyncFunc: func(context.Context, types.Tx) (*coretypes.ResultBroadcastTx, error) {
				return &coretypes.ResultBroadcastTx{Code: abci.CodeTypeOK}, nil
			}},
		AccountRetriever: &mock2.AccountRetrieverMock{},
		ChainID:          rand.StrBetween(5, 20),
		TxConfig:         app.MakeEncodingConfig().TxConfig,
		Keyring: &mock2.KeyringMock{
			KeyFunc: func(string) (keyring.Info, error) {
				return key, nil
			},
		},
	}

	fs := pflag.NewFlagSet("test", pflag.PanicOnError)
	txf := tx.NewFactoryCLI(ctx, fs).WithSignMode(txsigning.SignMode_SIGN_MODE_UNSPECIFIED)
	p := NewPipelineWithRetry(100000, 10, func(int) time.Duration {
		return 0
	}, log.TestingLogger())

	b := WithRefund(NewBroadcaster(txf, ctx, p, 3, 15, log.TestingLogger()))
	return b, ctx
}

func createMsgsWithSigner(signer sdk.AccAddress, count int64) []sdk.Msg {
	return slices.Expand(func(_ int) sdk.Msg {
		return vote.NewVoteRequest(signer, exported.PollID(rand.I64Between(10, 100)), &evmtypes.VoteEvents{})
	}, int(count))
}
