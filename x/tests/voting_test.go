package tests

import (
	"context"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/store"
	test_utils "github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/mock"
	"github.com/axelarnetwork/axelar-core/x/broadcast"
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	btcMock "github.com/axelarnetwork/axelar-core/x/btc_bridge/tests/mock"
	btcTypes "github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
	"github.com/axelarnetwork/axelar-core/x/voting"
	vExported "github.com/axelarnetwork/axelar-core/x/voting/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/keeper"
	axTypes "github.com/axelarnetwork/axelar-core/x/voting/types"
)

/*
This file should function as an example of how to use the blockchain mock to run integration tests for
Cosmos modules without spinning up Tendermint consensus and multiple real nodes
*/

/*
Test data
while the hash and addresses are correctly formatted, these transactions are not real
*/
var (
	txIds = []string{
		"9cd961aca555c49f8e15011f64eae821fcefdb675aa880e901a6ea6c86700f60",
		"5bf532819c06bfe1dffe3a4d71ca9f5aff0f61699c84be797910f379c7dab48c",
		"03de73d454813a5909a8b3565dfef6852ed3418baa6930e3b7dbb9117702cf07",
		"9b9ef444466cd50c85e88f2dca957ffa66dcf79d47652c0667ea6b1f3108b77a",
		"74d39e87c810a80faff70dcbd988c661dbe283a27f903cd587ab9c0b221cc602"}
	hash1, _     = chainhash.NewHashFromStr(txIds[0])
	hash2, _     = chainhash.NewHashFromStr(txIds[1])
	hash3, _     = chainhash.NewHashFromStr(txIds[2])
	hash4, _     = chainhash.NewHashFromStr(txIds[3])
	hash5, _     = chainhash.NewHashFromStr(txIds[4])
	destinations = []string{
		"2NGZSCz4iug4677pdNAFtTJhRhBU7k7g6dY",
		"2Mv9yBkCHbmG3viJzFFSDsbhyNihYWnhbiB",
		"2MujoFWjkfm8vwn8bFWCwS1UP9KLLk7Eqyj",
		"2MwU72uP9DWeXxPoq4VBRPH4UkDkH2zkhah",
		"tb1q9mncjrazn5xgqdcqyjc0q0vzaytx2uzfc69q0x"}
	txs = map[string]*btcjson.TxRawResult{
		txIds[0]: {Txid: txIds[0], Hash: hash1.String(), Vout: []btcjson.Vout{vout(1, destinations[0])}, Confirmations: 9},
		txIds[1]: {Txid: txIds[1], Hash: hash2.String(), Vout: []btcjson.Vout{vout(2, destinations[1])}, Confirmations: 17},
		txIds[2]: {Txid: txIds[2], Hash: hash3.String(), Vout: []btcjson.Vout{vout(3, destinations[2])}, Confirmations: 9},
		txIds[3]: {Txid: txIds[3], Hash: hash4.String(), Vout: []btcjson.Vout{vout(4, destinations[3])}, Confirmations: 8},
		txIds[4]: {Txid: txIds[4], Hash: hash5.String(), Vout: []btcjson.Vout{vout(5, destinations[4])}, Confirmations: 12}}
)

func Test_3Validators_VoteOn5Tx_Agree(t *testing.T) {

	// setting up the test infrastructure
	val1 := mock.NewTestValidator(sdk.ValAddress("val1"), 100)
	val2 := mock.NewTestValidator(sdk.ValAddress("val2"), 80)
	val3 := mock.NewTestValidator(sdk.ValAddress("val3"), 170)
	staker := mock.NewTestStaker(val1, val2, val3)

	// Choose block size and optionally timeout according to the needs of the test
	blockChain := mock.NewBlockchain().WithBlockSize(2).WithBlockTimeOut(10 * time.Millisecond)

	b1 := mock.NewBroadcaster(test_utils.Codec(), sdk.AccAddress("broadcaster1"), val1.GetOperator(), blockChain.Input())
	b2 := mock.NewBroadcaster(test_utils.Codec(), sdk.AccAddress("broadcaster2"), val2.GetOperator(), blockChain.Input())
	b3 := mock.NewBroadcaster(test_utils.Codec(), sdk.AccAddress("broadcaster3"), val3.GetOperator(), blockChain.Input())

	n1, v1 := newNode("node1", b1, staker)
	n2, v2 := newNode("node2", b2, staker)
	n3, v3 := newNode("node3", b3, staker)
	nodes := []mock.Node{n1, n2, n3}
	voters := []btcTypes.Voter{v1, v2, v3}

	blockChain.AddNodes(nodes...)
	blockChain.Start()

	in := blockChain.Input()
	defer close(in)

	verifyMsgs := []sdk.Msg{
		prepareVerifyMsg(hash1, destinations[0], 1),
		prepareVerifyMsg(hash2, destinations[1], 2),
		prepareVerifyMsg(hash3, destinations[2], 3),
		prepareVerifyMsg(hash4, destinations[3], 4),
		prepareVerifyMsg(hash5, destinations[4], 5),
	}

	// test begin

	in <- broadcastTypes.NewMsgRegisterProxy(val1.GetOperator(), b1.Proxy)
	in <- broadcastTypes.NewMsgRegisterProxy(val2.GetOperator(), b2.Proxy)
	in <- broadcastTypes.NewMsgRegisterProxy(val3.GetOperator(), b3.Proxy)

	for _, msg := range verifyMsgs {
		in <- msg
	}

	timeOut, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
	reachedHeight25 := notifyOnBlock25(blockChain)

loop:
	for {
		select {
		case <-timeOut.Done():
			break loop
		case <-reachedHeight25:
			break loop
		default:
			confirmed := allTxConfirmed(nodes, voters)
			if confirmed {
				break loop
			}
			time.Sleep(20 * time.Millisecond)
		}
	}

	assert.True(t, allTxConfirmed(nodes, voters))
}

func notifyOnBlock25(blockChain mock.BlockChain) chan struct{} {
	reachedHeight25 := make(chan struct{})
	go func() {
		for {
			if blockChain.CurrentHeight() > 25 {
				close(reachedHeight25)
				break
			}
			time.Sleep(1 * time.Second)
		}

	}()
	return reachedHeight25
}

func vout(amount int, destination string) btcjson.Vout {
	return btcjson.Vout{
		Value:        btcutil.Amount(amount).ToBTC(),
		ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{destination}},
	}
}

func prepareVerifyMsg(hash *chainhash.Hash, destination string, amount int) sdk.Msg {
	return btcTypes.NewMsgVerifyTx(sdk.AccAddress("user1"), hash, 0, btcTypes.BtcAddress{
		Chain:         "testnet3",
		EncodedString: destination,
	}, btcutil.Amount(amount))
}

func newNode(moniker string, broadcaster bcExported.Broadcaster, staker axTypes.Staker) (mock.Node, btcTypes.Voter) {
	/*
		Multistore is mocked so we can more easily manipulate existing state and assert that specific state changes happen.
		For now, we never use the Header information, so we can just initialize an empty struct.
		We only simulate the actual transaction execution, not the test run before adding a transaction to the mempool,
		so isCheckTx should always be false.
		Tendermint already has a logger for tests defined, so that's probably good enough.
	*/
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	// Initialize all keepers and handlers you want to involve in the test
	vK := keeper.NewKeeper(test_utils.Codec(), mock.NewKVStoreKey(axTypes.StoreKey), store.NewSubjectiveStore(), staker, broadcaster)
	r := mock.NewRouter()
	vH := voting.NewHandler(vK, r)

	btcK := btcKeeper.NewBtcKeeper(test_utils.Codec(), mock.NewKVStoreKey(btcTypes.StoreKey))
	// We use a mock for the bitcoin rpc client so we can control the responses from the "bitcoin" network
	btcH := btc_bridge.NewHandler(btcK, vK, &btcMock.TestRPC{RawTxs: txs}, nil)

	broadcastH := broadcast.NewHandler(broadcaster)

	// Set the correct initial state in the keepers
	voting.InitGenesis(ctx, vK, axTypes.DefaultGenesisState())
	btc_bridge.InitGenesis(ctx, btcK, btcTypes.DefaultGenesisState())

	// Define all functions that should run at the end of a block
	eb := func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
		return voting.EndBlocker(ctx, req, vK)
	}

	// route all handlers
	r.AddRoute(axTypes.ModuleName, vH).
		AddRoute(btcTypes.ModuleName, btcH).
		AddRoute(broadcastTypes.ModuleName, broadcastH)
	return mock.NewNode(moniker, ctx, r).WithEndBlockers(eb), vK
}

func allTxConfirmed(nodes []mock.Node, voters []btcTypes.Voter) bool {
	allConfirmed := true
	for i, voter := range voters {
		for _, txId := range txIds {
			poll := vExported.PollMeta{Module: btcTypes.RouterKey, Type: btcTypes.MsgVerifyTx{}.Type(), ID: txId}
			if voter.Result(nodes[i].Ctx, poll) == nil {
				allConfirmed = false
				break
			}
		}
	}
	return allConfirmed
}
