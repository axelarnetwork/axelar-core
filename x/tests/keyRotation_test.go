package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

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
	"github.com/axelarnetwork/axelar-core/testutils/mock"
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

	// result of the keygen
	sk, _   = ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sigLock = &sync.Mutex{}
	// result of the sign, ensure signature exists when queried to return
	sig []byte

	// return from the bitcoin rpc client when queried by txHash
	btcTx *btcjson.TxRawResult
)

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
	chain := mock.NewBlockchain().WithBlockTimeOut(10 * time.Millisecond)

	stringGen := testutils.RandStrings(5, 50).Distinct()
	defer stringGen.Stop()

	for i, valAddr := range stringGen.Take(nodeCount) {
		validator := staking.Validator{
			OperatorAddress: sdk.ValAddress(valAddr),
			Tokens:          sdk.TokensFromConsensusPower(testutils.RandIntBetween(100, 1000)),
			Status:          sdk.Bonded,
		}
		validators = append(validators, validator)
		node := newNode("node"+strconv.Itoa(i), validator.OperatorAddress, chain)
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

	// create first key
	keyID := stringGen.Next()
	res = <-chain.Submit(tssTypes.MsgKeygenStart{
		Sender:    sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		NewKeyID:  keyID,
		Threshold: int(testutils.RandIntBetween(1, int64(len(validators)))),
	})
	assert.NoError(t, res.Error)

	// wait for voting to be done
	<-chain.WaitNBlocks(12)

	// assign key as bitcoin master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		Chain:  "bitcoin",
		KeyID:  keyID,
	})
	assert.NoError(t, res.Error)

	// rotate to the first master key
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		Chain:  "bitcoin",
	})
	assert.NoError(t, res.Error)

	// track bitcoin transactions for address derived from master key
	res = <-chain.Submit(btcTypes.NewMsgTrackPubKeyWithMasterKey(
		sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		btcTypes.Chain(chaincfg.MainNetParams.Name), false))
	assert.NoError(t, res.Error)

	// simulate deposit to master key address
	txHash, err := chainhash.NewHash([]byte(testutils.RandString(32)))
	if err != nil {
		panic(err)
	}

	btcAddr, err := btcTypes.ParseBtcAddress(string(res.Data), btcTypes.Chain(chaincfg.MainNetParams.Name))
	if err != nil {
		panic(err)
	}

	vout := make([]btcjson.Vout, testutils.RandIntBetween(1, 100))
	voutIdx := testutils.RandIntBetween(0, int64(len(vout)))
	amount := btcutil.Amount(testutils.RandIntBetween(1, 10000000))
	vout[voutIdx] = btcjson.Vout{
		N:            uint32(voutIdx),
		Value:        amount.ToBTC(),
		ScriptPubKey: btcjson.ScriptPubKeyResult{Addresses: []string{btcAddr.String()}},
	}

	btcTx = &btcjson.TxRawResult{
		Txid:          txHash.String(),
		Hash:          txHash.String(),
		Vout:          vout,
		Confirmations: uint64(testutils.RandIntBetween(1, 10000)),
	}

	// verify deposit to master key
	res = <-chain.Submit(btcTypes.NewMsgVerifyTx(
		sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		txHash,
		uint32(voutIdx),
		btcAddr,
		amount))
	assert.NoError(t, res.Error)

	// wait for voting to be done
	<-chain.WaitNBlocks(12)

	// second snapshot
	res = <-chain.Submit(snapTypes.MsgSnapshot{Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress)})
	assert.NoError(t, res.Error)

	// second keygen with validator set of second snapshot
	keyID2 := stringGen.Next()
	res = <-chain.Submit(tssTypes.MsgKeygenStart{
		Sender:    sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		NewKeyID:  keyID2,
		Threshold: int(testutils.RandIntBetween(1, int64(len(validators)))),
	})
	assert.NoError(t, res.Error)

	// wait for voting to be done
	<-chain.WaitNBlocks(12)

	// assign second key to be the second master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		Chain:  "bitcoin",
		KeyID:  keyID2,
	})
	assert.NoError(t, res.Error)

	// create a tx to transfer funds from master key 1 to master key 2
	res = <-chain.Submit(btcTypes.NewMsgRawTxForMasterKey(
		sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		btcTypes.Chain(chaincfg.MainNetParams.Name),
		txHash,
		btcutil.Amount(int64(amount)-testutils.RandIntBetween(1, int64(amount)-1))))
	assert.NoError(t, res.Error)

	// sign transfer tx
	sigID := stringGen.Next()
	res = <-chain.Submit(tssTypes.MsgMasterKeySignStart{
		Sender:    sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		NewSigID:  sigID,
		Chain:     "bitcoin",
		MsgToSign: res.Data,
	})
	assert.NoError(t, res.Error)

	// wait for voting to be done
	<-chain.WaitNBlocks(12)

	// execute transfer tx -> will be set as new return for rpc query by hash
	res = <-chain.Submit(
		btcTypes.NewMsgTransferToNewMasterKey(sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
			txHash.String(),
			sigID))
	assert.NoError(t, res.Error)

	// TODO: add functionality to verify master key transfer

	// rotate master key to key 2
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: sdk.AccAddress(validators[testutils.RandIntBetween(0, nodeCount)].OperatorAddress),
		Chain:  "bitcoin",
	})
	assert.NoError(t, res.Error)
}

func newNode(moniker string, validator sdk.ValAddress, chain mock.BlockChain) mock.Node {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	stakingKeeper := &snapMock.StakingKeeperMock{
		IterateValidatorsFunc: func(ctx sdk.Context, fn func(index int64, validator sdkExported.ValidatorI) (stop bool)) {
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
		GetRawTransactionVerboseFunc: func(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
			if hash.String() == btcTx.Hash {
				return btcTx, nil
			}
			return nil, fmt.Errorf("tx %s not found", hash.String())
		},
		SendRawTransactionFunc: func(*wire.MsgTx, bool) (*chainhash.Hash, error) { return nil, nil }}

	toSign := make(chan []byte, 1)
	tssdClient := &tssdMock.TSSDClientMock{
		KeygenFunc: func(context.Context, ...grpc.CallOption) (tssd.GG18_KeygenClient, error) {
			return &tssdMock.TSSDKeyGenClientMock{
				RecvFunc: func() (*tssd.MessageOut, error) {
					pk, _ := convert.PubkeyToBytes(sk.PublicKey)
					return &tssd.MessageOut{
						Data: &tssd.MessageOut_KeygenResult{KeygenResult: pk}}, nil
				},
				SendFunc:      func(*tssd.MessageIn) error { return nil },
				CloseSendFunc: func() error { return nil },
			}, nil
		},
		SignFunc: func(context.Context, ...grpc.CallOption) (tssd.GG18_SignClient, error) {
			return &tssdMock.TSSDSignClientMock{
				SendFunc: func(messageIn *tssd.MessageIn) error {
					msg := messageIn.Data.(*tssd.MessageIn_SignInit)
					toSign <- msg.SignInit.MessageToSign
					return nil
				},
				RecvFunc: func() (*tssd.MessageOut, error) {
					r, s, err := ecdsa.Sign(rand.Reader, sk, <-toSign)
					if err != nil {
						panic(err)
					}

					sigBz, err := convert.SigToBytes(r.Bytes(), s.Bytes())
					if err != nil {
						panic(err)
					}

					sigLock.Lock()
					if sig == nil {
						sig = sigBz
					}
					sigLock.Unlock()
					return &tssd.MessageOut{
						Data: &tssd.MessageOut_SignResult{SignResult: sig}}, nil
				},
				CloseSendFunc: func() error { return nil },
			}, nil
		}}

	broadcaster := mock.NewBroadcaster(testutils.Codec(), validator, chain.Submit)

	snapKeeper := snapshotKeeper.NewKeeper(testutils.Codec(), mock.NewKVStoreKey(snapTypes.StoreKey), stakingKeeper)
	vKeeper := voteKeeper.NewKeeper(testutils.Codec(), mock.NewKVStoreKey(voteTypes.StoreKey), store.NewSubjectiveStore(), snapKeeper, broadcaster)
	btcKeeper := bitcoinKeeper.NewBtcKeeper(testutils.Codec(), mock.NewKVStoreKey(btcTypes.StoreKey))
	tKeeper := tssKeeper.NewKeeper(testutils.Codec(), mock.NewKVStoreKey(tssTypes.StoreKey), tssdClient,
		params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), tssTypes.DefaultParamspace),
		broadcaster,
	)

	vKeeper.SetVotingInterval(ctx, voteTypes.DefaultGenesisState().VotingInterval)
	vKeeper.SetVotingThreshold(ctx, voteTypes.DefaultGenesisState().VotingThreshold)

	tKeeper.SetParams(ctx, tssTypes.DefaultParams())
	router := mock.NewRouter()

	broadcastHandler := broadcast.NewHandler(broadcaster)
	btcHandler := bitcoin.NewHandler(btcKeeper, vKeeper, btcClient, tKeeper)
	snapHandler := snapshot.NewHandler(snapKeeper)
	tssHandler := tss.NewHandler(tKeeper, snapKeeper, vKeeper)
	voteHandler := vote.NewHandler(vKeeper, router)

	router = router.
		AddRoute(broadcastTypes.RouterKey, broadcastHandler).
		AddRoute(btcTypes.RouterKey, btcHandler).
		AddRoute(snapTypes.RouterKey, snapHandler).
		AddRoute(voteTypes.RouterKey, voteHandler).
		AddRoute(tssTypes.RouterKey, tssHandler)

	node := mock.NewNode(moniker, ctx, router).
		WithEndBlockers(func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
			return vote.EndBlocker(ctx, req, vKeeper)
		})
	return node
}
