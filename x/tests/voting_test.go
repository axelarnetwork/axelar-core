package tests

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"
	"golang.org/x/crypto/ripemd160"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	"github.com/axelarnetwork/axelar-core/x/broadcast"
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote"
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

/*
This file should function as an example of how to use the blockchain fake to run integration tests for
Cosmos modules without spinning up Tendermint consensus and multiple real nodes
*/

var txs = map[string]btcTypes.OutPointInfo{}

func Test_3Validators_VoteOn5Tx_Agree(t *testing.T) {
	// test data
	txCount := 5
	var outPoints []*wire.OutPoint
	var verifyMsgs []btcTypes.MsgVerifyTx
	for i := 0; i < txCount; i++ {
		txHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
		if err != nil {
			panic(err)
		}
		blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
		if err != nil {
			panic(err)
		}
		outPoints = append(outPoints, wire.NewOutPoint(txHash, 0))
		amount := testutils.RandIntBetween(0, 100000)
		confirmations := uint64(testutils.RandIntBetween(7, 10000))
		// deposit tx
		info := btcTypes.OutPointInfo{
			OutPoint:      outPoints[i],
			BlockHash:     blockHash,
			Amount:        btcutil.Amount(amount),
			Address:       randomAddress().EncodeAddress(),
			Confirmations: confirmations,
		}
		txs[blockHash.String()+txHash.String()] = info
		verifyMsgs = append(verifyMsgs, btcTypes.MsgVerifyTx{Sender: sdk.AccAddress("user1"), OutPointInfo: info})
	}

	// setting up the test infrastructure
	val1 := newValidator(sdk.ValAddress("val1"), 100)
	val2 := newValidator(sdk.ValAddress("val2"), 80)
	val3 := newValidator(sdk.ValAddress("val3"), 170)
	counter := testutils.RandIntBetween(1, 10000)
	validators := []exported.Validator{val1, val2, val3}
	staker := &snapMock.SnapshotterMock{
		GetLatestCounterFunc: func(sdk.Context) int64 { return counter },
		GetSnapshotFunc: func(sdk.Context, int64) (exported.Snapshot, bool) {
			return exported.Snapshot{Validators: validators, TotalPower: sdk.NewInt(350)}, true
		},
	}

	// Choose block size and optionally timeout according to the needs of the test
	blockChain := fake.NewBlockchain().WithBlockSize(2).WithBlockTimeOut(10 * time.Millisecond)

	b1 := fake.NewBroadcaster(testutils.Codec(), val1.GetOperator(), blockChain.Submit)
	b2 := fake.NewBroadcaster(testutils.Codec(), val2.GetOperator(), blockChain.Submit)
	b3 := fake.NewBroadcaster(testutils.Codec(), val3.GetOperator(), blockChain.Submit)

	n1, k1 := newNodeForVote("node1", b1, staker)
	n2, k2 := newNodeForVote("node2", b2, staker)
	n3, k3 := newNodeForVote("node3", b3, staker)
	nodes := []*fake.Node{n1, n2, n3}
	btcKeepers := []btcKeeper.Keeper{k1, k2, k3}

	for _, msg := range verifyMsgs {
		address, err := btcutil.DecodeAddress(msg.OutPointInfo.Address, k1.GetNetwork(n1.Ctx).Params)
		if err != nil {
			panic(err)
		}
		keyID := testutils.RandString(10)
		k1.SetKeyIDByAddress(n1.Ctx, address, keyID)
		k2.SetKeyIDByAddress(n2.Ctx, address, keyID)
		k3.SetKeyIDByAddress(n3.Ctx, address, keyID)
	}

	blockChain.AddNodes(nodes...)
	blockChain.Start()

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

	blockChain.WaitNBlocks(15)

	assert.True(t, allTxVoteCompleted(nodes, btcKeepers, outPoints))
}

func allTxVoteCompleted(nodes []*fake.Node, btcKeeper []btcKeeper.Keeper, outPoints []*wire.OutPoint) bool {
	allConfirmed := true
	for i, k := range btcKeeper {
		for _, out := range outPoints {
			if _, ok := k.GetVerifiedOutPointInfo(nodes[i].Ctx, out); !ok {
				allConfirmed = false
				break
			}
		}
	}
	return allConfirmed
}

func newNodeForVote(moniker string, broadcaster bcExported.Broadcaster, staker voteTypes.Snapshotter) (*fake.Node, btcKeeper.Keeper) {
	/*
		Multistore is mocked so we can more easily manipulate existing state and assert that specific state changes happen.
		For now, we never use the Header information, so we can just initialize an empty struct.
		We only simulate the actual transaction execution, not the test run before adding a transaction to the mempool,
		so isCheckTx should always be false.
		Tendermint already has a logger for tests defined, so that's probably good enough.
	*/
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	// Initialize all keepers and handlers you want to involve in the test
	vK := keeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(voteTypes.StoreKey), dbadapter.Store{DB: db.NewMemDB()}, staker, broadcaster)
	r := fake.NewRouter()
	vH := vote.NewHandler()

	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	btcK := btcKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(btcTypes.StoreKey), btcSubspace)
	btcK.SetParams(ctx, btcTypes.DefaultParams())

	nexusSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
	nexK := nexusKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(nexTypes.StoreKey), nexusSubspace)
	nexK.SetParams(ctx, nexusTypes.DefaultParams())

	// We use a fake for the bitcoin rpc client so we can control the responses from the "bitcoin" network
	btcH := bitcoin.NewHandler(btcK, vK, &btcMock.RPCClientMock{
		GetOutPointInfoFunc: func(bHash *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
			return txs[bHash.String()+out.Hash.String()], nil
		}}, nil, nexK)

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
	return fake.NewNode(moniker, ctx, r, nil).WithEndBlockers(eb), btcK
}

func newValidator(address sdk.ValAddress, power int64) *snapMock.ValidatorMock {
	return &snapMock.ValidatorMock{
		GetOperatorFunc:       func() sdk.ValAddress { return address },
		GetConsensusPowerFunc: func() int64 { return power }}
}

func randomAddress() btcutil.Address {
	addr, err := btcutil.NewAddressScriptHashFromHash(testutils.RandBytes(ripemd160.Size), btcTypes.DefaultParams().Network.Params)
	if err != nil {
		panic(err)
	}
	return addr
}
