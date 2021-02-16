package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/axelarnetwork/axelar-core/testutils"
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

	const nodeCount = 10

	stringGen := testutils.RandStrings(5, 50).Distinct()
	defer stringGen.Stop()

	chain, validators, mocks, nodes := createChain(nodeCount, &stringGen)

	registerProxies(chain, validators, nodeCount, &stringGen, t)

	takeSnapshot(chain, validators, nodeCount, t)

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

	// 1. Get a deposit address for the given Ethereum recipient address
	ethAddr := balance.CrossChainAddress{Chain: balance.Ethereum, Address: testutils.RandStringBetween(5, 20)}
	res := <-chain.Submit(btcTypes.NewMsgLink(randomSender(validators, nodeCount), ethAddr))
	assert.NoError(t, res.Error)
	depositAddr := string(res.Data)

	// 2. Send BTC to the deposit address and wait until confirmed
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

	// 3. Collect all information that needs to be verified about the deposit
	bz, err := nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.QueryOutInfo, blockHash.String()}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(expectedOut)})
	assert.NoError(t, err)
	var info btcTypes.OutPointInfo
	testutils.Codec().MustUnmarshalJSON(bz, &info)

	// 4. Verify the previously received information
	res = <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender(validators, nodeCount), info))
	assert.NoError(t, res.Error)

	// 5. Wait until verification is complete
	chain.WaitNBlocks(12)

	// 6. Sign all pending transfers to Ethereum
	// set up tssd mock for signing
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

	res = <-chain.Submit(ethTypes.NewMsgSignPendingTransfersTx(randomSender(validators, nodeCount)))
	assert.NoError(t, res.Error)
	commandID := common.BytesToHash(res.Data)
	<-closeTimeout.Done()
	assert.Equal(t, nodeCount, len(mocks.Sign.CloseSendCalls()))

	sender := randomSender(validators, nodeCount)
	contractAddress := randomSender(validators, nodeCount)

	// wait for voting to be done
	// Q: Why do we have to wait for 22 blocks instead of 12?
	chain.WaitNBlocks(22)

	// Q: Does SendAndSign need to check anything?
	mocks.ETH.SendAndSignTransactionFunc = func(_ context.Context, _ goEth.CallMsg) (string, error) {
		return "", nil
	}

	// 7. Submit the minting command from an externally controlled address to AxelarGateway
	bz, err = nodes[0].Query(
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
