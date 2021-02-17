package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/assert"
)

// takeSnapshot takes a snapshot of the current validators
func takeSnapshot(chain *fake.BlockChain, validators []staking.Validator, nodeCount int64, t *testing.T) {
	res := <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender(validators, nodeCount)})
	assert.NoError(t, res.Error)
}

// createMasterKeyID creates a master key ID and guarantees that .Send and
// .Close are called by all nodes
func createMasterKeyID(
	chain *fake.BlockChain,
	validators []staking.Validator,
	nodeCount int,
	stringGen *testutils.RandDistinctStringGen,
	mocks testMocks) (*fake.Result, string) {

	// ensure all nodes call .Send()
	prevSendCount := len(mocks.Keygen.SendCalls())
	sendTimeout, sendCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer sendCancel()
	if len(mocks.Keygen.SendCalls())-prevSendCount == nodeCount {
		sendCancel()
	}

	// ensure all nodes call .Close()
	prevCloseCount := len(mocks.Keygen.CloseSendCalls())
	closeTimeout, closeCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer closeCancel()
	if len(mocks.Keygen.CloseSendCalls())-prevCloseCount == nodeCount {
		closeCancel()
	}

	// create and submit master key id
	masterKeyID := stringGen.Next()
	res := <-chain.Submit(tssTypes.MsgKeygenStart{
		Sender:    randomSender(validators, int64(nodeCount)),
		NewKeyID:  masterKeyID,
		Threshold: int(testutils.RandIntBetween(1, int64(len(validators)))),
	})

	// assert tssd was properly called
	<-sendTimeout.Done()
	<-closeTimeout.Done()

	return res, masterKeyID
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
