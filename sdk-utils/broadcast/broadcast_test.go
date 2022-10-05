package broadcast_test

import (
	"context"
	"errors"
	"fmt"
	mathRand "math/rand"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	tx2 "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tm "github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/sdk-utils/broadcast"
	mock2 "github.com/axelarnetwork/axelar-core/sdk-utils/broadcast/mock"
	errors2 "github.com/axelarnetwork/axelar-core/utils/errors"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
	"github.com/axelarnetwork/utils/test/rand"
)

func TestStatefulBroadcaster(t *testing.T) {
	var (
		clientCtx        client.Context
		clientMock       *mock2.ClientMock
		accountRetriever *mock2.AccountRetrieverMock
		txf              tx.Factory
		broadcaster      broadcast.Broadcaster
		msgCount         int
		expectedResponse *coretypes.ResultTx
	)

	givenClientContext := Given("a client context in sync mode", func() {
		clientMock = &mock2.ClientMock{}
		clientCtx = client.Context{
			BroadcastMode: flags.BroadcastSync,
			Client:        clientMock,
			TxConfig:      app.MakeEncodingConfig().TxConfig,
		}
	})
	txFactory := Given("a tx factory", func() {
		accountRetriever = &mock2.AccountRetrieverMock{}
		keyringInfoMock := &mock2.InfoMock{
			GetPubKeyFunc: func() cryptotypes.PubKey { return ed25519.GenPrivKey().PubKey() },
		}
		txf = tx.Factory{}.
			WithChainID(rand.StrBetween(5, 20)).
			WithSimulateAndExecute(true).
			WithAccountRetriever(accountRetriever).
			WithTxConfig(clientCtx.TxConfig).
			WithKeybase(&mock2.KeyringMock{
				KeyFunc: func(string) (keyring.Info, error) {
					return keyringInfoMock, nil
				},
				SignFunc: func(string, []byte) ([]byte, cryptotypes.PubKey, error) {
					return rand.Bytes(10), nil, nil
				},
				ListFunc: func() ([]keyring.Info, error) {
					return []keyring.Info{keyringInfoMock}, nil
				},
			})
	})
	statefulBroadcaster := Given("a stateful broadcaster", func() {
		broadcaster = broadcast.WithStateManager(
			clientCtx,
			txf,
			log.TestingLogger(),
			broadcast.WithPollingInterval(1*time.Nanosecond),
			broadcast.WithResponseTimeout(10*time.Millisecond),
		)
	})

	accountExists := When("the account exists", func() {
		accountRetriever.EnsureExistsFunc = func(client.Context, sdk.AccAddress) error { return nil }
		accountRetriever.GetAccountNumberSequenceFunc = func(client.Context, sdk.AccAddress) (uint64, uint64, error) {
			return mathRand.Uint64(), mathRand.Uint64(), nil
		}
	})

	simulationSucceeds := When("the simulation succeeds", func() {
		clientMock.ABCIQueryWithOptionsFunc = func(context.Context, string, bytes.HexBytes, rpcclient.ABCIQueryOptions) (*coretypes.ResultABCIQuery, error) {
			bz, _ := (&tx2.SimulateResponse{GasInfo: &sdk.GasInfo{}}).Marshal()
			return &coretypes.ResultABCIQuery{Response: abci.ResponseQuery{Value: bz}}, nil
		}
	})

	getAccountSequenceMismatch := When("get an account seuqence mismatch", func() {
		clientMock.ABCIQueryWithOptionsFunc = func(context.Context, string, bytes.HexBytes, rpcclient.ABCIQueryOptions) (*coretypes.ResultABCIQuery, error) {
			return nil, sdkerrors.ErrWrongSequence
		}
	})

	broadcastSucceeds := When("broadcast succeeds", func() {
		clientMock.BroadcastTxSyncFunc = func(context.Context, tm.Tx) (*coretypes.ResultBroadcastTx, error) {
			return &coretypes.ResultBroadcastTx{Code: abci.CodeTypeOK}, nil
		}
	})

	txsGetExecuted := When("txs get executed correctly", func() {
		clientMock.TxFunc = func(context.Context, []byte, bool) (*coretypes.ResultTx, error) {
			expectedResponse = &coretypes.ResultTx{TxResult: abci.ResponseDeliverTx{
				Code: abci.CodeTypeOK,
				Log:  "some log",
			}}
			return expectedResponse, nil
		}
		clientMock.BlockFunc = func(_ context.Context, height *int64) (*coretypes.ResultBlock, error) {
			return &coretypes.ResultBlock{Block: &tm.Block{}}, nil
		}
	})

	txExecutionFailed := When("tx execution failed", func() {
		clientMock.TxFunc = func(context.Context, []byte, bool) (*coretypes.ResultTx, error) {
			expectedResponse = &coretypes.ResultTx{TxResult: abci.ResponseDeliverTx{
				Code: mathRand.Uint32(),
				Log:  "tx failed",
			}}
			return expectedResponse, nil
		}

		clientMock.BlockFunc = func(_ context.Context, height *int64) (*coretypes.ResultBlock, error) {
			return &coretypes.ResultBlock{Block: &tm.Block{}}, nil
		}
	})

	txNotFound := When("tx is not found", func() {
		clientMock.TxFunc = func(context.Context, []byte, bool) (*coretypes.ResultTx, error) {
			return nil, errors.New("not found")
		}
	})

	noError := Then("return no error", func(t *testing.T) {
		res, err := broadcaster.Broadcast(context.Background(), randomMsgs(3)...)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.TxResult.Log, res.RawLog)
	})

	timeout := Then("broadcast times out", func(t *testing.T) {
		_, err := broadcaster.Broadcast(context.Background(), randomMsgs(3)...)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timed out")
	})

	sendingMultipleMessages := When("sending multiple messages", func() {
		msgCount = 3
	})

	sendingNoMessages := When("sending no messages", func() {
		msgCount = 0
	})

	returnErrorWithCode := Then("return an error code", func(t *testing.T) {
		_, err := broadcaster.Broadcast(context.Background(), randomMsgs(msgCount)...)
		assert.Error(t, err)
		assert.True(t, errors2.Is[*sdkerrors.Error](err))
	})

	returnError := Then("return an error", func(t *testing.T) {
		_, err := broadcaster.Broadcast(context.Background(), randomMsgs(msgCount)...)
		assert.Error(t, err)
	})

	givenSetup := givenClientContext.
		Given2(txFactory).
		Given2(statefulBroadcaster)

	givenSetup.
		When2(sendingMultipleMessages).
		When2(accountExists).
		When2(simulationSucceeds).
		When2(broadcastSucceeds).
		When2(txsGetExecuted).
		Then2(noError).Run(t)

	givenSetup.
		When2(sendingMultipleMessages).
		When2(accountExists).
		When2(simulationSucceeds).
		When2(broadcastSucceeds).
		When2(txNotFound).
		Then2(timeout).Run(t)

	givenSetup.
		When2(sendingMultipleMessages).
		When2(accountExists).
		When2(simulationSucceeds).
		When2(broadcastSucceeds).
		When2(txExecutionFailed).
		Then2(returnErrorWithCode).Run(t)

	givenSetup.
		When2(sendingMultipleMessages).
		When2(accountExists).
		When2(getAccountSequenceMismatch).
		Then2(returnErrorWithCode).Run(t)

	givenSetup.
		When2(sendingMultipleMessages).
		When2(accountExists).
		When2(getAccountSequenceMismatch).
		Then2(returnErrorWithCode).Run(t)

	givenSetup.
		When2(sendingNoMessages).
		Then2(returnError).Run(t)
}

func TestWithRefund(t *testing.T) {
	var (
		broadcaster *mock2.BroadcasterMock
		refunder    broadcast.Broadcaster
	)

	Given("a refunding broadcaster", func() {
		broadcaster = &mock2.BroadcasterMock{}
		refunder = broadcast.WithRefund(broadcaster)
	}).
		When("the response contains the msgs of the tx", func() {
			broadcaster.BroadcastFunc = func(_ context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
				anyTx, _ := codectypes.NewAnyWithValue(&tx2.Tx{Body: &tx2.TxBody{Messages: slices.Map(msgs, unsafePack)}})
				return &sdk.TxResponse{Tx: anyTx}, nil
			}
		}).
		Then("all messages are of type RefundMsgRequest", func(t *testing.T) {
			res, err := refunder.Broadcast(context.Background(), randomMsgs(3)...)
			assert.NoError(t, err)
			for _, msg := range res.Tx.GetCachedValue().(sdk.Tx).GetMsgs() {
				assert.IsType(t, &types.RefundMsgRequest{}, msg)
			}
		}).Run(t)
}

func TestInBatches(t *testing.T) {
	var (
		broadcaster      *mock2.BroadcasterMock
		batched          broadcast.Broadcaster
		msgs             []sdk.Msg
		broadcastCalled  chan struct{}
		unblockBroadcast chan struct{}
	)

	Given("a batched broadcaster", func() {
		broadcaster = &mock2.BroadcasterMock{}
		batched = broadcast.Batched(broadcaster, 1, 5, log.TestingLogger())
	}).Branch(
		When("trying to broadcast 0 msgs", func() {
			msgs = []sdk.Msg{}
			broadcaster.BroadcastFunc = func(context.Context, ...sdk.Msg) (*sdk.TxResponse, error) { return nil, nil }
		}).
			Then("return error", func(t *testing.T) {
				_, err := batched.Broadcast(context.Background(), msgs...)
				assert.Error(t, err)
			}),

		When("there is low traffic", func() {
			broadcaster.BroadcastFunc = func(context.Context, ...sdk.Msg) (*sdk.TxResponse, error) {
				return &sdk.TxResponse{Data: "expected"}, nil
			}
		}).
			Then("send msgs one by one", func(t *testing.T) {
				for i := 0; i < 9; i++ {
					response, err := batched.Broadcast(context.Background(), randomMsgs(1)...)
					assert.NoError(t, err)
					assert.Equal(t, "expected", response.Data)
				}
				assert.Len(t, broadcaster.BroadcastCalls(), 9)
			}),
		When("there is high traffic", func() {
			broadcastCalled = make(chan struct{})
			unblockBroadcast = make(chan struct{})
			once := &sync.Once{}
			broadcaster.BroadcastFunc = func(_ context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
				once.Do(func() { close(broadcastCalled) })
				<-unblockBroadcast

				events := slices.Map(msgs, func(msg sdk.Msg) abci.Event { return abci.Event{Type: msg.String()} })
				return &sdk.TxResponse{Events: events}, nil
			}
		}).
			Then("batch msgs", func(t *testing.T) {
				wg := &sync.WaitGroup{}
				wg.Add(1)
				// block the broadcast pipeline with one message
				go func() {
					defer wg.Done()
					msgs := randomMsgs(1)
					response, err := batched.Broadcast(context.Background(), msgs...)
					assert.NoError(t, err)
					assert.Equal(t, msgs[0].String(), response.Events[0].Type)
				}()
				<-broadcastCalled
				// accumulate msgs in the backlog
				for i := 0; i < 9; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						msgs := randomMsgs(1)
						response, err := batched.Broadcast(context.Background(), msgs...)
						assert.NoError(t, err)
						// make sure the expected msg is part of the response
						assert.True(t,
							slices.Any(response.Events,
								func(event abci.Event) bool { return msgs[0].String() == event.Type }))
					}()
				}
				close(unblockBroadcast)
				wg.Wait()
				assert.Less(t, len(broadcaster.BroadcastCalls()), 10)
			}),
	).Run(t)
}

func TestWithRetry(t *testing.T) {
	var (
		broadcaster *mock2.BroadcasterMock
		retry       broadcast.Broadcaster
	)

	Given("a retry broadcaster", func() {
		broadcaster = &mock2.BroadcasterMock{}
		retry = broadcast.WithRetry(broadcaster, 3, 1*time.Nanosecond, log.TestingLogger())
	}).Branch(
		When("one of the msgs fails", func() {
			once := &sync.Once{}
			broadcaster.BroadcastFunc = func(_ context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
				var err error
				once.Do(func() {
					err = fmt.Errorf("some error; message index: %d", rand.I64Between(0, len(msgs)))
				})
				if err != nil {
					return nil, err
				}

				events := slices.Map(msgs, func(msg sdk.Msg) abci.Event { return abci.Event{Type: msg.String()} })
				return &sdk.TxResponse{Events: events}, nil
			}
		}).
			Then("retry without that msg", func(t *testing.T) {
				msgs := randomMsgs(10)
				res, err := retry.Broadcast(context.Background(), msgs...)
				assert.NoError(t, err)

				matches := 0
				for _, msg := range msgs {
					if slices.Any(res.Events,
						func(event abci.Event) bool {
							return msg.String() == event.Type
						}) {
						matches++
					}
				}
				assert.Equal(t, 9, matches)
			}),

		AsTestCases[error](sdkerrors.ErrWrongSequence, sdkerrors.ErrOutOfGas, errors.New("some error")).
			Map(
				func(err error) Runner {
					return When("the broadcast fails because of a retriable error", func() {
						broadcaster.BroadcastFunc = func(context.Context, ...sdk.Msg) (*sdk.TxResponse, error) {
							return nil, err
						}
					}).
						Then("retry up to the max parameter", func(t *testing.T) {
							_, err := retry.Broadcast(context.Background(), randomMsgs(10)...)
							assert.Error(t, err)
							assert.Len(t, broadcaster.BroadcastCalls(), 4) // try once + 3 retries
						})
				},
			),

		When("the msg execution fails", func() {
			broadcaster.BroadcastFunc = func(context.Context, ...sdk.Msg) (*sdk.TxResponse, error) {
				return nil, sdkerrors.New("codespace", mathRand.Uint32(), "error")
			}
		}).
			Then("don't retry broadcast", func(t *testing.T) {
				_, err := retry.Broadcast(context.Background(), randomMsgs(10)...)
				assert.Error(t, err)
				assert.Len(t, broadcaster.BroadcastCalls(), 1)
			}),
	).Run(t)
}

func TestSuppressExecutionErrs(t *testing.T) {
	broadcaster := &mock2.BroadcasterMock{
		BroadcastFunc: func(context.Context, ...sdk.Msg) (*sdk.TxResponse, error) {
			return nil, sdkerrors.New("codespace", mathRand.Uint32(), "error")
		}}

	suppressor := broadcast.SuppressExecutionErrs(broadcaster, log.TestingLogger())

	_, err := suppressor.Broadcast(context.Background(), randomMsgs(5)...)
	assert.NoError(t, err)
}

func unsafePack(value sdk.Msg) *codectypes.Any {
	return codectypes.UnsafePackAny(value)
}

func randomMsgs(count int) []sdk.Msg {
	var msgs []sdk.Msg
	sender := rand.AccAddr()
	for i := 0; i < count; i++ {
		msg := evm.NewLinkRequest(
			sender,
			rand.StrBetween(5, 10),
			rand.StrBetween(5, 10),
			evm.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))).Hex(),
			rand.StrBetween(5, 10),
		)
		msgs = append(msgs, msg)
	}
	return msgs
}
