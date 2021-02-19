package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing"

	"github.com/axelarnetwork/axelar-core/utils/denom"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	balanceKeeper "github.com/axelarnetwork/axelar-core/x/balance/keeper"
	balanceTypes "github.com/axelarnetwork/axelar-core/x/balance/types"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"

	"github.com/axelarnetwork/axelar-core/store"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
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
	BTC     *btcMock.RPCClientMock
	Keygen  *tssdMock.TSSDKeyGenClientMock
	Sign    *tssdMock.TSSDSignClientMock
	Staker  *snapMock.StakingKeeperMock
	TSSD    *tssdMock.TSSDClientMock
	Slasher *snapMock.SlasherMock
}

// Testing the key rotation functionality.
// (0. Register proxies for all validators)
//  1. Create an initial validator snapshot
//  2. Create a key (wait for vote)
//  3. Designate that key to be the first master key for bitcoin
//  4. Rotate to the designated master key
//  5. Simulate bitcoin deposit to the current master key
//  6. Query deposit tx info
//  7. Verify the deposit is confirmed on bitcoin (wait for vote)
//  8. Create a second snapshot
//  9. Create a new key with the second snapshot's validator set (wait for vote)
// 10. Designate that key to be the next master key for bitcoin
// 11. Create a raw tx to transfer funds from the first master key address to the second key's address
// 12. Sign the raw tx with the OLD snapshot's validator set (wait for vote)
// 13. Send the signed transaction to bitcoin
// 14. Query transfer tx info
// 15. Verify the fund transfer is confirmed on bitcoin (wait for vote)
// 16. Rotate to the new master key
func TestKeyRotation(t *testing.T) {
	chain := fake.NewBlockchain().WithBlockTimeOut(10 * time.Millisecond)

	stringGen := testutils.RandStrings(5, 50).Distinct()
	defer stringGen.Stop()

	mocks := createMocks()

	var nodes []fake.Node
	for i, valAddr := range stringGen.Take(nodeCount) {
		validator := staking.Validator{
			OperatorAddress: sdk.ValAddress(valAddr),
			Tokens:          sdk.TokensFromConsensusPower(testutils.RandIntBetween(100, 1000)),
			Status:          sdk.Bonded,
			ConsPubKey:      ed25519.GenPrivKey().PubKey(),
		}
		validators = append(validators, validator)
		nodes = append(nodes, newNode("node"+strconv.Itoa(i), validator.OperatorAddress, mocks, chain))
		chain.AddNodes(nodes[i])
	}
	// Check to suppress any nil warnings from IDEs
	if nodes == nil {
		panic("need at least one node")
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
	res := <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender()})
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
		Sender:    randomSender(),
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
	chain.WaitNBlocks(12)

	// assign key as bitcoin master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: randomSender(),
		Chain:  btc.Bitcoin.Name,
		KeyID:  masterKeyID1,
	})
	assert.NoError(t, res.Error)

	// rotate to the first master key
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender(),
		Chain:  btc.Bitcoin.Name,
	})
	assert.NoError(t, res.Error)

	// get deposit address for ethereum transfer
	ethAddr := balance.CrossChainAddress{Chain: eth.Ethereum, Address: testutils.RandStringBetween(5, 20)}
	res = <-chain.Submit(btcTypes.NewMsgLink(randomSender(), ethAddr.Address, ethAddr.Chain.Name))
	assert.NoError(t, res.Error)
	depositAddr := string(res.Data)

	// simulate deposit to master key address
	txHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}

	voutIdx := uint32(testutils.RandIntBetween(0, 100))
	expectedOut := wire.NewOutPoint(txHash, voutIdx)
	amount := btcutil.Amount(testutils.RandIntBetween(1, 10000000))
	confirmations := uint64(testutils.RandIntBetween(1, 10000))
	mocks.BTC.GetOutPointInfoFunc = func(bHash *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
		if bHash.String() == blockHash.String() && out.String() == expectedOut.String() {
			return btcTypes.OutPointInfo{
				OutPoint:      expectedOut,
				BlockHash:     blockHash,
				Amount:        amount,
				Address:       depositAddr,
				Confirmations: confirmations,
			}, nil
		}

		return btcTypes.OutPointInfo{}, fmt.Errorf("tx %s not found", out.String())
	}

	// query for deposit info
	bz, err := nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.QueryOutInfo, blockHash.String()}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(expectedOut)})
	assert.NoError(t, err)
	var info btcTypes.OutPointInfo
	testutils.Codec().MustUnmarshalJSON(bz, &info)

	// verify deposit to master key
	res = <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender(), info))
	assert.NoError(t, res.Error)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// second snapshot
	res = <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender()})
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
		Sender:    randomSender(),
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
	chain.WaitNBlocks(12)

	// assign second key to be the second master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: randomSender(),
		Chain:  btc.Bitcoin.Name,
		KeyID:  keyID2,
	})
	assert.NoError(t, res.Error)

	// get consolidation address
	bz, err = nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.QueryConsolidationAddress, depositAddr}, abci.RequestQuery{})
	assert.NoError(t, err)
	consAddr := string(bz)

	// create a tx to transfer funds from deposit address to consolidation address
	amount = btcutil.Amount(int64(amount) - testutils.RandIntBetween(1, int64(amount)-1))

	bz, err = nodes[0].Query(
		[]string{btcTypes.QuerierRoute, btcKeeper.QueryRawTx},
		abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(
			btcTypes.RawTxParams{
				OutPoint:    expectedOut,
				DepositAddr: consAddr,
				Satoshi:     sdk.NewInt64Coin(denom.Sat, int64(amount)),
			})},
	)
	assert.NoError(t, err)
	var rawTx *wire.MsgTx
	testutils.Codec().MustUnmarshalJSON(bz, &rawTx)

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
	res = <-chain.Submit(btcTypes.NewMsgSignTx(
		randomSender(),
		expectedOut,
		rawTx))
	assert.NoError(t, res.Error)
	// assert tssd was properly called
	<-closeTimeout.Done()
	assert.Equal(t, nodeCount, len(mocks.Sign.CloseSendCalls()))

	// wait for voting to be done
	chain.WaitNBlocks(22)

	// send tx to Bitcoin
	bz, err = nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.SendTx},
		abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(expectedOut)})
	assert.NoError(t, err)

	// set up btc mock to return the new tx
	var transferHash *chainhash.Hash
	testutils.Codec().MustUnmarshalJSON(bz, &transferHash)
	blockHash, err = chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	voutIdx = 0
	confirmations = uint64(testutils.RandIntBetween(1, 10000))
	transferOut := wire.NewOutPoint(transferHash, voutIdx)
	mocks.BTC.GetOutPointInfoFunc = func(_ *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
		if out.String() == transferOut.String() {
			return btcTypes.OutPointInfo{
				OutPoint:      transferOut,
				BlockHash:     blockHash,
				Amount:        amount,
				Address:       consAddr,
				Confirmations: confirmations,
			}, nil
		}

		return btcTypes.OutPointInfo{}, fmt.Errorf("tx %s not found", out.String())
	}

	// query for transfer info
	bz, err = nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.QueryOutInfo, blockHash.String()}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(transferOut)})
	assert.NoError(t, err)
	testutils.Codec().MustUnmarshalJSON(bz, &info)

	// verify master key transfer
	res = <-chain.Submit(
		btcTypes.NewMsgVerifyTx(randomSender(), info))
	assert.NoError(t, res.Error)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// rotate master key to key 2
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender(),
		Chain:  btc.Bitcoin.Name,
	})
	assert.NoError(t, res.Error)
}

func randomSender() sdk.AccAddress {
	return sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress)
}

func newNode(moniker string, validator sdk.ValAddress, mocks testMocks, chain *fake.BlockChain) fake.Node {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	broadcaster := fake.NewBroadcaster(testutils.Codec(), validator, chain.Submit)

	snapSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snap")

	snapKeeper := snapshotKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(snapTypes.StoreKey), snapSubspace, mocks.Staker, mocks.Slasher)
	snapKeeper.SetParams(ctx, snapTypes.DefaultParams())

	voter := voteKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(voteTypes.StoreKey), store.NewSubjectiveStore(), snapKeeper, broadcaster)

	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	bitcoinKeeper := btcKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(btcTypes.StoreKey), btcSubspace)
	btcParams := btcTypes.DefaultParams()
	btcParams.Network = mocks.BTC.Network()
	bitcoinKeeper.SetParams(ctx, btcParams)

	signer := tssKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(tssTypes.StoreKey), mocks.TSSD,
		params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), tssTypes.DefaultParamspace),
		voter, broadcaster,
	)
	signer.SetParams(ctx, tssTypes.DefaultParams())

	balanceSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("balanceKey"), sdk.NewKVStoreKey("tbalanceKey"), "balance")
	balancer := balanceKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(balanceTypes.StoreKey), balanceSubspace)
	balancer.SetParams(ctx, balanceTypes.DefaultParams())

	voter.SetVotingInterval(ctx, voteTypes.DefaultGenesisState().VotingInterval)
	voter.SetVotingThreshold(ctx, voteTypes.DefaultGenesisState().VotingThreshold)

	router := fake.NewRouter()

	broadcastHandler := broadcast.NewHandler(broadcaster)
	btcHandler := bitcoin.NewHandler(bitcoinKeeper, voter, mocks.BTC, signer, snapKeeper, balancer)
	snapHandler := snapshot.NewHandler(snapKeeper)
	tssHandler := tss.NewHandler(signer, snapKeeper, balancer, voter, mocks.Staker)
	voteHandler := vote.NewHandler()

	router = router.
		AddRoute(broadcastTypes.RouterKey, broadcastHandler).
		AddRoute(btcTypes.RouterKey, btcHandler).
		AddRoute(snapTypes.RouterKey, snapHandler).
		AddRoute(voteTypes.RouterKey, voteHandler).
		AddRoute(tssTypes.RouterKey, tssHandler)

	queriers := map[string]sdk.Querier{btcTypes.QuerierRoute: btcKeeper.NewQuerier(bitcoinKeeper, signer, balancer, mocks.BTC)}

	node := fake.NewNode(moniker, ctx, router, queriers).
		WithEndBlockers(func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
			return vote.EndBlocker(ctx, req, voter)
		})
	return node
}

func createMocks() testMocks {

	slasher := &snapMock.SlasherMock{
		GetValidatorSigningInfoFunc: func(ctx sdk.Context, address sdk.ConsAddress) (types.ValidatorInfo, bool) {
			newInfo := slashingtypes.NewValidatorSigningInfo(
				address,
				int64(0),        // height at which validator was first a candidate OR was unjailed
				int64(3),        // index offset into signed block bit array. TODO: check if needs to be set correctly.
				time.Unix(0, 0), // jailed until
				false,           // tomstoned
				int64(0),        // missed blocks
			)
			retinfo := types.ValidatorInfo{newInfo}
			return retinfo, true
		},
	}

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
		SendRawTransactionFunc: func(tx *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
			hash := tx.TxHash()
			return &hash, nil
		},
		NetworkFunc: func() btcTypes.Network { return btcTypes.Mainnet }}

	keygen := &tssdMock.TSSDKeyGenClientMock{}
	sign := &tssdMock.TSSDSignClientMock{}
	tssdClient := &tssdMock.TSSDClientMock{
		KeygenFunc: func(context.Context, ...grpc.CallOption) (tssd.GG18_KeygenClient, error) { return keygen, nil },
		SignFunc:   func(context.Context, ...grpc.CallOption) (tssd.GG18_SignClient, error) { return sign, nil }}
	return testMocks{
		BTC:     btcClient,
		TSSD:    tssdClient,
		Keygen:  keygen,
		Sign:    sign,
		Staker:  stakingKeeper,
		Slasher: slasher,
	}
}
