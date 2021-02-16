package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	ethKeeper "github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	goEth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
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

	// 0. Set up chain
	const nodeCount = 10
	stringGen := testutils.RandStrings(5, 50).Distinct()
	defer stringGen.Stop()

	// create a chain with nodes and assign them as validators
	chain, validators, mocks, nodes := createChain(nodeCount, &stringGen)
	registerProxies(chain, validators, nodeCount, &stringGen, t)
	takeSnapshot(chain, validators, nodeCount, t)

	// create master keys for btc and eth
	btcMasterKeyID, _ := createMasterKeyID(chain, validators, nodeCount, &stringGen, mocks, t)
	ethMasterKeyID, ethMasterKey := createMasterKeyID(chain, validators, nodeCount, &stringGen, mocks, t)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// assign bitcoin master key
	assignMasterKey(chain, validators, nodeCount, btcMasterKeyID, balance.Bitcoin, t)
	// rotate to the first btc master key
	rotateMasterKey(chain, validators, nodeCount, balance.Bitcoin, t)

	// assign key as ethereum master key
	assignMasterKey(chain, validators, nodeCount, ethMasterKeyID, balance.Ethereum, t)
	// rotate to the first eth master key
	rotateMasterKey(chain, validators, nodeCount, balance.Ethereum, t)

	// steps followed as per https://github.com/axelarnetwork/axelarate#mint-erc20-wrapped-bitcoin-tokens-on-ethereum

	// 1. Get a deposit address for an Ethereum recipient address
	// we don't provide an actual recipient address, so it is created automatically
	depositAddr, _ := getCrossChainAddress(balance.CrossChainAddress{}, balance.Ethereum, chain, validators, nodeCount, t)

	// 2. Send BTC to the deposit address and wait until confirmed
	blockHash, expectedOut, _ := sendBTCtoDepositAddress(depositAddr, mocks)

	// 3. Collect all information that needs to be verified about the deposit
	info := queryOutPointInfo(nodes, blockHash, expectedOut, t)

	// 4. Verify the previously received information
	verifyTx(chain, validators, nodeCount, info, t)

	// 5. Wait until verification is complete
	chain.WaitNBlocks(12)

	// 6. Sign all pending transfers to Ethereum
	commandID := signPendingTransfersTx(chain, validators, nodeCount, mocks, ethMasterKeyID, ethMasterKey, t)

	// 7. Submit the minting command from an externally controlled address to AxelarGateway
	submitAndSignTX(validators, nodes, nodeCount, mocks, commandID, t)
}

// getCrossChainAddress returns the deposit address for an existing chain recipient address
// if no recipient address is provided, then it is generated
func getCrossChainAddress(
	crosschainAddr balance.CrossChainAddress,
	balanceChain balance.Chain,
	chain *fake.BlockChain,
	validators []staking.Validator,
	nodeCount int64,
	t *testing.T) (string, balance.CrossChainAddress) {

	// if no crosschain address is provided, create one
	if crosschainAddr == (balance.CrossChainAddress{}) {
		crosschainAddr = balance.CrossChainAddress{Chain: balanceChain, Address: testutils.RandStringBetween(5, 20)}
	}
	res := <-chain.Submit(btcTypes.NewMsgLink(randomSender(validators, nodeCount), crosschainAddr))
	assert.NoError(t, res.Error)
	depositAddr := string(res.Data)
	return depositAddr, crosschainAddr
}

// sendBTCtoDepositAddress sends a predefined amount to the deposit address
func sendBTCtoDepositAddress(
	depositAddr string,
	mocks testMocks) (*chainhash.Hash, *wire.OutPoint, btcTypes.OutPointInfo) {
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
	outPointInfo := btcTypes.OutPointInfo{
		OutPoint:      wire.NewOutPoint(txHash, voutIdx),
		BlockHash:     blockHash,
		Amount:        btcutil.Amount(testutils.RandIntBetween(1, 10000000)),
		Address:       depositAddr,
		Confirmations: uint64(testutils.RandIntBetween(1, 10000)),
	}

	mocks.BTC.GetOutPointInfoFunc = func(bHash *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
		if bHash.String() == blockHash.String() && out.String() == expectedOut.String() {
			return outPointInfo, nil
		}
		return btcTypes.OutPointInfo{}, fmt.Errorf("tx %s not found", out.String())
	}

	return blockHash, expectedOut, outPointInfo
}

// queryOutPointInfo collects all information that needs to be verified about the deposit
func queryOutPointInfo(nodes []fake.Node, blockHash *chainhash.Hash, expectedOut *wire.OutPoint, t *testing.T) btcTypes.OutPointInfo {
	bz, err := nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.QueryOutInfo, blockHash.String()}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(expectedOut)})
	assert.NoError(t, err)
	var info btcTypes.OutPointInfo
	testutils.Codec().MustUnmarshalJSON(bz, &info)
	return info
}

// verifyTX verifies the deposit information
func verifyTx(chain *fake.BlockChain, validators []staking.Validator, nodeCount int64, info btcTypes.OutPointInfo, t *testing.T) {
	res := <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender(validators, nodeCount), info))
	assert.NoError(t, res.Error)
}

// signPendingTransfersTX signs all pending transfers to Ethereum
func signPendingTransfersTx(chain *fake.BlockChain,
	validators []staking.Validator,
	nodeCount int,
	mocks testMocks,
	ethMasterKeyID string,
	ethMasterKey *ecdsa.PrivateKey,
	t *testing.T) common.Hash {

	msgToSign := make(chan []byte, nodeCount)
	mocks.Sign.SendFunc = func(messageIn *tssd.MessageIn) error {
		assert.Equal(t, ethMasterKeyID, messageIn.GetSignInit().KeyUid)
		msgToSign <- messageIn.GetSignInit().MessageToSign
		return nil
	}
	sigChan := make(chan []byte, 1)
	go func() {
		// Q: No error is produced even if the btcMasterKey is used here.
		// Is there any way to assert that the correct master key was provided?
		r, s, err := ecdsa.Sign(rand.Reader, ethMasterKey, <-msgToSign)
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

	closeTimeout, closeCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer closeCancel()
	mocks.Sign.CloseSendFunc = func() error {
		return nil
	}
	if len(mocks.Sign.CloseSendCalls()) == nodeCount {
		closeCancel()
	}

	res := <-chain.Submit(ethTypes.NewMsgSignPendingTransfersTx(randomSender(validators, int64(nodeCount))))
	assert.NoError(t, res.Error)
	commandID := common.BytesToHash(res.Data)
	<-closeTimeout.Done()
	assert.Equal(t, nodeCount, len(mocks.Sign.CloseSendCalls()))

	// wait for voting to be done
	// Q: Why do we have to wait for 22 blocks instead of 12?
	chain.WaitNBlocks(22)

	return commandID
}

// submitAndSignTX submits the minting command from an externally controlled address to AxelarGateway
func submitAndSignTX(validators []staking.Validator,
	nodes []fake.Node,
	nodeCount int64,
	mocks testMocks,
	commandID common.Hash,
	t *testing.T) {
	// Q: Does SendAndSign need to check anything?
	mocks.ETH.SendAndSignTransactionFunc = func(_ context.Context, _ goEth.CallMsg) (string, error) {
		return "", nil
	}

	sender := randomSender(validators, nodeCount)
	contractAddress := randomSender(validators, nodeCount)

	_, err := nodes[0].Query(
		[]string{
			ethTypes.QuerierRoute,
			ethKeeper.SendCommand,
		},
		abci.RequestQuery{
			Data: testutils.Codec().MustMarshalJSON(
				ethTypes.CommandParams{
					CommandID:    ethTypes.CommandID(commandID),
					Sender:       sender.String(),
					ContractAddr: contractAddress.String(),
				})},
	)
	assert.NoError(t, err)
}
