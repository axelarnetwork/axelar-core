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

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	ethKeeper "github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
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
		res := <-chain.Submit(&broadcastTypes.MsgRegisterProxy{PrincipalAddr: operatorAddress, ProxyAddr: rand.Bytes(sdk.AddrLen)})
		assert.NoError(t, res.Error)
	}

	// start keygen
	btcMasterKeyID := randStrings.Next()
	btcKeygenResult := <-chain.Submit(tssTypes.NewMsgKeygenStart(randomSender(), btcMasterKeyID, 0, tss.WeightedByStake))
	assert.NoError(t, btcKeygenResult.Error)

	// start keygen
	ethMasterKeyID := randStrings.Next()
	ethKeygenResult := <-chain.Submit(tssTypes.NewMsgKeygenStart(randomSender(), ethMasterKeyID, 0, tss.WeightedByStake))
	assert.NoError(t, ethKeygenResult.Error)

	// wait for voting to be done
	if err := waitFor(listeners.keygenDone, 2); err != nil {
		assert.FailNow(t, "keygen", err)
	}

	chains := []string{btc.Bitcoin.Name, eth.Ethereum.Name}
	for _, c := range chains {
		masterKeyID := randStrings.Next()
		masterKeygenResult := <-chain.Submit(tssTypes.NewMsgKeygenStart(randomSender(), masterKeyID, 0, tss.WeightedByStake))
		assert.NoError(t, masterKeygenResult.Error)

		// wait for voting to be done
		if err := waitFor(listeners.keygenDone, 1); err != nil {
			assert.FailNow(t, "keygen", err)
		}

		assignMasterKeyResult := <-chain.Submit(tssTypes.NewMsgAssignNextKey(randomSender(), c, masterKeyID, tss.MasterKey))
		assert.NoError(t, assignMasterKeyResult.Error)

		rotateMasterKeyResult := <-chain.Submit(tssTypes.NewMsgRotateKey(randomSender(), c, tss.MasterKey))
		assert.NoError(t, rotateMasterKeyResult.Error)

		if c == btc.Bitcoin.Name {
			secondaryKeyID := randStrings.Next()
			secondaryKeygenResult := <-chain.Submit(tssTypes.NewMsgKeygenStart(randomSender(), secondaryKeyID, 0, tss.OnePerValidator))
			assert.NoError(t, secondaryKeygenResult.Error)

			// wait for voting to be done
			if err := waitFor(listeners.keygenDone, 1); err != nil {
				assert.FailNow(t, "keygen", err)
			}

			assignSecondaryKeyResult := <-chain.Submit(tssTypes.NewMsgAssignNextKey(randomSender(), c, secondaryKeyID, tss.SecondaryKey))
			assert.NoError(t, assignSecondaryKeyResult.Error)

			rotateSecondaryKeyResult := <-chain.Submit(tssTypes.NewMsgRotateKey(randomSender(), c, tss.SecondaryKey))
			assert.NoError(t, rotateSecondaryKeyResult.Error)
		}
	}

	// setup axelar gateway
	bz, err := nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.CreateDeployTx},
		abci.RequestQuery{
			Data: cdc.MustMarshalJSON(
				ethTypes.DeployParams{
					GasPrice: sdk.NewInt(1),
					GasLimit: 3000000,
				})},
	)
	assert.NoError(t, err)
	var result ethTypes.DeployResult
	cdc.MustUnmarshalJSON(bz, &result)

	deployGatewayResult := <-chain.Submit(
		&ethTypes.MsgSignTx{Sender: randomSender(), Tx: cdc.MustMarshalJSON(result.Tx)})
	assert.NoError(t, deployGatewayResult.Error)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(listeners.signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	_, err = nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.SendTx, string(deployGatewayResult.Data)},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)

	// deploy token
	deployTokenResult := <-chain.Submit(
		&ethTypes.MsgSignDeployToken{Sender: randomSender(), Capacity: sdk.NewInt(100000), Decimals: 8, Symbol: "satoshi", TokenName: "Satoshi"})
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
		[]string{ethTypes.QuerierRoute, ethKeeper.SendCommand},
		abci.RequestQuery{
			Data: cdc.MustMarshalJSON(
				ethTypes.CommandParams{
					CommandID: ethTypes.CommandID(commandID1),
					Sender:    sender1.String(),
				})},
	)
	assert.NoError(t, err)

	// confirm the token deployment
	var txHashHex string
	cdc.MustUnmarshalJSON(bz, &txHashHex)
	txHash := common.HexToHash(txHashHex)

	bz, err = nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.QueryTokenAddress, "satoshi"},
		abci.RequestQuery{Data: nil},
	)
	assert.NoError(t, err)
	tokenAddr := common.BytesToAddress(bz)
	bz, err = nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.QueryAxelarGatewayAddress},
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

	confirmResult := <-chain.Submit(ethTypes.NewMsgConfirmERC20TokenDeploy(randomSender(), txHash, "satoshi"))
	assert.NoError(t, confirmResult.Error)

	if err := waitFor(listeners.ethTokenDone, 1); err != nil {
		assert.FailNow(t, "confirmation", err)
	}

	// steps followed as per https://github.com/axelarnetwork/axelarate#mint-erc20-wrapped-bitcoin-tokens-on-ethereum
	totalDepositCount := int(rand.I64Between(1, 4))
	for i := 0; i < totalDepositCount; i++ {
		// Get a deposit address for an Ethereum recipient address
		// we don't provide an actual recipient address, so it is created automatically
		crosschainAddr := nexus.CrossChainAddress{Chain: eth.Ethereum, Address: rand.StrBetween(5, 20)}
		res := <-chain.Submit(btcTypes.NewMsgLink(randomSender(), crosschainAddr.Address, crosschainAddr.Chain.Name))
		assert.NoError(t, res.Error)
		depositAddr := string(res.Data)

		// Simulate deposit
		depositInfo := randomOutpointInfo(depositAddr)

		// confirm the previously received information
		res = <-chain.Submit(btcTypes.NewMsgConfirmOutpoint(randomSender(), depositInfo))
		assert.NoError(t, res.Error)
	}

	// Wait until confirm is complete
	if err := waitFor(listeners.btcDone, totalDepositCount); err != nil {
		assert.FailNow(t, "confirmation", err)
	}

	// Sign all pending transfers to Ethereum
	res := <-chain.Submit(ethTypes.NewMsgSignPendingTransfers(randomSender()))
	assert.NoError(t, res.Error)

	commandID2 := common.BytesToHash(res.Data)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(listeners.signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	// Submit the minting command from an externally controlled address to AxelarGateway
	sender2 := randomEthSender()

	_, err = nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.SendCommand},
		abci.RequestQuery{Data: cdc.MustMarshalJSON(
			ethTypes.CommandParams{CommandID: ethTypes.CommandID(commandID2), Sender: sender2.String()})},
	)
	assert.NoError(t, err)
}
