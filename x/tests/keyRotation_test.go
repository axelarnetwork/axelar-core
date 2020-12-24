package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strconv"
	"testing"
	"time"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	balanceKeeper "github.com/axelarnetwork/axelar-core/x/balance/keeper"
	balanceTypes "github.com/axelarnetwork/axelar-core/x/balance/types"
	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/store"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	bitcoinKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	"github.com/axelarnetwork/axelar-core/x/broadcast"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot"
	snapshotKeeper "github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
	"github.com/axelarnetwork/axelar-core/x/tss"
	tssKeeper "github.com/axelarnetwork/axelar-core/x/tss/keeper"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	tssdMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	"github.com/axelarnetwork/axelar-core/x/vote"
	voteKeeper "github.com/axelarnetwork/axelar-core/x/vote/keeper"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

const nodeCount = 10

// globally available storage variables to control the behaviour of the mocks
var (
	// set of validators known to the staking keeper
	validators = make([]staking.Validator, 0, nodeCount)
)

type testMocks struct {
	BTC    *btcMock.RPCClientMock
	Keygen *tssdMock.TSSDKeyGenClientMock
	Sign   *tssdMock.TSSDSignClientMock
	Staker *snapMock.StakingKeeperMock
	TSSD   *tssdMock.TSSDClientMock
}

// Testing the key rotation functionality.
// (0. Register proxies for all validators)
//  1. Create an initial validator snapshot
//  2. Create a key
//  3. Designate that key to be the first master key for bitcoin
//  4. Rotate to the designated master key
//  5. Track the bitcoin address corresponding to the master key
//  6. Simulate bitcoin deposit to the current master key
//  7. Verify the deposit is confirmed on bitcoin
//  8. Create a second snapshot
//  9. Create a new key with the second snapshot's validator set
// 10. Designate that key to be the next master key for bitcoin
// 11. Create a raw tx to transfer funds from the first master key address to the second key's address
// 12. Sign the hash of the raw tx with the OLD snapshot's validator set
// 13. Send the signed transaction to bitcoin
// 14. Verify the fund transfer is confirmed on bitcoin
// 15. Rotate to the new master key
func TestKeyRotation(t *testing.T) {
	chain := fake.NewBlockchain().WithBlockTimeOut(10 * time.Millisecond)

	stringGen := testutils.RandStrings(5, 50).Distinct()
	defer stringGen.Stop()

	mocks := createMocks()

	for i, valAddr := range stringGen.Take(nodeCount) {
		validator := staking.Validator{
			OperatorAddress: sdk.ValAddress(valAddr),
			Tokens:          sdk.TokensFromConsensusPower(testutils.RandIntBetween(100, 1000)),
			Status:          sdk.Bonded,
		}
		validators = append(validators, validator)
		node := newNode("node"+strconv.Itoa(i), validator.OperatorAddress, mocks, chain)
		chain.AddNodes(node)
	}
	chain.Start()

	// register proxies
	for i := 0; i < nodeCount; i++ {
		res := <-chain.Submit(broadcastTypes.MsgRegisterProxy{
			Principal: validators[i].OperatorAddress,
			Proxy:     sdk.AccAddress(stringGen.Next()),
		})
		assert.NoError(t, res.Error)
	}

	// take first validator snapshot
	res := <-chain.Submit(snapTypes.MsgSnapshot{Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress)})
	assert.NoError(t, res.Error)

	// set up tssd mock for first keygen
	masterKey1, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	mocks.Keygen.RecvFunc = func() (*tssd.MessageOut, error) {
		pk, _ := convert.PubkeyToBytes(masterKey1.PublicKey)
		return &tssd.MessageOut{
			Data: &tssd.MessageOut_KeygenResult{KeygenResult: pk}}, nil
	}
	// ensure all nodes call .Send() and .CloseSend()
	sendTimeout, sendCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	closeTimeout, closeCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	mocks.Keygen.SendFunc = func(_ *tssd.MessageIn) error {
		if len(mocks.Keygen.SendCalls()) == nodeCount {
			sendCancel()
		}
		return nil
	}
	mocks.Keygen.CloseSendFunc = func() error {
		if len(mocks.Keygen.CloseSendCalls()) == nodeCount {
			closeCancel()
		}
		return nil
	}

	// create first key
	masterKeyID1 := stringGen.Next()
	res = <-chain.Submit(tssTypes.MsgKeygenStart{
		Sender:    sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		NewKeyID:  masterKeyID1,
		Threshold: int(testutils.RandIntBetween(1, int64(len(validators)))),
	})
	assert.NoError(t, res.Error)

	// assert tssd was properly called
	<-sendTimeout.Done()
	<-closeTimeout.Done()
	assert.Equal(t, nodeCount, len(mocks.Keygen.SendCalls()))
	assert.Equal(t, nodeCount, len(mocks.Keygen.CloseSendCalls()))

	// wait for voting to be done
	<-chain.WaitNBlocks(12)

	// assign key as bitcoin master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		Chain:  balance.Bitcoin,
		KeyID:  masterKeyID1,
	})
	assert.NoError(t, res.Error)

	// rotate to the first master key
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		Chain:  balance.Bitcoin,
	})
	assert.NoError(t, res.Error)

	// track bitcoin transactions for address derived from master key
	res = <-chain.Submit(btcTypes.NewMsgTrackPubKeyWithMasterKey(
		sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		btcTypes.Network(chaincfg.MainNetParams.Name), false))
	assert.NoError(t, res.Error)

	// simulate deposit to master key address
	prevSK, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	prevTxHash, err := chainhash.NewHash([]byte(testutils.RandString(32)))
	if err != nil {
		panic(err)
	}
	prevVoutIdx := uint32(testutils.RandIntBetween(0, 100))

	prevPK := btcec.PublicKey(prevSK.PublicKey)
	sender, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(prevPK.SerializeCompressed()), &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}

	txHash, err := chainhash.NewHash([]byte(testutils.RandString(32)))
	if err != nil {
		panic(err)
	}

	recipient, err := btcTypes.ParseBtcAddress(string(res.Data), btcTypes.Network(chaincfg.MainNetParams.Name))
	if err != nil {
		panic(err)
	}

	voutIdx := uint32(testutils.RandIntBetween(0, 100))
	amount := btcutil.Amount(testutils.RandIntBetween(1, 10000000))

	mocks.BTC.GetRawTransactionVerboseFunc = func(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
		if hash.IsEqual(txHash) {
			vin := make([]btcjson.Vin, 1)
			vin[0] = btcjson.Vin{Vout: prevVoutIdx, Txid: prevTxHash.String()}

			vout := make([]btcjson.Vout, testutils.RandIntBetween(int64(voutIdx)+1, 100))
			vout[voutIdx] = btcjson.Vout{
				N:            voutIdx,
				Value:        amount.ToBTC(),
				ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{recipient.String()}},
			}
			return &btcjson.TxRawResult{
				Txid:          txHash.String(),
				Hash:          txHash.String(),
				Vin:           vin,
				Vout:          vout,
				Confirmations: uint64(testutils.RandIntBetween(1, 10000)),
			}, nil
		}

		if hash.IsEqual(prevTxHash) {
			vout := make([]btcjson.Vout, testutils.RandIntBetween(int64(prevVoutIdx)+1, 100))
			vout[prevVoutIdx] = btcjson.Vout{
				N:            prevVoutIdx,
				ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{sender.String()}},
			}
			return &btcjson.TxRawResult{
				Txid:          prevTxHash.String(),
				Hash:          prevTxHash.String(),
				Vout:          vout,
				Confirmations: uint64(testutils.RandIntBetween(1, 10000)),
			}, nil
		}
		return nil, fmt.Errorf("tx %s not found", hash.String())
	}

	// verify deposit to master key
	res = <-chain.Submit(btcTypes.NewMsgVerifyTx(
		sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		txHash,
		voutIdx,
		btcTypes.BtcAddress{
			Network:       btcTypes.Network(chaincfg.MainNetParams.Name),
			EncodedString: sender.String(),
		},
		recipient,
		amount))
	assert.NoError(t, res.Error)

	// wait for voting to be done
	<-chain.WaitNBlocks(12)

	// second snapshot
	res = <-chain.Submit(snapTypes.MsgSnapshot{Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress)})
	assert.NoError(t, res.Error)

	// set up tssd mock for second keygen
	masterKey2, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	mocks.Keygen.RecvFunc = func() (*tssd.MessageOut, error) {
		pk, _ := convert.PubkeyToBytes(masterKey2.PublicKey)
		return &tssd.MessageOut{
			Data: &tssd.MessageOut_KeygenResult{KeygenResult: pk}}, nil
	}
	// ensure all nodes call .Send() and .CloseSend()
	sendTimeout, sendCancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	closeTimeout, closeCancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	mocks.Keygen.SendFunc = func(_ *tssd.MessageIn) error {
		if len(mocks.Keygen.SendCalls()) == 2*nodeCount {
			sendCancel()
		}
		return nil
	}
	mocks.Keygen.CloseSendFunc = func() error {
		if len(mocks.Keygen.CloseSendCalls()) == 2*nodeCount {
			closeCancel()
		}
		return nil
	}

	// second keygen with validator set of second snapshot
	keyID2 := stringGen.Next()
	res = <-chain.Submit(tssTypes.MsgKeygenStart{
		Sender:    sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		NewKeyID:  keyID2,
		Threshold: int(testutils.RandIntBetween(1, int64(len(validators)))),
	})
	assert.NoError(t, res.Error)

	// assert tssd was properly called
	<-sendTimeout.Done()
	<-closeTimeout.Done()
	assert.Equal(t, 2*nodeCount, len(mocks.Keygen.SendCalls()))
	assert.Equal(t, 2*nodeCount, len(mocks.Keygen.CloseSendCalls()))

	// wait for voting to be done
	<-chain.WaitNBlocks(12)

	// assign second key to be the second master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		Chain:  balance.Bitcoin,
		KeyID:  keyID2,
	})
	assert.NoError(t, res.Error)

	// create a tx to transfer funds from master key 1 to master key 2
	amount = btcutil.Amount(int64(amount) - testutils.RandIntBetween(1, int64(amount)-1))
	res = <-chain.Submit(btcTypes.NewMsgRawTxForNextMasterKey(
		sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		btcTypes.Network(chaincfg.MainNetParams.Name),
		txHash,
		amount))
	assert.NoError(t, res.Error)

	// set up tssd mock for signing
	msgToSign := make(chan []byte, nodeCount)
	mocks.Sign.SendFunc = func(messageIn *tssd.MessageIn) error {
		assert.Equal(t, masterKeyID1, messageIn.GetSignInit().KeyUid)
		msgToSign <- messageIn.GetSignInit().MessageToSign
		return nil
	}
	sigChan := make(chan []byte, 1)
	go func() {
		r, s, err := ecdsa.Sign(rand.Reader, masterKey1, <-msgToSign)
		if err != nil {
			panic(err)
		}
		sig, err := convert.SigToBytes(r.Bytes(), s.Bytes())
		if err != nil {
			panic(err)
		}
		sigChan <- sig
	}()
	mocks.Sign.RecvFunc = func() (*tssd.MessageOut, error) {
		sig := <-sigChan
		sigChan <- sig
		return &tssd.MessageOut{Data: &tssd.MessageOut_SignResult{SignResult: sig}}, nil
	}

	// ensure and .CloseSend() is called
	closeTimeout, closeCancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	mocks.Sign.CloseSendFunc = func() error {
		if len(mocks.Sign.CloseSendCalls()) == nodeCount {
			closeCancel()
		}
		return nil
	}

	// sign transfer tx
	sigID := stringGen.Next()
	res = <-chain.Submit(tssTypes.MsgSignStart{
		Sender:    sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		SigID:     sigID,
		Chain:     balance.Bitcoin,
		MsgToSign: res.Data,
		Mode:      tssTypes.ModeMasterKey,
	})
	assert.NoError(t, res.Error)

	// assert tssd was properly called
	<-closeTimeout.Done()
	assert.Equal(t, nodeCount, len(mocks.Sign.CloseSendCalls()))

	// wait for voting to be done
	<-chain.WaitNBlocks(12)

	// execute transfer tx
	res = <-chain.Submit(
		btcTypes.NewMsgSendTx(sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
			txHash.String(),
			sigID))
	assert.NoError(t, res.Error)

	// set up btc mock to return the new tx
	nextMasterPK := btcec.PublicKey(masterKey2.PublicKey)
	outAddr, _ := btcutil.NewAddressPubKey(nextMasterPK.SerializeUncompressed(), &chaincfg.MainNetParams)
	voutIdx = uint32(testutils.RandIntBetween(0, 100))
	transferTxHash := &chainhash.Hash{}
	assert.NoError(t, transferTxHash.SetBytes(res.Data))
	mocks.BTC.GetRawTransactionVerboseFunc = func(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
		if hash.IsEqual(transferTxHash) {
			vout := make([]btcjson.Vout, testutils.RandIntBetween(int64(voutIdx)+1, 100))
			vout[voutIdx] = btcjson.Vout{
				N:            voutIdx,
				Value:        amount.ToBTC(),
				ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{outAddr.String()}},
			}
			return &btcjson.TxRawResult{
				Txid:          transferTxHash.String(),
				Hash:          transferTxHash.String(),
				Vout:          vout,
				Confirmations: uint64(testutils.RandIntBetween(1, 10000)),
			}, nil
		}
		return nil, fmt.Errorf("tx %s not found", hash.String())
	}

	// verify master key transfer
	res = <-chain.Submit(
		btcTypes.NewMsgVerifyTxForNextMasterKey(sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
			transferTxHash, voutIdx, amount, btcTypes.Network(chaincfg.MainNetParams.Name)))
	assert.NoError(t, res.Error)

	// wait for voting to be done
	<-chain.WaitNBlocks(12)

	// rotate master key to key 2
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		Chain:  balance.Bitcoin,
	})
	assert.NoError(t, res.Error)
}

func newNode(moniker string, validator sdk.ValAddress, mocks testMocks, chain fake.BlockChain) fake.Node {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	broadcaster := fake.NewBroadcaster(testutils.Codec(), validator, chain.Submit)

	snapKeeper := snapshotKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(snapTypes.StoreKey), mocks.Staker)
	vKeeper := voteKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(voteTypes.StoreKey), store.NewSubjectiveStore(), snapKeeper, broadcaster)
	btcKeeper := bitcoinKeeper.NewBtcKeeper(testutils.Codec(), sdk.NewKVStoreKey(btcTypes.StoreKey))
	tKeeper := tssKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(tssTypes.StoreKey), mocks.TSSD,
		params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), tssTypes.DefaultParamspace),
		broadcaster,
	)
	balancer := balanceKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(balanceTypes.StoreKey))

	vKeeper.SetVotingInterval(ctx, voteTypes.DefaultGenesisState().VotingInterval)
	vKeeper.SetVotingThreshold(ctx, voteTypes.DefaultGenesisState().VotingThreshold)

	tKeeper.SetParams(ctx, tssTypes.DefaultParams())
	router := fake.NewRouter()

	broadcastHandler := broadcast.NewHandler(broadcaster)
	btcHandler := bitcoin.NewHandler(btcKeeper, vKeeper, mocks.BTC, tKeeper, balancer)
	snapHandler := snapshot.NewHandler(snapKeeper)
	tssHandler := tss.NewHandler(tKeeper, snapKeeper, vKeeper)
	voteHandler := vote.NewHandler()

	router = router.
		AddRoute(broadcastTypes.RouterKey, broadcastHandler).
		AddRoute(btcTypes.RouterKey, btcHandler).
		AddRoute(snapTypes.RouterKey, snapHandler).
		AddRoute(voteTypes.RouterKey, voteHandler).
		AddRoute(tssTypes.RouterKey, tssHandler)

	node := fake.NewNode(moniker, ctx, router).
		WithEndBlockers(func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
			return vote.EndBlocker(ctx, req, vKeeper)
		})
	return node
}

func createMocks() testMocks {
	stakingKeeper := &snapMock.StakingKeeperMock{
		IterateLastValidatorsFunc: func(ctx sdk.Context, fn func(index int64, validator sdkExported.ValidatorI) (stop bool)) {
			for j, val := range validators {
				if fn(int64(j), val) {
					break
				}
			}
		},
		GetLastTotalPowerFunc: func(ctx sdk.Context) sdk.Int {
			totalPower := sdk.ZeroInt()
			for _, val := range validators {
				totalPower = totalPower.AddRaw(val.ConsensusPower())
			}
			return totalPower
		},
	}

	btcClient := &btcMock.RPCClientMock{
		ImportAddressRescanFunc: func(string, string, bool) error { return nil },
		SendRawTransactionFunc: func(tx *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
			hash := tx.TxHash()
			return &hash, nil
		}}

	keygen := &tssdMock.TSSDKeyGenClientMock{}
	sign := &tssdMock.TSSDSignClientMock{}
	tssdClient := &tssdMock.TSSDClientMock{
		KeygenFunc: func(context.Context, ...grpc.CallOption) (tssd.GG18_KeygenClient, error) { return keygen, nil },
		SignFunc:   func(context.Context, ...grpc.CallOption) (tssd.GG18_SignClient, error) { return sign, nil }}
	return testMocks{
		BTC:    btcClient,
		TSSD:   tssdClient,
		Keygen: keygen,
		Sign:   sign,
		Staker: stakingKeeper,
	}
}
