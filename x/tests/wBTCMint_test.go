package tests

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	goEth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	goEthTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gogo/protobuf/proto"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
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
	cdc := testutils.MakeEncodingConfig().Amino

	// 0. Set up chain
	const nodeCount = 10

	// create a chain with nodes and assign them as validators
	chain, nodeData := initChain(nodeCount, "mint")
	listeners := registerWaitEventListeners(nodeData[0])

	// register proxies for all validators
	for i := 0; i < nodeCount; i++ {
		operatorAddress, err := sdk.ValAddressFromBech32(nodeData[i].Validator.OperatorAddress)
		if err != nil {
			panic(err)
		}
		res := <-chain.Submit(&broadcastTypes.RegisterProxyRequest{PrincipalAddr: operatorAddress, ProxyAddr: nodeData[i].Proxy})
		assert.NoError(t, res.Error)
	}

	// start keygen
	btcMasterKeyID := randStrings.Next()
	btcKeygenResult := <-chain.Submit(types.NewStartKeygenRequest(randomSender(), btcMasterKeyID, 0, tss.WeightedByStake))
	assert.NoError(t, btcKeygenResult.Error)

	// start keygen
	ethMasterKeyID := randStrings.Next()
	ethKeygenResult := <-chain.Submit(types.NewStartKeygenRequest(randomSender(), ethMasterKeyID, 0, tss.WeightedByStake))
	assert.NoError(t, ethKeygenResult.Error)

	// wait for voting to be done
	if err := waitFor(listeners.keygenDone, 2); err != nil {
		assert.FailNow(t, "keygen", err)
	}

	chains := []string{btc.Bitcoin.Name, evm.Ethereum.Name}
	for _, c := range chains {
		masterKeyID := randStrings.Next()
		masterKeygenResult := <-chain.Submit(types.NewStartKeygenRequest(randomSender(), masterKeyID, 0, tss.WeightedByStake))
		assert.NoError(t, masterKeygenResult.Error)

		// wait for voting to be done
		if err := waitFor(listeners.keygenDone, 1); err != nil {
			assert.FailNow(t, "keygen", err)
		}

		assignMasterKeyResult := <-chain.Submit(types.NewAssignKeyRequest(randomSender(), c, masterKeyID, tss.MasterKey))
		assert.NoError(t, assignMasterKeyResult.Error)

		rotateMasterKeyResult := <-chain.Submit(types.NewRotateKeyRequest(randomSender(), c, tss.MasterKey))
		assert.NoError(t, rotateMasterKeyResult.Error)

		if c == btc.Bitcoin.Name {
			secondaryKeyID := randStrings.Next()
			secondaryKeygenResult := <-chain.Submit(types.NewStartKeygenRequest(randomSender(), secondaryKeyID, 0, tss.OnePerValidator))
			assert.NoError(t, secondaryKeygenResult.Error)

			// wait for voting to be done
			if err := waitFor(listeners.keygenDone, 1); err != nil {
				assert.FailNow(t, "keygen", err)
			}

			assignSecondaryKeyResult := <-chain.Submit(types.NewAssignKeyRequest(randomSender(), c, secondaryKeyID, tss.SecondaryKey))
			assert.NoError(t, assignSecondaryKeyResult.Error)

			rotateSecondaryKeyResult := <-chain.Submit(types.NewRotateKeyRequest(randomSender(), c, tss.SecondaryKey))
			assert.NoError(t, rotateSecondaryKeyResult.Error)
		}
	}

	// setup axelar gateway
	bz, err := nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.CreateDeployTx},
		abci.RequestQuery{
			Data: cdc.MustMarshalJSON(
				evmTypes.DeployParams{
					GasPrice: sdk.NewInt(1),
					GasLimit: 3000000,
				})},
	)
	assert.NoError(t, err)
	var result evmTypes.DeployResult
	cdc.MustUnmarshalJSON(bz, &result)

	deployGatewayResult := <-chain.Submit(
		&evmTypes.SignTxRequest{Sender: randomSender(), Tx: cdc.MustMarshalJSON(result.Tx)})
	assert.NoError(t, deployGatewayResult.Error)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(listeners.signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	var signTxResponse evmTypes.SignTxResponse
	assert.NoError(t, proto.Unmarshal(deployGatewayResult.Data, &signTxResponse))
	_, err = nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.SendTx, signTxResponse.TxID},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)

	// deploy token
	deployTokenResult := <-chain.Submit(
		&evmTypes.SignDeployTokenRequest{Sender: randomSender(), Capacity: sdk.NewInt(100000), Decimals: 8, Symbol: "satoshi", TokenName: "Satoshi"})
	assert.NoError(t, deployTokenResult.Error)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(listeners.signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	// send token deployment tx to ethereum
	commandID1 := common.BytesToHash(deployTokenResult.Data)
	nodeData[0].Mocks.ETH.SendAndSignTransactionFunc = func(_ context.Context, _ goEth.CallMsg) (string, error) {
		return "", nil
	}

	sender1 := randomEthSender()
	bz, err = nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.SendCommand},
		abci.RequestQuery{
			Data: cdc.MustMarshalJSON(
				evmTypes.CommandParams{
					CommandID: evmTypes.CommandID(commandID1),
					Sender:    sender1.String(),
				})},
	)
	assert.NoError(t, err)

	// confirm the token deployment
	txHash := common.BytesToHash(bz)

	bz, err = nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QueryTokenAddress, "satoshi"},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)
	tokenAddr := common.BytesToAddress(bz)
	bz, err = nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.QueryAxelarGatewayAddress},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)
	gatewayAddr := common.BytesToAddress(bz)
	logs := createTokenDeployLogs(gatewayAddr, tokenAddr)
	ethBlock := rand.I64Between(10, 100)

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

	confirmResult := <-chain.Submit(evmTypes.NewConfirmTokenRequest(randomSender(), txHash, "satoshi"))
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
	res := <-chain.Submit(evmTypes.NewSignPendingTransfersRequest(randomSender()))
	assert.NoError(t, res.Error)

	commandID2 := common.BytesToHash(res.Data)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(listeners.signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	// Submit the minting command from an externally controlled address to AxelarGateway
	sender2 := randomEthSender()

	_, err = nodeData[0].Node.Query(
		[]string{evmTypes.QuerierRoute, evmKeeper.SendCommand},
		abci.RequestQuery{Data: cdc.MustMarshalJSON(
			evmTypes.CommandParams{CommandID: evmTypes.CommandID(commandID2), Sender: sender2.String()})},
	)
	assert.NoError(t, err)
}
