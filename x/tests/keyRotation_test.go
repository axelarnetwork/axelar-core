package tests

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
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
func TestBitcoinKeyRotation(t *testing.T) {
	randStrings := testutils.RandStrings(5, 50)
	defer randStrings.Stop()

	// set up chain
	const nodeCount = 10
	chain, nodeData := initChain(nodeCount)

	// register proxies for all validators
	for i, proxy := range randStrings.Take(nodeCount) {
		res := <-chain.Submit(broadcastTypes.MsgRegisterProxy{Principal: nodeData[i].Validator.OperatorAddress, Proxy: sdk.AccAddress(proxy)})
		assert.NoError(t, res.Error)
	}

	// take the first snapshot
	snapshotResult1 := <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender()})
	assert.NoError(t, snapshotResult1.Error)

	// create master key for btc
	threshold1 := int(testutils.RandIntBetween(1, int64(nodeCount)))
	masterKeyID1 := randStrings.Next()
	masterKey1, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// prepare mocks with the master key
	var correctKeygens1 []<-chan bool
	for _, n := range nodeData {
		correctKeygens1 = append(correctKeygens1, prepareKeygen(n.Mocks.Keygen, masterKeyID1, masterKey1.PublicKey))
	}

	// start keygen
	keygenResult1 := <-chain.Submit(
		tssTypes.MsgKeygenStart{Sender: randomSender(), NewKeyID: masterKeyID1, Threshold: threshold1})
	assert.NoError(t, keygenResult1.Error)
	for _, isCorrect := range correctKeygens1 {
		assert.True(t, <-isCorrect)
	}

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// assign bitcoin master key
	assignKeyResult1 := <-chain.Submit(
		tssTypes.MsgAssignNextMasterKey{Sender: randomSender(), Chain: btc.Bitcoin.Name, KeyID: masterKeyID1})
	assert.NoError(t, assignKeyResult1.Error)

	// rotate to the first btc master key
	rotateResult1 := <-chain.Submit(
		tssTypes.MsgRotateMasterKey{Sender: randomSender(), Chain: btc.Bitcoin.Name})
	assert.NoError(t, rotateResult1.Error)

	// get deposit address for ethereum transfer
	crossChainAddr := nexus.CrossChainAddress{Chain: eth.Ethereum, Address: randStrings.Next()}
	linkResult := <-chain.Submit(btcTypes.NewMsgLink(randomSender(), crossChainAddr.Address, crossChainAddr.Chain.Name))
	assert.NoError(t, linkResult.Error)
	depositAddr := string(linkResult.Data)

	// simulate deposit to master key address
	expectedDepositInfo := randomOutpointInfo(depositAddr)
	for _, n := range nodeData {
		n.Mocks.BTC.GetOutPointInfoFunc = func(bHash *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
			if bHash.IsEqual(expectedDepositInfo.BlockHash) && out.String() == expectedDepositInfo.OutPoint.String() {
				return expectedDepositInfo, nil
			}
			return btcTypes.OutPointInfo{}, fmt.Errorf("outpoint info not found")
		}
	}

	// query for deposit info
	bz, err := nodeData[0].Node.Query(
		[]string{btcTypes.QuerierRoute, btcKeeper.QueryOutInfo, expectedDepositInfo.BlockHash.String()},
		abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(expectedDepositInfo.OutPoint)},
	)
	assert.NoError(t, err)
	var actualDepositInfo btcTypes.OutPointInfo
	testutils.Codec().MustUnmarshalJSON(bz, &actualDepositInfo)
	assert.Equal(t, expectedDepositInfo, actualDepositInfo)

	// verify deposit to master key
	verifyResult1 := <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender(), expectedDepositInfo))
	assert.NoError(t, verifyResult1.Error)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// second snapshot
	snapshotResult2 := <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender()})
	assert.NoError(t, snapshotResult2.Error)

	// create new master key for btc
	threshold2 := int(testutils.RandIntBetween(1, int64(nodeCount)))
	masterKeyID2 := randStrings.Next()
	masterKey2, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// prepare mocks with new master key
	var correctKeygens2 []<-chan bool
	for _, n := range nodeData {
		correctKeygens2 = append(correctKeygens2, prepareKeygen(n.Mocks.Keygen, masterKeyID2, masterKey2.PublicKey))
	}

	// start new keygen
	keygenResult2 := <-chain.Submit(
		tssTypes.MsgKeygenStart{Sender: randomSender(), NewKeyID: masterKeyID2, Threshold: threshold2})
	assert.NoError(t, keygenResult2.Error)
	for _, isCorrect := range correctKeygens2 {
		assert.True(t, <-isCorrect)
	}

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// assign second key to be the new master key
	assignKeyResult2 := <-chain.Submit(
		tssTypes.MsgAssignNextMasterKey{Sender: randomSender(), Chain: btc.Bitcoin.Name, KeyID: masterKeyID2})
	assert.NoError(t, assignKeyResult2.Error)

	// prepare mocks to sign consolidation transaction with first master key
	var correctSigns []<-chan bool
	msgToSign := NewSyncedBytes()
	for _, n := range nodeData {
		correctSign := prepareSign(n.Mocks.Sign, masterKeyID1, masterKey1, msgToSign)
		correctSigns = append(correctSigns, correctSign)
	}

	// sign the consolidation transaction
	fee := btcutil.Amount(testutils.RandIntBetween(1, int64(actualDepositInfo.Amount)))
	signResult := <-chain.Submit(btcTypes.NewMsgSign(randomSender(), fee))
	assert.NoError(t, signResult.Error)
	for _, isCorrect := range correctSigns {
		assert.True(t, <-isCorrect)
	}

	// wait for voting to be done (signing takes longer to tally up)
	chain.WaitNBlocks(20)

	// send tx to Bitcoin
	bz, err = nodeData[0].Node.Query([]string{btcTypes.QuerierRoute, btcKeeper.SendTx}, abci.RequestQuery{})
	assert.NoError(t, err)

	actualTx := nodeData[0].Mocks.BTC.SendRawTransactionCalls()[0].Tx
	consolidationAddr := createBTCAddress(masterKey2, nodeData[0].Mocks.BTC.Network())
	assert.True(t, txCorrectlyFormed(actualTx, actualDepositInfo, fee, consolidationAddr))

	// simulate confirmed tx to master address 2
	var consolidationTxHash *chainhash.Hash
	testutils.Codec().MustUnmarshalJSON(bz, &consolidationTxHash)

	eConsolidationInfo := randomOutpointInfo(consolidationAddr.EncodeAddress())
	eConsolidationInfo.Amount = btcutil.Amount(actualTx.TxOut[0].Value)
	eConsolidationInfo.OutPoint = wire.NewOutPoint(consolidationTxHash, 0)
	for _, n := range nodeData {
		n.Mocks.BTC.GetOutPointInfoFunc = func(blockHash *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
			if blockHash.IsEqual(eConsolidationInfo.BlockHash) && out.String() == eConsolidationInfo.OutPoint.String() {
				return eConsolidationInfo, nil
			}
			return btcTypes.OutPointInfo{}, fmt.Errorf("outpoint info not found")
		}
	}

	// query for consolidation info
	bz, err = nodeData[0].Node.Query(
		[]string{btcTypes.QuerierRoute, btcKeeper.QueryOutInfo, eConsolidationInfo.BlockHash.String()},
		abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(eConsolidationInfo.OutPoint)},
	)
	assert.NoError(t, err)
	var aConsolidationInfo btcTypes.OutPointInfo
	testutils.Codec().MustUnmarshalJSON(bz, &aConsolidationInfo)
	assert.Equal(t, eConsolidationInfo, aConsolidationInfo)

	// verify master key transfer
	verifyResult2 := <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender(), aConsolidationInfo))
	assert.NoError(t, verifyResult2.Error)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// rotate master key to key 2
	rotateResult2 := <-chain.Submit(tssTypes.MsgRotateMasterKey{Sender: randomSender(), Chain: btc.Bitcoin.Name})
	assert.NoError(t, rotateResult2.Error)
}

func txCorrectlyFormed(tx *wire.MsgTx, deposit btcTypes.OutPointInfo, fee btcutil.Amount, addr btcutil.Address) bool {
	script, err := txscript.PayToAddrScript(addr)
	if err != nil {
		panic(err)
	}
	return len(tx.TxOut) == 1 && // one TxOut
		bytes.Equal(tx.TxOut[0].PkScript, script) && // address matches
		btcutil.Amount(tx.TxOut[0].Value) == deposit.Amount-fee && // amount matches
		tx.TxIn[0].PreviousOutPoint.String() == deposit.OutPoint.String() // input matches
}

func createBTCAddress(key *ecdsa.PrivateKey, network btcTypes.Network) *btcutil.AddressWitnessScriptHash {
	script, err := btcTypes.CreateMasterRedeemScript(btcec.PublicKey(key.PublicKey))
	if err != nil {
		panic(err)
	}
	consolidationAddr, err := btcTypes.CreateDepositAddress(network, script)
	if err != nil {
		panic(err)
	}

	return consolidationAddr
}

func randomOutpointInfo(recipient string) btcTypes.OutPointInfo {
	txHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}

	voutIdx := uint32(testutils.RandIntBetween(0, 100))
	return btcTypes.OutPointInfo{
		OutPoint:      wire.NewOutPoint(txHash, voutIdx),
		BlockHash:     blockHash,
		Amount:        btcutil.Amount(testutils.RandIntBetween(1, 10000000)),
		Address:       recipient,
		Confirmations: uint64(testutils.RandIntBetween(1, 10000)),
	}
}
