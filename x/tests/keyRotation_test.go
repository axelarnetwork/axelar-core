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
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

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
// 11. Sign a consolidation transaction (wait for vote)
// 12. Send the signed transaction to bitcoin
// 13. Query transfer tx info
// 14. Verify the fund transfer is confirmed on bitcoin (wait for vote)
// 15. Rotate to the new master key
func TestKeyRotation(t *testing.T) {

	const nodeCount = 10
	validators := make([]staking.Validator, 0, nodeCount)

	chain := fake.NewBlockchain().WithBlockTimeOut(10 * time.Millisecond)

	stringGen := testutils.RandStrings(5, 50).Distinct()
	defer stringGen.Stop()

	mocks := createMocks(&validators)

	var nodes []fake.Node
	for i, valAddr := range stringGen.Take(nodeCount) {
		validator := staking.Validator{
			OperatorAddress: sdk.ValAddress(valAddr),
			Tokens:          sdk.TokensFromConsensusPower(testutils.RandIntBetween(100, 1000)),
			Status:          sdk.Bonded,
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
	res := <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender(validators, nodeCount)})
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
		Sender:    randomSender(validators, nodeCount),
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
		Sender: randomSender(validators[:], nodeCount),
		Chain:  btc.Bitcoin.Name,
		KeyID:  masterKeyID1,
	})
	assert.NoError(t, res.Error)

	// rotate to the first master key
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender(validators[:], nodeCount),
		Chain:  btc.Bitcoin.Name,
	})
	assert.NoError(t, res.Error)

	// get deposit address for ethereum transfer
	ethAddr := nexus.CrossChainAddress{Chain: eth.Ethereum, Address: testutils.RandStringBetween(5, 20)}
	res = <-chain.Submit(btcTypes.NewMsgLink(randomSender(validators[:], nodeCount), ethAddr.Address, ethAddr.Chain.Name))
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
	res = <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender(validators, nodeCount), info))
	assert.NoError(t, res.Error)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// second snapshot
	res = <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender(validators, nodeCount)})
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
		Sender:    randomSender(validators, nodeCount),
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
		Sender: randomSender(validators[:], nodeCount),
		Chain:  btc.Bitcoin.Name,
		KeyID:  keyID2,
	})
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
	res = <-chain.Submit(btcTypes.NewMsgSign(randomSender(validators[:], nodeCount), btcutil.Amount(testutils.RandIntBetween(1, int64(amount)))))
	assert.NoError(t, res.Error)
	// assert tssd was properly called
	<-closeTimeout.Done()
	assert.Equal(t, nodeCount, len(mocks.Sign.CloseSendCalls()))

	// wait for voting to be done
	chain.WaitNBlocks(22)

	// send tx to Bitcoin
	bz, err = nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.SendTx},
		abci.RequestQuery{Data: nil})
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
	script, _ := btcTypes.CreateMasterRedeemScript(btcec.PublicKey(masterKey2.PublicKey))
	consolidationAddr, _ := btcTypes.CreateDepositAddress(mocks.BTC.Network(), script)
	mocks.BTC.GetOutPointInfoFunc = func(_ *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
		if out.String() == transferOut.String() {
			return btcTypes.OutPointInfo{
				OutPoint:      transferOut,
				BlockHash:     blockHash,
				Amount:        amount,
				Address:       consolidationAddr.EncodeAddress(),
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
		btcTypes.NewMsgVerifyTx(randomSender(validators, nodeCount), info))
	assert.NoError(t, res.Error)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// rotate master key to key 2
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender(validators[:], nodeCount),
		Chain:  btc.Bitcoin.Name,
	})
	assert.NoError(t, res.Error)
}
