package tests

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/test-util/mock"
	"github.com/axelarnetwork/axelar-core/x/axelar"
	"github.com/axelarnetwork/axelar-core/x/axelar/keeper"
	axTypes "github.com/axelarnetwork/axelar-core/x/axelar/types"
	"github.com/axelarnetwork/axelar-core/x/broadcast"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	btcMock "github.com/axelarnetwork/axelar-core/x/btc_bridge/tests/mock"
	btcTypes "github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
	axMock "github.com/axelarnetwork/axelar-core/x/tests/mock"
)

// test data
var (
	// while the hash and addresses are correctly formatted, these transactions are not real
	txId1    = "9cd961aca555c49f8e15011f64eae821fcefdb675aa880e901a6ea6c86700f60"
	hash1, _ = chainhash.NewHashFromStr(txId1)
	dest1    = "2NGZSCz4iug4677pdNAFtTJhRhBU7k7g6dY"
	txId2    = "5bf532819c06bfe1dffe3a4d71ca9f5aff0f61699c84be797910f379c7dab48c"
	hash2, _ = chainhash.NewHashFromStr(txId2)
	dest2    = "2Mv9yBkCHbmG3viJzFFSDsbhyNihYWnhbiB"
	txId3    = "03de73d454813a5909a8b3565dfef6852ed3418baa6930e3b7dbb9117702cf07"
	hash3, _ = chainhash.NewHashFromStr(txId3)
	dest3    = "2MujoFWjkfm8vwn8bFWCwS1UP9KLLk7Eqyj"
	txId4    = "9b9ef444466cd50c85e88f2dca957ffa66dcf79d47652c0667ea6b1f3108b77a"
	hash4, _ = chainhash.NewHashFromStr(txId4)
	dest4    = "2MwU72uP9DWeXxPoq4VBRPH4UkDkH2zkhah"
	txId5    = "74d39e87c810a80faff70dcbd988c661dbe283a27f903cd587ab9c0b221cc602"
	hash5, _ = chainhash.NewHashFromStr(txId5)
	dest5    = "tb1q9mncjrazn5xgqdcqyjc0q0vzaytx2uzfc69q0x"
	txs      = map[string]*btcjson.TxRawResult{
		txId1: {
			Txid: txId1,
			Hash: hash1.String(),
			Vout: []btcjson.Vout{{
				Value: btcutil.Amount(1).ToBTC(), ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{dest1}},
			}},
			Confirmations: 9,
		},
		txId2: {
			Txid: txId2,
			Hash: hash2.String(),
			Vout: []btcjson.Vout{{
				Value: btcutil.Amount(2).ToBTC(), ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{dest2}},
			}},
			Confirmations: 17,
		},
		txId3: {
			Txid: txId3,
			Hash: hash3.String(),
			Vout: []btcjson.Vout{{
				Value: btcutil.Amount(3).ToBTC(), ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{dest3}},
			}},
			Confirmations: 9,
		},
		txId4: {
			Txid: txId4,
			Hash: hash4.String(),
			Vout: []btcjson.Vout{{
				Value: btcutil.Amount(4).ToBTC(), ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{dest4}},
			}},
			Confirmations: 8,
		},
		txId5: {
			Txid: txId5,
			Hash: hash5.String(),
			Vout: []btcjson.Vout{{
				Value: btcutil.Amount(5).ToBTC(), ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{dest5}},
			}},
			Confirmations: 12,
		}}
)

func prepareMsgs() []sdk.Msg {
	tx1 := btcTypes.NewMsgVerifyTx(sdk.AccAddress("user1"), hash1, 0, btcTypes.BtcAddress{
		Chain:         "testnet3",
		EncodedString: dest1,
	}, btcutil.Amount(1))

	tx2 := btcTypes.NewMsgVerifyTx(sdk.AccAddress("user1"), hash2, 0, btcTypes.BtcAddress{
		Chain:         "testnet3",
		EncodedString: dest2,
	}, btcutil.Amount(2))

	tx3 := btcTypes.NewMsgVerifyTx(sdk.AccAddress("user1"), hash3, 0, btcTypes.BtcAddress{
		Chain:         "testnet3",
		EncodedString: dest3,
	}, btcutil.Amount(3))

	tx4 := btcTypes.NewMsgVerifyTx(sdk.AccAddress("user1"), hash4, 0, btcTypes.BtcAddress{
		Chain:         "testnet3",
		EncodedString: dest4,
	}, btcutil.Amount(4))

	tx5 := btcTypes.NewMsgVerifyTx(sdk.AccAddress("user1"), hash5, 0, btcTypes.BtcAddress{
		Chain:         "testnet3",
		EncodedString: dest5,
	}, btcutil.Amount(5))

	return []sdk.Msg{tx1, tx2, tx3, tx4, tx5}
}

func Test_3Validators_VoteOn5Tx_Agree(t *testing.T) {

	// setting up the test infrastructure

	vAddr1 := sdk.ValAddress("val1")
	vAddr2 := sdk.ValAddress("val2")
	vAddr3 := sdk.ValAddress("val3")
	val1 := axMock.NewTestValidator(vAddr1, 100)
	val2 := axMock.NewTestValidator(vAddr2, 80)
	val3 := axMock.NewTestValidator(vAddr3, 170)
	staker := axMock.NewTestStaker(val1, val2, val3)

	b1 := sdk.AccAddress("broadcaster1")
	b2 := sdk.AccAddress("broadcaster2")
	b3 := sdk.AccAddress("broadcaster3")

	blockChain := mock.NewBlockchain().WithBlockSize(2).WithBlockTimeOut(100 * time.Millisecond)

	node1 := newNode(val1, b1, staker, blockChain.Input())
	node2 := newNode(val2, b2, staker, blockChain.Input())
	node3 := newNode(val3, b3, staker, blockChain.Input())

	blockChain.AddNodes(node1, node2, node3)
	blockChain.Start()

	in := blockChain.Input()
	defer close(in)

	// test begin

	in <- broadcastTypes.NewMsgRegisterProxy(val1.GetOperator(), b1)
	in <- broadcastTypes.NewMsgRegisterProxy(val2.GetOperator(), b2)
	in <- broadcastTypes.NewMsgRegisterProxy(val3.GetOperator(), b3)

	for _, msg := range prepareMsgs() {
		in <- msg
	}

	// TODO: implement meaningful assertions instead of just checking the log output while the test is running
	time.Sleep(10 * time.Minute)
}

func newNode(val exported.ValidatorI, broadcasterAddr sdk.AccAddress, staker axTypes.Staker, blockchainIn chan<- sdk.Msg) mock.Node {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	// register the types needed for marshaling, do not forget sdk.Msg!
	cdc := codec.New()
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	axTypes.RegisterCodec(cdc)
	btcTypes.RegisterCodec(cdc)

	broadcaster := mock.NewBroadcaster(cdc, broadcasterAddr, val.GetOperator(), blockchainIn)
	axK := keeper.NewKeeper(cdc, sdk.NewKVStoreKey(axTypes.StoreKey), staker, broadcaster)
	axK.SetVotingInterval(ctx, 10)
	axK.SetVotingThreshold(ctx, axTypes.VotingThreshold{Numerator: 2, Denominator: 3})
	axH := axelar.NewHandler(axK)

	btcK := btcKeeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey(btcTypes.StoreKey))
	btcH := btc_bridge.NewHandler(btcK, axK, &btcMock.TestRPC{RawTxs: txs}, nil)

	broadcastH := broadcast.NewHandler(broadcaster)

	eb := func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
		return axelar.EndBlocker(ctx, req, axK)
	}
	return mock.NewNode(val.GetOperator().String(), ctx).
		WithHandler(axTypes.ModuleName, axH).
		WithHandler(btcTypes.ModuleName, btcH).
		WithHandler(broadcastTypes.ModuleName, broadcastH).
		WithEndBlockers(eb)
}
