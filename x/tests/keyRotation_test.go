package tests

import (
	"encoding/hex"
	"math/big"
	rand2 "math/rand"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshotTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	types2 "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// Testing the key rotation functionality.
// (0. Register proxies for all validators)
//  1. Create a key (creates a snapshot automatically
//  2. Wait for vote
//  3. Designate that key to be the first master key for bitcoin
//  4. Rotate to the designated master key
//  5. Simulate bitcoin deposit to the current master key
//  6. Query deposit tx info
//  7. Confirm the deposit is confirmed on bitcoin
//  8. Wait for vote
//  9. Create a new key (with the second snapshot)
// 10. Wait for vote
// 11. Designate that key to be the next master key for bitcoin
// 12. Sign a consolidation transaction
// 13. Wait for vote
// 14. Send the signed transaction to bitcoin
// 15. Query transfer tx info
// 16. Confirm the consolidation transfer is confirmed on bitcoin
// 17. Wait for vote
// 18. Rotate to the new master key
func TestBitcoinKeyRotation(t *testing.T) {
	randStrings := rand.Strings(5, 20)
	cdc := app.MakeEncodingConfig().Amino

	// set up chain
	const nodeCount = 10
	chain, nodeData := initChain(nodeCount, "keyRotation")
	listeners := registerWaitEventListeners(nodeData[0])

	// register proxies for all validators
	for i := 0; i < nodeCount; i++ {
		operatorAddress, err := sdk.ValAddressFromBech32(nodeData[i].Validator.OperatorAddress)
		if err != nil {
			panic(err)
		}
		res := <-chain.Submit(&snapshotTypes.RegisterProxyRequest{PrincipalAddr: operatorAddress, ProxyAddr: nodeData[i].Proxy})
		assert.NoError(t, res.Error)
	}

	chains := []string{btc.Bitcoin.Name, evm.Ethereum.Name}

	for _, c := range chains {
		masterKeyID := randStrings.Next()
		masterKeygenResult := <-chain.Submit(types2.NewStartKeygenRequest(randomSender(), masterKeyID, tss.MasterKey))
		assert.NoError(t, masterKeygenResult.Error)

		// wait for voting to be done
		if err := waitFor(listeners.keygenDone, 1); err != nil {
			assert.FailNow(t, "keygen", err)
		}

		rotateMasterKeyResult := <-chain.Submit(types2.NewRotateKeyRequest(randomSender(), c, tss.MasterKey, masterKeyID))
		assert.NoError(t, rotateMasterKeyResult.Error)

		secondaryKeyID := randStrings.Next()
		secondaryKeygenResult := <-chain.Submit(types2.NewStartKeygenRequest(randomSender(), secondaryKeyID, tss.SecondaryKey))
		assert.NoError(t, secondaryKeygenResult.Error)

		// wait for voting to be done
		if err := waitFor(listeners.keygenDone, 1); err != nil {
			assert.FailNow(t, "keygen", err)
		}

		rotateSecondaryKeyResult := <-chain.Submit(types2.NewRotateKeyRequest(randomSender(), c, tss.SecondaryKey, secondaryKeyID))
		assert.NoError(t, rotateSecondaryKeyResult.Error)
	}

	// setup axelar gateway
	bytecode, err := nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QBytecode, "ethereum", evmKeeper.BCGatewayDeployment},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)

	bz, err := nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QAddressByKeyRole, "ethereum", "master"},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)

	var queryAddressResponse evmTypes.QueryAddressResponse
	evmTypes.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &queryAddressResponse)

	bz, err = nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QAddressByKeyRole, "ethereum", "secondary"},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)
	evmTypes.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &queryAddressResponse)

	nonce := rand2.Uint64()
	gasLimit := rand2.Uint64()
	gasPrice := big.NewInt(rand2.Int63())

	tx := gethTypes.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, bytecode)

	deployGatewayResult := <-chain.Submit(
		&evmTypes.SignTxRequest{Sender: randomSender(), Chain: "ethereum", Tx: cdc.MustMarshalJSON(tx)})
	assert.NoError(t, deployGatewayResult.Error)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(listeners.signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	var signTxResponse evmTypes.SignTxResponse
	assert.NoError(t, proto.Unmarshal(deployGatewayResult.Data, &signTxResponse))
	_, err = nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QSignedTx, "ethereum", signTxResponse.TxID},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)

	// deploy token
	createDeployTokenResult := <-chain.Submit(
		&evmTypes.CreateDeployTokenRequest{Sender: randomSender(), Chain: "ethereum", OriginChain: "bitcoin", Capacity: sdk.NewInt(100000), Decimals: 8, Symbol: "satoshi", TokenName: "Satoshi"})
	assert.NoError(t, createDeployTokenResult.Error)
	signDeployTokenResult := <-chain.Submit(
		&evmTypes.SignCommandsRequest{Sender: randomSender(), Chain: "ethereum"})
	assert.NoError(t, signDeployTokenResult.Error)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(listeners.signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	// confirm the token deployment
	txHash := common.BytesToHash(bz)

	bz, err = nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QTokenAddress, "ethereum", "satoshi"},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)

	bz, err = nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QAxelarGatewayAddress, "ethereum"},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)

	confirmResult1 := <-chain.Submit(evmTypes.NewConfirmTokenRequest(randomSender(), "ethereum", "bitcoin", txHash))
	assert.NoError(t, confirmResult1.Error)

	if err := waitFor(listeners.ethTokenDone, 1); err != nil {
		assert.FailNow(t, "confirmation", err)
	}

	// simulate deposits
	totalDepositCount := int(rand.I64Between(1, 20))
	var totalDepositAmount int64
	deposits := make(map[string]btcTypes.OutPointInfo)
	var outpointsToSign []btcTypes.OutPointToSign

	for i := 0; i < totalDepositCount; i++ {
		// get deposit address for ethereum transfer
		crossChainAddr := nexus.CrossChainAddress{Chain: evm.Ethereum, Address: randStrings.Next()}
		linkResult := <-chain.Submit(btcTypes.NewLinkRequest(randomSender(), crossChainAddr.Address, crossChainAddr.Chain.Name))
		assert.NoError(t, linkResult.Error)

		// simulate deposit to master key address
		var linkResponse btcTypes.LinkResponse
		assert.NoError(t, proto.Unmarshal(linkResult.Data, &linkResponse))
		depositInfo := randomOutpointInfo(linkResponse.DepositAddr)

		// confirm deposit to master key
		confirmResult1 := <-chain.Submit(btcTypes.NewConfirmOutpointRequest(randomSender(), depositInfo))
		assert.NoError(t, confirmResult1.Error)

		// store this information for later in the test
		totalDepositAmount += int64(depositInfo.Amount)
		deposits[depositInfo.OutPoint] = depositInfo

		randomPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}

		outpointsToSign = append(outpointsToSign, btcTypes.OutPointToSign{
			OutPointInfo: depositInfo,
			AddressInfo: btcTypes.NewDepositAddress(
				tss.Key{ID: rand.Str(10), Value: randomPrivateKey.PublicKey, Role: tss.MasterKey},
				tss.Key{ID: rand.Str(10), Value: randomPrivateKey.PublicKey, Role: tss.SecondaryKey},
				btcTypes.DefaultParams().Network,
				crossChainAddr,
			),
		})
	}

	// wait for voting to be done
	if err := waitFor(listeners.btcDone, totalDepositCount); err != nil {
		assert.FailNow(t, "confirmation", err)
	}

	// start new keygen
	secondaryKeyID2 := randStrings.Next()
	keygenResult := <-chain.Submit(types2.NewStartKeygenRequest(randomSender(), secondaryKeyID2, tss.SecondaryKey))
	assert.NoError(t, keygenResult.Error)

	// wait for voting to be done
	if err := waitFor(listeners.keygenDone, 1); err != nil {
		assert.FailNow(t, "keygen", err)
	}

	// create the consolidation transaction
	createResult := <-chain.Submit(btcTypes.NewCreatePendingTransfersTxRequest(randomSender(), secondaryKeyID2, 0))
	assert.NoError(t, createResult.Error)

	// sign the consolidation transaction
	signResult := <-chain.Submit(btcTypes.NewSignTxRequest(randomSender(), tss.SecondaryKey))
	assert.NoError(t, signResult.Error)

	// wait for voting to be done
	if err := waitFor(listeners.signDone, totalDepositCount); err != nil {
		assert.FailNow(t, "signing", err)
	}

	// wait for the end-block trigger to match signatures with the tx
	chain.WaitNBlocks(2 * btcTypes.DefaultParams().SigCheckInterval)

	// get signed tx to Bitcoin
	bz, err = nodeData[0].Node.Query([]string{btcTypes.QuerierRoute, btcKeeper.QLatestTxByKeyRole, tss.SecondaryKey.SimpleString()}, abci.RequestQuery{})
	assert.NoError(t, err)

	var txRes btcTypes.QueryTxResponse
	btcTypes.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &txRes)

	buf, err := hex.DecodeString(txRes.Tx)
	assert.NoError(t, err)
	signedTx := btcTypes.MustDecodeTx(buf)

	fee := btcTypes.EstimateTxSize(signedTx, outpointsToSign)

	satoshi, err := btcTypes.ToSatoshiCoin(btcTypes.DefaultParams().MinOutputAmount)
	if err != nil {
		panic(err)
	}
	assert.True(t, txCorrectlyFormed(&signedTx, deposits, totalDepositAmount-fee-satoshi.Amount.Int64()))

	// expected consolidation info
	consAddr := getAddress(signedTx.TxOut[0], btcTypes.DefaultParams().Network.Params())
	consolidationInfo := randomOutpointInfo(consAddr.EncodeAddress())
	consolidationInfo.Amount = btcutil.Amount(signedTx.TxOut[0].Value)
	hash := signedTx.TxHash()
	consolidationInfo.OutPoint = wire.NewOutPoint(&hash, 0).String()

	bz, err = nodeData[0].Node.Query([]string{btcTypes.QuerierRoute, btcKeeper.QConsolidationAddressByKeyRole, tss.SecondaryKey.SimpleString()}, abci.RequestQuery{})
	assert.NoError(t, err)

	var addressRes btcTypes.QueryAddressResponse
	btcTypes.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &addressRes)

	assert.Equal(t, secondaryKeyID2, addressRes.KeyID)
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
	for _, in := range tx.TxIn {
		if _, ok := deposits[in.PreviousOutPoint.String()]; !ok || in.Witness == nil {
			return false
		}
	}

	satoshi, err := btcTypes.ToSatoshiCoin(btcTypes.DefaultParams().MinOutputAmount)
	if err != nil {
		panic(err)
	}

	return len(tx.TxOut) == 2 && // two TxOut's
		tx.TxOut[1].Value == txAmount && // change TxOut
		tx.TxOut[0].Value == satoshi.Amount.Int64() // anyone-can-spend TxOut
}
