package tests

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	goEth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	goEthTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	ethKeeper "github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// Testing the key rotation functionality.
// (0. Register proxies for all validators)
//  2. Create a key (creates a snapshot automatically
//  2. Wait for vote
//  3. Designate that key to be the first master key for bitcoin
//  4. Rotate to the designated master key
//  5. Simulate bitcoin deposit to the current master key
//  6. Query deposit tx info
//  7. Verify the deposit is confirmed on bitcoin
//  8. Wait for vote
//  9. Create a new key (with the second snapshot)
// 10. Wait for vote
// 11. Designate that key to be the next master key for bitcoin
// 12. Sign a consolidation transaction
// 13. Wait for vote
// 14. Send the signed transaction to bitcoin
// 15. Query transfer tx info
// 16. Verify the consolidation transfer is confirmed on bitcoin
// 17. Wait for vote
// 18. Rotate to the new master key
func TestBitcoinKeyRotation(t *testing.T) {
	randStrings := rand.Strings(5, 50)

	// set up chain
	const nodeCount = 10
	chain, nodeData := initChain(nodeCount, "keyRotation")
	keygenDone, verifyDone, signDone := registerWaitEventListeners(nodeData[0])

	// register proxies for all validators
	for i, proxy := range randStrings.Take(nodeCount) {
		res := <-chain.Submit(broadcastTypes.MsgRegisterProxy{Principal: nodeData[i].Validator.OperatorAddress, Proxy: sdk.AccAddress(proxy)})
		assert.NoError(t, res.Error)
	}

	chains := []string{btc.Bitcoin.Name, eth.Ethereum.Name}

	// start keygen
	masterKeyID1 := randStrings.Next()
	keygenResult1 := <-chain.Submit(tssTypes.MsgKeygenStart{Sender: randomSender(), NewKeyID: masterKeyID1})
	assert.NoError(t, keygenResult1.Error)

	// wait for voting to be done
	if err := waitFor(keygenDone, 1); err != nil {
		assert.FailNow(t, "keygen", err)
	}
	// assign chain master key
	for _, c := range chains {
		assignKeyResult := <-chain.Submit(
			tssTypes.MsgAssignNextMasterKey{Sender: randomSender(), Chain: c, KeyID: masterKeyID1})
		assert.NoError(t, assignKeyResult.Error)

	}

	// rotate chain master key
	for _, c := range chains {
		rotateEthResult := <-chain.Submit(tssTypes.MsgRotateMasterKey{Sender: randomSender(), Chain: c})
		assert.NoError(t, rotateEthResult.Error)
	}

	// setup axelar gateway
	bz, err := nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.CreateDeployTx},
		abci.RequestQuery{
			Data: testutils.Codec().MustMarshalJSON(
				ethTypes.DeployParams{
					GasPrice: sdk.NewInt(1),
					GasLimit: 3000000,
				})},
	)
	assert.NoError(t, err)
	var result ethTypes.DeployResult
	testutils.Codec().MustUnmarshalJSON(bz, &result)

	deployGatewayResult := <-chain.Submit(
		ethTypes.MsgSignTx{Sender: randomSender(), Tx: testutils.Codec().MustMarshalJSON(result.Tx)})
	assert.NoError(t, deployGatewayResult.Error)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	bz, err = nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.SendTx, string(deployGatewayResult.Data)},
		abci.RequestQuery{Data: nil},
	)

	// deploy token
	deployTokenResult := <-chain.Submit(
		ethTypes.MsgSignDeployToken{Sender: randomSender(), Capacity: sdk.NewInt(100000), Decimals: 8, Symbol: "satoshi", TokenName: "Satoshi"})
	assert.NoError(t, deployTokenResult.Error)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	// send token deployment tx to ethereum
	commandID := common.BytesToHash(deployTokenResult.Data)
	nodeData[0].Mocks.ETH.SendAndSignTransactionFunc = func(_ context.Context, _ goEth.CallMsg) (string, error) {
		return "", nil
	}

	sender := randomEthSender()
	bz, err = nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.SendCommand},
		abci.RequestQuery{
			Data: testutils.Codec().MustMarshalJSON(
				ethTypes.CommandParams{
					CommandID: ethTypes.CommandID(commandID),
					Sender:    sender.String(),
				})},
	)
	assert.NoError(t, err)

	// verify the token deployment
	var txHashHex string
	testutils.Codec().MustUnmarshalJSON(bz, &txHashHex)
	txHash := common.HexToHash(txHashHex)

	bz, err = nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.QueryTokenAddress, "satoshi"},
		abci.RequestQuery{Data: nil},
	)
	tokenAddr := common.BytesToAddress(bz)
	bz, err = nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.QueryAxelarGatewayAddress},
		abci.RequestQuery{Data: nil},
	)
	gatewayAddr := common.BytesToAddress(bz)
	logs := createTokenDeployLogs(gatewayAddr, tokenAddr)
	var ethBlock int64
	ethBlock = testutils.RandIntBetween(10, 100)

	for _, node := range nodeData {

		node.Mocks.ETH.BlockNumberFunc = func(ctx context.Context) (uint64, error) {
			return uint64(ethBlock), nil
		}
		node.Mocks.ETH.TransactionReceiptFunc = func(ctx context.Context, hash common.Hash) (*goEthTypes.Receipt, error) {

			if bytes.Equal(txHash.Bytes(), hash.Bytes()) {
				return &goEthTypes.Receipt{TxHash: hash, BlockNumber: big.NewInt(ethBlock - 5), Logs: logs}, nil
			}
			return &goEthTypes.Receipt{}, fmt.Errorf("tx not found")
		}
	}

	verifyResult1 := <-chain.Submit(ethTypes.NewMsgVerifyErc20TokenDeploy(randomSender(), txHash, "satoshi"))
	assert.NoError(t, verifyResult1.Error)

	if err := waitFor(verifyDone, 1); err != nil {
		assert.FailNow(t, "verification", err)
	}

	// simulate deposits
	totalDepositCount := int(rand.I64Between(1, 20))
	var totalDepositAmount int64
	deposits := make(map[string]btcTypes.OutPointInfo)

	for i := 0; i < totalDepositCount; i++ {
		// get deposit address for ethereum transfer
		crossChainAddr := nexus.CrossChainAddress{Chain: eth.Ethereum, Address: randStrings.Next()}
		linkResult := <-chain.Submit(btcTypes.NewMsgLink(randomSender(), crossChainAddr.Address, crossChainAddr.Chain.Name))
		assert.NoError(t, linkResult.Error)

		// simulate deposit to master key address
		depositAddr := string(linkResult.Data)
		expectedDepositInfo := randomOutpointInfo(depositAddr)
		for _, n := range nodeData {
			n.Mocks.BTC.GetOutPointInfoFunc = func(bHash *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
				if !bHash.IsEqual(expectedDepositInfo.BlockHash) || out.String() != expectedDepositInfo.OutPoint.String() {
					return btcTypes.OutPointInfo{}, fmt.Errorf("outpoint info not found")
				}
				return expectedDepositInfo, nil
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

		// store this information for later in the test
		totalDepositAmount += int64(actualDepositInfo.Amount)
		deposits[actualDepositInfo.OutPoint.String()] = actualDepositInfo
	}

	// wait for voting to be done
	if err := waitFor(verifyDone, totalDepositCount); err != nil {
		assert.FailNow(t, "verification", err)
	}

	// start new keygen
	masterKeyID2 := randStrings.Next()
	keygenResult2 := <-chain.Submit(tssTypes.MsgKeygenStart{Sender: randomSender(), NewKeyID: masterKeyID2})
	assert.NoError(t, keygenResult2.Error)

	// wait for voting to be done
	if err := waitFor(keygenDone, 1); err != nil {
		assert.FailNow(t, "keygen", err)
	}

	// assign second key to be the new master key
	assignKeyResult := <-chain.Submit(
		tssTypes.MsgAssignNextMasterKey{Sender: randomSender(), Chain: btc.Bitcoin.Name, KeyID: masterKeyID2})
	assert.NoError(t, assignKeyResult.Error)

	// sign the consolidation transaction
	fee := rand.I64Between(1, totalDepositAmount)
	signResult := <-chain.Submit(btcTypes.NewMsgSignPendingTransfers(randomSender(), btcutil.Amount(fee)))
	assert.NoError(t, signResult.Error)

	// wait for voting to be done
	if err := waitFor(signDone, totalDepositCount); err != nil {
		assert.FailNow(t, "signing", err)
	}

	// send tx to Bitcoin
	bz, err = nodeData[0].Node.Query([]string{btcTypes.QuerierRoute, btcKeeper.SendTx}, abci.RequestQuery{})
	assert.NoError(t, err)

	actualTx := nodeData[0].Mocks.BTC.SendRawTransactionCalls()[0].Tx
	assert.True(t, txCorrectlyFormed(actualTx, deposits, totalDepositAmount-fee))

	// expected consolidation info
	consAddr := getAddress(actualTx.TxOut[0], nodeData[0].Mocks.BTC.Network().Params)
	eConsolidationInfo := randomOutpointInfo(consAddr.EncodeAddress())
	eConsolidationInfo.Amount = btcutil.Amount(actualTx.TxOut[0].Value)
	var consolidationTxHash *chainhash.Hash
	testutils.Codec().MustUnmarshalJSON(bz, &consolidationTxHash)
	eConsolidationInfo.OutPoint = wire.NewOutPoint(consolidationTxHash, 0)

	// simulate confirmed tx to master address 2
	for _, n := range nodeData {
		n.Mocks.BTC.GetOutPointInfoFunc = func(blockHash *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
			if !blockHash.IsEqual(eConsolidationInfo.BlockHash) || out.String() != eConsolidationInfo.OutPoint.String() {
				return btcTypes.OutPointInfo{}, fmt.Errorf("outpoint info not found")
			}
			return eConsolidationInfo, nil
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
	if err := waitFor(verifyDone, 1); err != nil {
		assert.FailNow(t, "verification", err)
	}

	// rotate master key to new key
	rotateResult := <-chain.Submit(tssTypes.MsgRotateMasterKey{Sender: randomSender(), Chain: btc.Bitcoin.Name})
	assert.NoError(t, rotateResult.Error)
}

func getAddress(txOut *wire.TxOut, chainParams *chaincfg.Params) btcutil.Address {
	script, err := txscript.ParsePkScript(txOut.PkScript)
	if err != nil {
		panic(err)
	}
	consAddr, err := script.Address(chainParams)
	if err != nil {
		panic(err)
	}
	return consAddr
}

func txCorrectlyFormed(tx *wire.MsgTx, deposits map[string]btcTypes.OutPointInfo, txAmount int64) bool {
	txInsCorrect := true
	for _, in := range tx.TxIn {
		if _, ok := deposits[in.PreviousOutPoint.String()]; !ok {
			txInsCorrect = false
			break
		}
	}

	return len(tx.TxOut) == 1 && // one TxOut
		tx.TxOut[0].Value == txAmount && // amount matches
		txInsCorrect // inputs match
}
