package tests

import (
	"math/big"
	rand2 "math/rand"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	snapshotTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// 0. Create and start a chain
// 1. Get a deposit address for the given Ethereum recipient address
// 2. Send BTC to the deposit address and wait until confirmed
// 3. Collect all information that needs to be confirmed about the deposit
// 4. Confirm the previously received information
// 5. Wait until confirmation is complete
// 6. Sign all pending transfers to Ethereum
// 7. Submit the minting command from an externally controlled address to AxelarGateway

func Test_wBTC_mint(t *testing.T) {
	randStrings := rand.Strings(5, 50)
	cdc := app.MakeEncodingConfig().Amino

	// 0. Set up chain
	const nodeCount = 10

	// create a chain with nodes and assign them as validators
	chain, nodeData := initChain(nodeCount, "mint")
	listeners := registerWaitEventListeners(nodeData[0])
	chains := []string{btc.Bitcoin.Name, evm.Ethereum.Name}

	// register proxies for all validators
	for i := 0; i < nodeCount; i++ {
		operatorAddress, err := sdk.ValAddressFromBech32(nodeData[i].Validator.OperatorAddress)
		if err != nil {
			panic(err)
		}
		res := <-chain.Submit(&snapshotTypes.ProxyReadyRequest{Sender: nodeData[i].Proxy, OperatorAddr: operatorAddress})
		assert.NoError(t, res.Error)

		res = <-chain.Submit(&snapshotTypes.RegisterProxyRequest{Sender: operatorAddress, ProxyAddr: nodeData[i].Proxy})
		assert.NoError(t, res.Error)

		res = <-chain.Submit(&nexusTypes.RegisterChainMaintainerRequest{Sender: nodeData[i].Proxy, Chains: chains})
		assert.NoError(t, res.Error)
	}

	if err := waitFor(listeners.chainActivated, 1); err != nil {
		assert.FailNow(t, "chain activation", err)
	}

	// wait for ack event
	if err := waitFor(listeners.ackRequested, 1); err != nil {
		assert.FailNow(t, "ack", err)
	}

	// start keygen
	btcMasterKeyID := randStrings.Next()
	btcKeygenResult := <-chain.Submit(types.NewStartKeygenRequest(randomSender(), btcMasterKeyID, tss.MasterKey, tss.Threshold))
	assert.NoError(t, btcKeygenResult.Error)

	// start keygen
	ethMasterKeyID := randStrings.Next()
	ethKeygenResult := <-chain.Submit(types.NewStartKeygenRequest(randomSender(), ethMasterKeyID, tss.MasterKey, tss.Threshold))
	assert.NoError(t, ethKeygenResult.Error)

	// wait for voting to be done
	if err := waitFor(listeners.keygenDone, 2); err != nil {
		assert.FailNow(t, "keygen", err)
	}

	for _, c := range chains {
		masterKeyID := randStrings.Next()
		masterKeygenResult := <-chain.Submit(types.NewStartKeygenRequest(randomSender(), masterKeyID, tss.MasterKey, tss.Threshold))
		assert.NoError(t, masterKeygenResult.Error)

		// wait for voting to be done
		if err := waitFor(listeners.keygenDone, 1); err != nil {
			assert.FailNow(t, "keygen", err)
		}

		rotateMasterKeyResult := <-chain.Submit(types.NewRotateKeyRequest(randomSender(), c, tss.MasterKey, masterKeyID))
		assert.NoError(t, rotateMasterKeyResult.Error)

		secondaryKeyID := randStrings.Next()
		secondaryKeygenResult := <-chain.Submit(types.NewStartKeygenRequest(randomSender(), secondaryKeyID, tss.SecondaryKey, tss.Threshold))
		assert.NoError(t, secondaryKeygenResult.Error)

		// wait for voting to be done
		if err := waitFor(listeners.keygenDone, 1); err != nil {
			assert.FailNow(t, "keygen", err)
		}

		rotateSecondaryKeyResult := <-chain.Submit(types.NewRotateKeyRequest(randomSender(), c, tss.SecondaryKey, secondaryKeyID))
		assert.NoError(t, rotateSecondaryKeyResult.Error)

		// wait for ack event
		if err := waitFor(listeners.ackRequested, 1); err != nil {
			assert.FailNow(t, "ack", err)
		}

		var externalKeys []tssTypes.RegisterExternalKeysRequest_ExternalKey
		for i := 0; i < int(tssTypes.DefaultParams().ExternalMultisigThreshold.Denominator); i++ {
			privKey, err := btcec.NewPrivateKey(btcec.S256())
			if err != nil {
				panic(err)
			}
			externalKeys = append(externalKeys, tssTypes.RegisterExternalKeysRequest_ExternalKey{
				ID:     testutils.RandKeyID(),
				PubKey: privKey.PubKey().SerializeCompressed(),
			})
		}

		registerExternalKeysResult := <-chain.Submit(tssTypes.NewRegisterExternalKeysRequest(randomSender(), c, externalKeys...))
		assert.NoError(t, registerExternalKeysResult.Error)
	}

	// setup axelar gateway
	bytecode, err := nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QBytecode, "ethereum", evmKeeper.BCGatewayDeployment},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)

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
	asset := evmTypes.NewAsset("bitcoin", "satoshi")
	tokenDetails := evmTypes.NewTokenDetails("Satoshi", "satoshi", 8, sdk.NewInt(100000))
	createDeployTokenResult := <-chain.Submit(
		&evmTypes.CreateDeployTokenRequest{Sender: randomSender(), Chain: "ethereum", Asset: asset, TokenDetails: tokenDetails})
	assert.NoError(t, createDeployTokenResult.Error)
	signDeployTokenResult := <-chain.Submit(
		&evmTypes.SignCommandsRequest{Sender: randomSender(), Chain: "ethereum"})
	assert.NoError(t, signDeployTokenResult.Error)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(listeners.signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	// confirm the token deployment
	bz, err := nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QTokenAddress, "ethereum", "satoshi"},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)
	txHash := common.BytesToHash(bz)

	_, err = nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QAxelarGatewayAddress, "ethereum"},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)

	confirmResult := <-chain.Submit(evmTypes.NewConfirmTokenRequest(randomSender(), "ethereum", asset, txHash))
	assert.NoError(t, confirmResult.Error)

	if err := waitFor(listeners.ethTokenDone, 1); err != nil {
		assert.FailNow(t, "confirmation", err)
	}

	// steps followed as per https://github.com/axelarnetwork/axelarate#mint-erc20-wrapped-bitcoin-tokens-on-ethereum
	totalDepositCount := int(rand.I64Between(1, 20))
	for i := 0; i < totalDepositCount; i++ {
		// Get a deposit address for an Ethereum recipient address
		// we don't provide an actual recipient address, so it is created automatically
		crosschainAddr := nexus.CrossChainAddress{Chain: evm.Ethereum, Address: rand.StrBetween(5, 20)}
		res := <-chain.Submit(btcTypes.NewLinkRequest(randomSender(), crosschainAddr.Address, crosschainAddr.Chain.Name))
		assert.NoError(t, res.Error)
		var linkResponse btcTypes.LinkResponse
		assert.NoError(t, proto.Unmarshal(res.Data, &linkResponse))

		// Simulate deposit
		depositInfo := randomOutpointInfo(linkResponse.DepositAddr)

		// confirm the previously received information
		res = <-chain.Submit(btcTypes.NewConfirmOutpointRequest(randomSender(), depositInfo))
		assert.NoError(t, res.Error)

	}

	// Wait until confirm is complete
	if err := waitFor(listeners.btcDone, totalDepositCount); err != nil {
		assert.FailNow(t, "confirmation", err)
	}

	// Sign all pending transfers to Ethereum
	createPendingTransfersResult := <-chain.Submit(evmTypes.NewCreatePendingTransfersRequest(randomSender(), "ethereum"))
	assert.NoError(t, createPendingTransfersResult.Error)
}
