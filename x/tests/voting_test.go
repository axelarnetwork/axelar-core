package tests

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/store"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
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
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

/*
This file should function as an example of how to use the blockchain fake to run integration tests for
Cosmos modules without spinning up Tendermint consensus and multiple real nodes
*/

var txs = map[string]*btcjson.TxRawResult{}

func Test_3Validators_VoteOn5Tx_Agree(t *testing.T) {
	// test data
	txCount := 5
	var hashes []*chainhash.Hash
	var verifyMsgs []sdk.Msg
	for i := 0; i < txCount; i++ {
		prevHash := createHash()
		hash := createHash()
		hashes = append(hashes, hash)
		recipient := createAddress()
		prevVoutIdx := uint32(testutils.RandIntBetween(0, 100))
		amount := testutils.RandIntBetween(0, 100000)
		confirmations := uint64(testutils.RandIntBetween(7, 10000))
		// deposit tx
		txs[hash.String()] = &btcjson.TxRawResult{
			Txid:          hash.String(),
			Hash:          hash.String(),
			Vin:           []btcjson.Vin{{Txid: prevHash.String(), Vout: prevVoutIdx}},
			Vout:          []btcjson.Vout{createVout(amount, recipient.String())},
			Confirmations: confirmations,
		}

		verifyMsgs = append(verifyMsgs, prepareVerifyMsg(hash, recipient.String(), amount))
	}

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

	n1, k1 := newNodeForVote("node1", b1, staker)
	n2, k2 := newNodeForVote("node2", b2, staker)
	n3, k3 := newNodeForVote("node3", b3, staker)
	nodes := []fake.Node{n1, n2, n3}
	btcKeepers := []btcKeeper.Keeper{k1, k2, k3}

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

	<-blockChain.WaitNBlocks(15)

	assert.True(t, allTxVoteCompleted(nodes, btcKeepers, hashes))
}

func allTxVoteCompleted(nodes []fake.Node, btcKeeper []btcKeeper.Keeper, hashes []*chainhash.Hash) bool {
	allConfirmed := true
	for i, k := range btcKeeper {
		for _, hash := range hashes {
			if _, ok := k.GetUTXO(nodes[i].Ctx, hash.String()); !ok {
				allConfirmed = false
				break
			}
		}
	}
	return allConfirmed
}

func createAddress() btcutil.Address {
	sk, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	pk := btcec.PublicKey(sk.PublicKey)
	addr, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk.SerializeCompressed()), &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}
	return addr
}

func createHash() *chainhash.Hash {
	var bz []byte
	for _, b := range testutils.RandIntsBetween(0, 256).Take(chainhash.HashSize) {
		bz = append(bz, byte(b))
	}
	hash, err := chainhash.NewHash(bz)
	if err != nil {
		panic(err)
	}
	return hash
}

func createVout(amount int64, recipient string) btcjson.Vout {
	return btcjson.Vout{
		Value:        btcutil.Amount(amount).ToBTC(),
		ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{recipient}},
	}
}

func prepareVerifyMsg(hash *chainhash.Hash, recipient string, amount int64) sdk.Msg {
	return btcTypes.NewMsgVerifyTx(sdk.AccAddress("user1"), hash, 0, btcTypes.BtcAddress{
		Network:       btcTypes.Network(chaincfg.MainNetParams.Name),
		EncodedString: recipient,
	}, btcutil.Amount(amount))
}

func newNodeForVote(moniker string, broadcaster bcExported.Broadcaster, staker voteTypes.Snapshotter) (fake.Node, btcKeeper.Keeper) {
	/*
		Multistore is mocked so we can more easily manipulate existing state and assert that specific state changes happen.
		For now, we never use the Header information, so we can just initialize an empty struct.
		We only simulate the actual transaction execution, not the test run before adding a transaction to the mempool,
		so isCheckTx should always be false.
		Tendermint already has a logger for tests defined, so that's probably good enough.
	*/
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	// Initialize all keepers and handlers you want to involve in the test
	vK := keeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(voteTypes.StoreKey), store.NewSubjectiveStore(), staker, broadcaster)
	r := fake.NewRouter()
	vH := vote.NewHandler()

	btcK := btcKeeper.NewBtcKeeper(testutils.Codec(), sdk.NewKVStoreKey(btcTypes.StoreKey))
	// We use a fake for the bitcoin rpc client so we can control the responses from the "bitcoin" network
	btcH := bitcoin.NewHandler(btcK, vK, &btcMock.RPCClientMock{
		GetOutPointInfoFunc: func(out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
			return txs[out.Hash.String()], nil
		}}, nil, &btcMock.BalancerMock{
		GetRecipientFunc: func(ctx sdk.Context, sender balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
			return balance.CrossChainAddress{}, false
		},
	})

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
	return fake.NewNode(moniker, ctx, r).WithEndBlockers(eb), btcK
}

func newValidator(address sdk.ValAddress, power int64) *snapMock.ValidatorMock {
	return &snapMock.ValidatorMock{
		GetOperatorFunc:       func() sdk.ValAddress { return address },
		GetConsensusPowerFunc: func() int64 { return power }}
}
