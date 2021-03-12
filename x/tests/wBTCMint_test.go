package tests

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
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

// 0. Create and start a chain
// 1. Get a deposit address for the given Ethereum recipient address
// 2. Send BTC to the deposit address and wait until confirmed
// 3. Collect all information that needs to be verified about the deposit
// 4. Verify the previously received information
// 5. Wait until verification is complete
// 6. Sign all pending transfers to Ethereum
// 7. Submit the minting command from an externally controlled address to AxelarGateway

func Test_wBTC_mint(t *testing.T) {
	randStrings := rand.Strings(5, 50)

	// 0. Set up chain
	const nodeCount = 10

	// create a chain with nodes and assign them as validators
	chain, nodeData := initChain(nodeCount, "mint")
	keygenDone, verifyDone, signDone := registerWaitEventListeners(nodeData[0])

	// register proxies for all validators
	for i, proxy := range randStrings.Take(nodeCount) {
		res := <-chain.Submit(broadcastTypes.MsgRegisterProxy{Principal: nodeData[i].Validator.OperatorAddress, Proxy: sdk.AccAddress(proxy)})
		assert.NoError(t, res.Error)
	}

	// start keygen
	btcMasterKeyID := randStrings.Next()
	btcKeygenResult := <-chain.Submit(tssTypes.MsgKeygenStart{Sender: randomSender(), NewKeyID: btcMasterKeyID})
	assert.NoError(t, btcKeygenResult.Error)

	// start keygen
	ethMasterKeyID := randStrings.Next()
	ethKeygenResult := <-chain.Submit(tssTypes.MsgKeygenStart{Sender: randomSender(), NewKeyID: ethMasterKeyID})
	assert.NoError(t, ethKeygenResult.Error)

	// wait for voting to be done
	if err := waitFor(keygenDone, 2); err != nil {
		assert.FailNow(t, "keygen", err)
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
	commandID1 := common.BytesToHash(deployTokenResult.Data)
	nodeData[0].Mocks.ETH.SendAndSignTransactionFunc = func(_ context.Context, _ goEth.CallMsg) (string, error) {
		return "", nil
	}

	sender1 := randomEthSender()
	bz, err = nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.SendCommand},
		abci.RequestQuery{
			Data: testutils.Codec().MustMarshalJSON(
				ethTypes.CommandParams{
					CommandID: ethTypes.CommandID(commandID1),
					Sender:    sender1.String(),
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
	ethBlock = rand.I64Between(10, 100)

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

	verifyResult := <-chain.Submit(ethTypes.NewMsgVerifyErc20TokenDeploy(randomSender(), txHash, "satoshi"))
	assert.NoError(t, verifyResult.Error)

	if err := waitFor(verifyDone, 1); err != nil {
		assert.FailNow(t, "verification", err)
	}

	// steps followed as per https://github.com/axelarnetwork/axelarate#mint-erc20-wrapped-bitcoin-tokens-on-ethereum
	totalDepositCount := int(rand.I64Between(1, 20))
	for i := 0; i < totalDepositCount; i++ {
		// 1. Get a deposit address for an Ethereum recipient address
		// we don't provide an actual recipient address, so it is created automatically
		crosschainAddr := nexus.CrossChainAddress{Chain: eth.Ethereum, Address: rand.StrBetween(5, 20)}
		res := <-chain.Submit(btcTypes.NewMsgLink(randomSender(), crosschainAddr.Address, crosschainAddr.Chain.Name))
		assert.NoError(t, res.Error)
		depositAddr := string(res.Data)

		// Prepare btc mocks for verification
		expectedDepositInfo := randomOutpointInfo(depositAddr)
		for _, n := range nodeData {
			n.Mocks.BTC.GetOutPointInfoFunc = func(bHash *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
				if !bHash.IsEqual(expectedDepositInfo.BlockHash) || out.String() != expectedDepositInfo.OutPoint.String() {
					return btcTypes.OutPointInfo{}, fmt.Errorf("outpoint info not found")
				}
				return expectedDepositInfo, nil
			}
		}

		// 3. Collect all information that needs to be verified about the deposit
		bz, err := nodeData[0].Node.Query(
			[]string{btcTypes.QuerierRoute, btcKeeper.QueryOutInfo, expectedDepositInfo.BlockHash.String()},
			abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(expectedDepositInfo.OutPoint)})
		assert.NoError(t, err)
		var info btcTypes.OutPointInfo
		testutils.Codec().MustUnmarshalJSON(bz, &info)

		// 4. Verify the previously received information
		res = <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender(), info))
		assert.NoError(t, res.Error)

	}

	// 5. Wait until verification is complete
	if err := waitFor(verifyDone, totalDepositCount); err != nil {
		assert.FailNow(t, "verification", err)
	}

	// 6. Sign all pending transfers to Ethereum
	res := <-chain.Submit(ethTypes.NewMsgSignPendingTransfers(randomSender()))
	assert.NoError(t, res.Error)

	commandID2 := common.BytesToHash(res.Data)

	// wait for voting to be done (signing takes longer to tally up)
	if err := waitFor(signDone, 1); err != nil {
		assert.FailNow(t, "signing", err)
	}

	// 7. Submit the minting command from an externally controlled address to AxelarGateway
	sender2 := randomEthSender()

	_, err = nodeData[0].Node.Query(
		[]string{ethTypes.QuerierRoute, ethKeeper.SendCommand},
		abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(
			ethTypes.CommandParams{CommandID: ethTypes.CommandID(commandID2), Sender: sender2.String()})},
	)
	assert.NoError(t, err)
}
