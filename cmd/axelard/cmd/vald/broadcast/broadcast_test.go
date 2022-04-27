package broadcast_test

import (
	"context"
	"errors"
	mathRand "math/rand"
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

	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
	"github.com/axelarnetwork/utils/test/rand"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/utils"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
)

func TestStatefulBroadcaster(t *testing.T) {
	var (
		clientCtx        client.Context
		clientMock       *mock.ClientMock
		accountRetriever *mock.AccountRetrieverMock
		txf              tx.Factory
		broadcaster      broadcast.Broadcaster
		msgCount         int
		expectedResponse *coretypes.ResultTx
	)

	givenClientContext := Given("a client context in sync mode", func() {
		clientMock = &mock.ClientMock{}
		clientCtx = client.Context{
			BroadcastMode: flags.BroadcastSync,
			Client:        clientMock,
			TxConfig:      app.MakeEncodingConfig().TxConfig,
		}
	})
	txFactory := Given("a tx factory", func() {
		accountRetriever = &mock.AccountRetrieverMock{}
		txf = tx.Factory{}.
			WithChainID(rand.StrBetween(5, 20)).
			WithSimulateAndExecute(true).
			WithAccountRetriever(accountRetriever).
			WithTxConfig(clientCtx.TxConfig).
			WithKeybase(&mock.KeyringMock{
				KeyFunc: func(string) (keyring.Info, error) {
					return &mock.InfoMock{
						GetPubKeyFunc: func() cryptotypes.PubKey { return ed25519.GenPrivKey().PubKey() },
					}, nil
				},
				SignFunc: func(string, []byte) ([]byte, cryptotypes.PubKey, error) {
					return rand.Bytes(10), nil, nil
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
		assert.True(t, utils.IsABCIError(err))
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
		broadcaster *mock.BroadcasterMock
		refunder    broadcast.Broadcaster
	)

	broadcaster = &mock.BroadcasterMock{}

	Given("a refunding broadcaster", func() {
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
