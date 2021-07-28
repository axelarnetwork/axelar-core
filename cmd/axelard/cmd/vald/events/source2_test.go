package events_test

import (
	"bytes"
	"context"
	"fmt"
	mathRand "math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tm "github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/events"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/events/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

type testEnv struct {
	testutils.ErrorCache
	context         context.Context
	cancel          context.CancelFunc
	notifierClient  *clientMock
	notifier        events.BlockNotifier
	receivedBlocks  <-chan int64
	notifierErrChan <-chan error

	notifierMock   *notifierMock
	resultsClient  *mock.BlockResultClientMock
	blockSource    events.BlockSource
	resultsErrChan <-chan error
	results        <-chan *coretypes.ResultBlockResults
	blocks         map[int64]*coretypes.ResultBlockResults
}

func (t *testEnv) NotifierClient() error {
	t.notifierClient = NewClientMock()
	return nil
}

func (t *testEnv) Context() error {
	t.context, t.cancel = context.WithCancel(context.Background())
	return nil
}

func (t *testEnv) BlockNotifierStartingAtBlock(start int64) error {
	t.notifier = events.NewBlockNotifier(t.notifierClient, start, log.TestingLogger(),
		events.Timeout(1*time.Millisecond), events.Retries(1), events.KeepAlive(1*time.Millisecond))
	return nil
}

func (t *testEnv) BlockAvailable(latest int64) error {
	t.notifierClient.NextBlock(latest)
	return nil
}

func (t *testEnv) ClientWithoutSubscription() error {
	t.notifierClient.SubscribeFunc = func(context.Context, string, string, ...int) (<-chan coretypes.ResultEvent, error) {
		return nil, nil
	}
	return nil
}

func (t *testEnv) ClientWithStaleQuery() error {
	t.notifierClient.LatestBlockHeightFunc = func(context.Context) (int64, error) { return 0, nil }
	return nil
}

func (t *testEnv) ClientSubscriptionFails() error {
	t.notifierClient.SubscribeFunc = func(context.Context, string, string, ...int) (<-chan coretypes.ResultEvent, error) {
		return nil, fmt.Errorf("some error")
	}
	return nil
}

func (t *testEnv) ClientQueryFails() error {
	t.notifierClient.LatestBlockHeightFunc = func(context.Context) (int64, error) {
		return 0, fmt.Errorf("some error")
	}
	return nil
}

func (t *testEnv) ResultsClientFails() error {
	t.resultsClient.BlockResultsFunc = func(context.Context, *int64) (*coretypes.ResultBlockResults, error) {
		return nil, fmt.Errorf("some error")
	}

	// trigger lookup once more
	t.notifierMock.BlockAvailable(rand.PosI64())
	return nil
}

func (t *testEnv) TryReceiveBlockHeights() error {
	t.receivedBlocks, t.notifierErrChan = t.notifier.BlockHeights(t.context)
	return nil
}

func (t *testEnv) ContextIsCanceled() error {
	t.cancel()
	return nil
}

func (t *testEnv) ReceiveAllBlocks(start, latest int64) error {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

loop:
	for i := int64(0); i <= latest-start; i++ {
		select {
		case <-timeout.Done():
			assert.Fail(t, "timed out", "delivered %d of %d blocks", i, latest-start+1)
			break loop
		case err := <-t.notifierErrChan:
			assert.Fail(t, "returned error", err.Error())
			break loop
		case receivedBlock := <-t.receivedBlocks:
			assert.Equal(t, start+i, receivedBlock)
		}
	}

	return t.Error
}

func (t *testEnv) BlockChannelGetsClosed() error {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

loop:
	for {
		select {
		case <-timeout.Done():
			assert.Fail(t, "channel should have been closed")
			break loop
		case err := <-t.notifierErrChan:
			assert.Fail(t, "returned error", err.Error())
			break loop
		case _, ok := <-t.receivedBlocks:
			if !ok {
				break loop
			}
		}
	}
	return t.Error
}

func (t *testEnv) ResultChannelGetsClosed() error {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

loop:
	for {
		select {
		case <-timeout.Done():
			assert.Fail(t, "channel should have been closed")
			break loop
		case err := <-t.resultsErrChan:
			assert.Fail(t, "returned error", err.Error())
			break loop
		case _, ok := <-t.results:
			if !ok {
				break loop
			}
		}
	}
	return t.Error
}

func (t *testEnv) BlockNotifierFails() error {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
loop:
	for {
		select {
		case <-timeout.Done():
			assert.Fail(t, "should have failed with error")
			break loop
		case err := <-t.notifierErrChan:
			assert.Error(t, err)
			break loop
		case <-t.receivedBlocks:
		}
	}
	return t.Error
}

func (t *testEnv) BlockResultSourceFails() error {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
loop:
	for {
		select {
		case <-timeout.Done():
			assert.Fail(t, "should have failed with error")
			break loop
		case err := <-t.resultsErrChan:
			assert.Error(t, err)
			break loop
		case <-t.results:
			continue
		}
	}
	return t.Error
}

func (t *testEnv) BlockNotifier() error {
	t.notifierMock = NewNotifierMock()
	return nil
}

func (t *testEnv) ResultsClient() error {
	t.resultsClient = &mock.BlockResultClientMock{
		BlockResultsFunc: func(_ context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
			b, ok := t.blocks[*height]
			if !ok {
				return nil, fmt.Errorf("not found")
			}
			return b, nil
		},
	}
	return nil
}

func (t *testEnv) BlockResultSource() error {
	t.blockSource = events.NewBlockSource(t.resultsClient, t.notifierMock, 1*time.Second)
	return nil
}

func (t *testEnv) BlocksAvailable(start int64, latest int64) error {
	t.blocks = make(map[int64]*coretypes.ResultBlockResults)
	for i := start; i <= latest; i++ {
		t.blocks[i] = &coretypes.ResultBlockResults{Height: i}
		t.notifierMock.BlockAvailable(i)
	}
	return nil
}

func (t *testEnv) ReceiveAllResults(start int64, latest int64) error {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

loop:
	for i := int64(0); i <= latest-start; i++ {
		select {
		case <-timeout.Done():
			assert.Fail(t, "timed out", "delivered %d of %d blocks", i, latest-start+1)
			break loop
		case err := <-t.resultsErrChan:
			assert.Fail(t, "returned error", err.Error())
			break loop
		case result := <-t.results:
			assert.Equal(t, start+i, result.Height)
		}
	}

	return t.Error
}

func (t *testEnv) TryReceiveBlockchainResults() error {
	t.results, t.resultsErrChan = t.blockSource.BlockResults(t.context)
	return nil
}

func (t *testEnv) MakeNotifierFail() error {
	t.notifierMock.Fail()
	return nil
}

func InitializeNotifierScenario(ctx *godog.ScenarioContext) {
	t := &testEnv{}
	ctx.Step(`^a block notifier starting at block (\-?\d+)$`, t.BlockNotifierStartingAtBlock)
	ctx.Step(`^a cancellable context$`, t.Context)
	ctx.Step(`^a client to get block heights$`, t.NotifierClient)
	ctx.Step(`^block (\d+) is available$`, t.BlockAvailable)
	ctx.Step(`^I receive all blocks from (\d+) to (\d+)$`, t.ReceiveAllBlocks)
	ctx.Step(`^I try to receive block heights$`, t.TryReceiveBlockHeights)
	ctx.Step(`^the block channel gets closed$`, t.BlockChannelGetsClosed)
	ctx.Step(`^the block notifier fails$`, t.BlockNotifierFails)
	ctx.Step(`^the client only provides blocks through a query$`, t.ClientWithoutSubscription)
	ctx.Step(`^the client only provides blocks through events$`, t.ClientWithStaleQuery)
	ctx.Step(`^the client subscription fails$`, t.ClientSubscriptionFails)
	ctx.Step(`^the client\'s query fails$`, t.ClientQueryFails)
	ctx.Step(`^the context is canceled$`, t.ContextIsCanceled)
}

func InitializeBlockResultsScenario(ctx *godog.ScenarioContext) {
	t := &testEnv{}
	ctx.Step(`^a block notifier$`, t.BlockNotifier)
	ctx.Step(`^a block result source$`, t.BlockResultSource)
	ctx.Step(`^a cancellable context$`, t.Context)
	ctx.Step(`^a client to get blockchain events$`, t.ResultsClient)
	ctx.Step(`^blocks (\d+) to (\d+) are available$`, t.BlocksAvailable)
	ctx.Step(`^I receive all results from (\d+) to (\d+)$`, t.ReceiveAllResults)
	ctx.Step(`^I try to receive blockchain results$`, t.TryReceiveBlockchainResults)
	ctx.Step(`^the block notifier fails$`, t.MakeNotifierFail)
	ctx.Step(`^the block result source fails$`, t.BlockResultSourceFails)
	ctx.Step(`^the client fails$`, t.ResultsClientFails)
	ctx.Step(`^the context is canceled$`, t.ContextIsCanceled)
	ctx.Step(`^the result channel gets closed$`, t.ResultChannelGetsClosed)
}

func TestBlockNotifier(t *testing.T) {
	var output = &bytes.Buffer{}
	testSuite := godog.TestSuite{
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			ctx.AfterSuite(func() {
				_, _ = os.Stdout.WriteString(output.String())
			})
		},
		ScenarioInitializer: InitializeNotifierScenario,
		Options:             &godog.Options{Paths: []string{"features/blockheights.feature"}, Strict: true, Format: "pretty", Output: output},
	}

	if testSuite.Run() != 0 {
		t.FailNow()
	}
}

func TestBlockSource(t *testing.T) {
	var output = &bytes.Buffer{}
	testSuite := godog.TestSuite{
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			ctx.AfterSuite(func() {
				_, _ = os.Stdout.WriteString(output.String())
			})
		},
		ScenarioInitializer: InitializeBlockResultsScenario,
		Options:             &godog.Options{Paths: []string{"features/blockresults.feature"}, Strict: true, Format: "pretty", Output: output},
	}
	if testSuite.Run() != 0 {
		t.FailNow()
	}
}

func randomTxResults(count int64) []*abci.ResponseDeliverTx {
	resp := make([]*abci.ResponseDeliverTx, 0, count)
	for i := 0; i < cap(resp); i++ {
		resp = append(resp, &abci.ResponseDeliverTx{
			Code:      mathRand.Uint32(),
			Data:      rand.Bytes(int(rand.I64Between(100, 200))),
			Log:       rand.StrBetween(5, 100),
			Info:      rand.StrBetween(5, 100),
			GasWanted: rand.PosI64(),
			GasUsed:   rand.PosI64(),
			Events:    randomEvents(rand.I64Between(1, 10)),
			Codespace: rand.StrBetween(5, 100),
		})
	}

	return resp
}

func toEvents(table *messages.PickleStepArgument_PickleTable) map[int64]*coretypes.ResultBlockResults {
	results := make(map[int64]*coretypes.ResultBlockResults)

	for i := 1; i < len(table.Rows); i++ {
		height, err := strconv.ParseInt(table.Rows[i].Cells[0].GetValue(), 10, 64)
		if err != nil {
			panic(err)
		}

		beginBlockEventCount, err := strconv.ParseInt(table.Rows[i].Cells[1].GetValue(), 10, 64)
		if err != nil {
			panic(err)
		}
		txResultCount, err := strconv.ParseInt(table.Rows[i].Cells[2].GetValue(), 10, 64)
		if err != nil {
			panic(err)
		}
		endBlockEventCount, err := strconv.ParseInt(table.Rows[i].Cells[3].GetValue(), 10, 64)
		if err != nil {
			panic(err)
		}

		results[height] = &coretypes.ResultBlockResults{
			Height:           height,
			TxsResults:       randomTxResults(txResultCount),
			BeginBlockEvents: randomEvents(beginBlockEventCount),
			EndBlockEvents:   randomEvents(endBlockEventCount),
		}
	}

	return results
}

func randomEvents(count int64) []abci.Event {
	e := make([]abci.Event, 0, count)
	for i := 0; i < cap(e); i++ {
		e = append(e, abci.Event{
			Type:       tm.EventTx,
			Attributes: randomAttributes(rand.I64Between(1, 10)),
		})
	}
	return e
}

func randomAttributes(count int64) []abci.EventAttribute {
	attributes := make([]abci.EventAttribute, 0, count)
	for i := 0; i < cap(attributes); i++ {
		attributes = append(attributes, abci.EventAttribute{
			Key:   rand.BytesBetween(5, 100),
			Value: rand.BytesBetween(5, 100),
			Index: rand.Bools(0.5).Next(),
		})
	}
	return attributes
}

type notifierMock struct {
	*mock.BlockNotifierMock
	blocks chan int64
	errors chan error
}

func NewNotifierMock() *notifierMock {
	notifier := &notifierMock{blocks: make(chan int64, 10000), errors: make(chan error, 10000)}
	notifier.BlockNotifierMock = &mock.BlockNotifierMock{
		BlockHeightsFunc: func(_ context.Context) (<-chan int64, <-chan error) {
			return notifier.blocks, notifier.errors
		}}

	return notifier
}

func (n *notifierMock) Fail() {
	n.errors <- fmt.Errorf("some error")
}

func (n *notifierMock) BlockAvailable(i int64) {
	n.blocks <- i
}
