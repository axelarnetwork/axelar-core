package tests

import (
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
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	"github.com/axelarnetwork/axelar-core/x/broadcast"
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote"
	vExported "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

/*
This file should function as an example of how to use the blockchain fake to run integration tests for
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
	val1 := newValidator(sdk.ValAddress("val1"), 100)
	val2 := newValidator(sdk.ValAddress("val2"), 80)
	val3 := newValidator(sdk.ValAddress("val3"), 170)
	round := testutils.RandIntBetween(1, 10000)
	validators := []exported.Validator{val1, val2, val3}
	staker := &snapMock.SnapshotterMock{
		GetLatestRoundFunc: func(ctx sdk.Context) int64 { return round },
		GetSnapshotFunc: func(ctx sdk.Context, round int64) (exported.Snapshot, bool) {
			return exported.Snapshot{Validators: validators, TotalPower: sdk.NewInt(350)}, true
		},
	}

	// Choose block size and optionally timeout according to the needs of the test
	blockChain := fake.NewBlockchain().WithBlockSize(2).WithBlockTimeOut(10 * time.Millisecond)

	b1 := fake.NewBroadcaster(testutils.Codec(), val1.GetOperator(), blockChain.Submit)
	b2 := fake.NewBroadcaster(testutils.Codec(), val2.GetOperator(), blockChain.Submit)
	b3 := fake.NewBroadcaster(testutils.Codec(), val3.GetOperator(), blockChain.Submit)

	n1, v1 := newNodeForVote("node1", b1, staker)
	n2, v2 := newNodeForVote("node2", b2, staker)
	n3, v3 := newNodeForVote("node3", b3, staker)
	nodes := []fake.Node{n1, n2, n3}
	voters := []btcTypes.Voter{v1, v2, v3}

	blockChain.AddNodes(nodes...)
	blockChain.Start()

	verifyMsgs := []sdk.Msg{
		prepareVerifyMsg(hash1, destinations[0], 1),
		prepareVerifyMsg(hash2, destinations[1], 2),
		prepareVerifyMsg(hash3, destinations[2], 3),
		prepareVerifyMsg(hash4, destinations[3], 4),
		prepareVerifyMsg(hash5, destinations[4], 5),
	}

	// test begin

	// register proxies
	res := <-blockChain.Submit(broadcastTypes.MsgRegisterProxy{Principal: val1.GetOperator(), Proxy: sdk.AccAddress("proxy1")})
	assert.NoError(t, res.Error)
	res = <-blockChain.Submit(broadcastTypes.MsgRegisterProxy{Principal: val2.GetOperator(), Proxy: sdk.AccAddress("proxy2")})
	assert.NoError(t, res.Error)
	res = <-blockChain.Submit(broadcastTypes.MsgRegisterProxy{Principal: val3.GetOperator(), Proxy: sdk.AccAddress("proxy3")})
	assert.NoError(t, res.Error)

	// verify txs
	for _, msg := range verifyMsgs {
		res := <-blockChain.Submit(msg)
		assert.NoError(t, res.Error)
	}

	<-blockChain.WaitNBlocks(15)

	assert.True(t, allTxConfirmed(nodes, voters))
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

func newNodeForVote(moniker string, broadcaster bcExported.Broadcaster, staker voteTypes.Snapshotter) (fake.Node, btcTypes.Voter) {
	/*
		Multistore is mocked so we can more easily manipulate existing state and assert that specific state changes happen.
		For now, we never use the Header information, so we can just initialize an empty struct.
		We only simulate the actual transaction execution, not the test run before adding a transaction to the mempool,
		so isCheckTx should always be false.
		Tendermint already has a logger for tests defined, so that's probably good enough.
	*/
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	// Initialize all keepers and handlers you want to involve in the test
	vK := keeper.NewKeeper(testutils.Codec(), fake.NewKVStoreKey(voteTypes.StoreKey), store.NewSubjectiveStore(), staker, broadcaster)
	r := fake.NewRouter()
	vH := vote.NewHandler()

	btcK := btcKeeper.NewBtcKeeper(testutils.Codec(), fake.NewKVStoreKey(btcTypes.StoreKey))
	// We use a fake for the bitcoin rpc client so we can control the responses from the "bitcoin" network
	btcH := bitcoin.NewHandler(btcK, vK, &btcMock.RPCClientMock{
		GetRawTransactionVerboseFunc: func(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
			return txs[hash.String()], nil
		}}, nil)

	broadcastH := broadcast.NewHandler(broadcaster)

	// Set the correct initial state in the keepers
	vote.InitGenesis(ctx, vK, voteTypes.DefaultGenesisState())
	bitcoin.InitGenesis(ctx, btcK, btcTypes.DefaultGenesisState())

	// Define all functions that should run at the end of a block
	eb := func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
		return vote.EndBlocker(ctx, req, vK)
	}

	// route all handlers
	r.AddRoute(voteTypes.ModuleName, vH).
		AddRoute(btcTypes.ModuleName, btcH).
		AddRoute(broadcastTypes.ModuleName, broadcastH)
	return fake.NewNode(moniker, ctx, r).WithEndBlockers(eb), vK
}

func allTxConfirmed(nodes []fake.Node, voters []btcTypes.Voter) bool {
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

func newValidator(address sdk.ValAddress, power int64) *snapMock.ValidatorMock {
	return &snapMock.ValidatorMock{
		GetOperatorFunc:       func() sdk.ValAddress { return address },
		GetConsensusPowerFunc: func() int64 { return power }}
}
