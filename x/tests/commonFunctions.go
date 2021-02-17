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
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
)

// takeSnapshot takes a snapshot of the current validators
func takeSnapshot(chain *fake.BlockChain, validators []staking.Validator, nodeCount int64, t *testing.T) {
	res := <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender(validators, nodeCount)})
	assert.NoError(t, res.Error)
}

// setTssdMock sets up tssd mock for btc keygen
func generateKey() *ecdsa.PrivateKey {
	masterKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	return masterKey
}

// createMasterKeyID creates a master key ID and guarantees that .Send and
// .Close are called by all nodes
func createMasterKeyID(
	chain *fake.BlockChain,
	validators []staking.Validator,
	nodeCount int,
	stringGen *testutils.RandDistinctStringGen,
	mocks testMocks,
	t *testing.T) (string, *ecdsa.PrivateKey) {

	masterKey := generateKey()

	mocks.Keygen.RecvFunc = func() (*tssd.MessageOut, error) {
		pk, _ := convert.PubkeyToBytes(masterKey.PublicKey)
		return &tssd.MessageOut{
			Data: &tssd.MessageOut_KeygenResult{KeygenResult: pk}}, nil
	}

	// hold the number of times Send and Close has been already called
	prevSendCount := len(mocks.Keygen.SendCalls())
	prevCloseCount := len(mocks.Keygen.CloseSendCalls())

	// ensure all nodes call .Send()
	sendTimeout, sendCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	mocks.Keygen.SendFunc = func(_ *tssd.MessageIn) error {
		if len(mocks.Keygen.SendCalls()) == nodeCount {
			sendCancel()
		}
		return nil
	}

	// ensure all nodes call .Close()
	closeTimeout, closeCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	mocks.Keygen.CloseSendFunc = func() error {
		if len(mocks.Keygen.CloseSendCalls()) == nodeCount {
			closeCancel()
		}
		return nil
	}

	// create and submit master key id
	masterKeyID := stringGen.Next()
	res := <-chain.Submit(tssTypes.MsgKeygenStart{
		Sender:    randomSender(validators, int64(nodeCount)),
		NewKeyID:  masterKeyID,
		Threshold: int(testutils.RandIntBetween(1, int64(len(validators)))),
	})
	assert.NoError(t, res.Error)

	// assert tssd was properly called
	<-sendTimeout.Done()
	<-closeTimeout.Done()
	assert.Equal(t, nodeCount, len(mocks.Keygen.SendCalls())-prevSendCount)
	assert.Equal(t, nodeCount, len(mocks.Keygen.CloseSendCalls())-prevCloseCount)

	return masterKeyID, masterKey
}

// assignMasterKey assigns a master key of a chain
func assignMasterKey(
	chain *fake.BlockChain,
	validators []staking.Validator,
	nodeCount int64,
	masterKeyID string,
	balanceChain balance.Chain,
	t *testing.T) {

	res := <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: randomSender(validators, nodeCount),
		Chain:  balanceChain,
		KeyID:  masterKeyID,
	})
	assert.NoError(t, res.Error)
}

// rotateMasterKey rotates a master key of a chain
func rotateMasterKey(
	chain *fake.BlockChain,
	validators []staking.Validator,
	nodeCount int64,
	balanceChain balance.Chain,
	t *testing.T) {

	res := <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender(validators, nodeCount),
		Chain:  balanceChain,
	})
	assert.NoError(t, res.Error)
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

// verifyTX verifies the deposit information
func verifyTx(chain *fake.BlockChain, validators []staking.Validator, nodeCount int64, info btcTypes.OutPointInfo, t *testing.T) {
	res := <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender(validators, nodeCount), info))
	assert.NoError(t, res.Error)
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
