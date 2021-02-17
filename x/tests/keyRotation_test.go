package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/utils/denom"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
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

	// set up chain
	const nodeCount = 10

	stringGen := testutils.RandStrings(5, 50).Distinct()
	defer stringGen.Stop()

	chain, validators, mocks, nodes := createChain(nodeCount, &stringGen)

	// register proxies for all validators
	registerProxies(chain, validators, nodeCount, &stringGen, t)
	takeSnapshot(chain, validators, nodeCount, t)

	// create master key for btc
	masterKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	mocks.Keygen.SendFunc = func(_ *tssd.MessageIn) error {
		return nil
	}
	mocks.Keygen.CloseSendFunc = func() error {
		return nil
	}
	mocks.Keygen.RecvFunc = func() (*tssd.MessageOut, error) {
		pk, _ := convert.PubkeyToBytes(masterKey.PublicKey)
		return &tssd.MessageOut{
			Data: &tssd.MessageOut_KeygenResult{KeygenResult: pk}}, nil
	}
	res, masterKeyID := createMasterKeyID(chain, validators, nodeCount, &stringGen, mocks)
	assert.NoError(t, res.Error)
	assert.Equal(t, nodeCount, len(mocks.Keygen.SendCalls()))
	assert.Equal(t, nodeCount, len(mocks.Keygen.CloseSendCalls()))

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// assign bitcoin master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: randomSender(validators, nodeCount),
		Chain:  balance.Bitcoin,
		KeyID:  masterKeyID,
	})
	assert.NoError(t, res.Error)

	// rotate to the first btc master key
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender(validators, nodeCount),
		Chain:  balance.Bitcoin,
	})
	assert.NoError(t, res.Error)

	// get deposit address for ethereum transfer
	crosschainAddr := balance.CrossChainAddress{Chain: balance.Bitcoin, Address: testutils.RandStringBetween(5, 20)}
	res = <-chain.Submit(btcTypes.NewMsgLink(randomSender(validators, nodeCount), crosschainAddr))
	assert.NoError(t, res.Error)
	depositAddr := string(res.Data)

	// simulate deposit to master key address
	blockHash, expectedOut, outPointInfo := sendBTCtoDepositAddress(depositAddr, mocks)

	// query for deposit info
	bz, err := nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.QueryOutInfo, blockHash.String()}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(expectedOut)})
	assert.NoError(t, err)
	var info btcTypes.OutPointInfo
	testutils.Codec().MustUnmarshalJSON(bz, &info)

	// verify deposit to master key
	res = <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender(validators, nodeCount), info))
	assert.NoError(t, res.Error)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// second snapshot
	res = <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender(validators, nodeCount)})
	assert.NoError(t, res.Error)

	// create another master key
	res, keyID2 := createMasterKeyID(chain, validators, nodeCount, &stringGen, mocks)
	assert.NoError(t, res.Error)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// assign second key to be the second master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: randomSender(validators, nodeCount),
		Chain:  balance.Bitcoin,
		KeyID:  keyID2,
	})
	assert.NoError(t, res.Error)

	// get consolidation address
	bz, err = nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.QueryConsolidationAddress, depositAddr}, abci.RequestQuery{})
	assert.NoError(t, err)
	consAddr := string(bz)

	// create a tx to transfer funds from deposit address to consolidation address
	prevAmount := int64(outPointInfo.Amount)
	amount := btcutil.Amount(prevAmount - testutils.RandIntBetween(1, prevAmount-1))

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

	// sign transaction
	msgToSign := make(chan []byte, nodeCount)
	mocks.Sign.SendFunc = func(messageIn *tssd.MessageIn) error {
		assert.Equal(t, masterKeyID, messageIn.GetSignInit().KeyUid)
		msgToSign <- messageIn.GetSignInit().MessageToSign
		return nil
	}
	sigChan := make(chan []byte, 1)
	go func() {
		r, s, err := ecdsa.Sign(rand.Reader, masterKey, <-msgToSign)
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
	closeTimeout, closeCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	mocks.Sign.CloseSendFunc = func() error {
		if len(mocks.Sign.CloseSendCalls()) == nodeCount {
			closeCancel()
		}
		return nil
	}

	// sign transfer tx
	res = <-chain.Submit(btcTypes.NewMsgSignTx(
		randomSender(validators, int64(nodeCount)),
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
	voutIdx := uint32(0)
	confirmations := uint64(testutils.RandIntBetween(1, 10000))
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
	res = <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender(validators, nodeCount), info))
	assert.NoError(t, res.Error)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// rotate master key to key 2
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender(validators, nodeCount),
		Chain:  balance.Bitcoin,
	})
	assert.NoError(t, res.Error)
}
