package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"strconv"
	"testing"
	"time"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/assert"
)

// createChain Creates a chain with given number of validators
func createChain(nodeCount int, stringGen *testutils.RandDistinctStringGen) (*fake.BlockChain, []staking.Validator, testMocks, []fake.Node) {

	validators := make([]staking.Validator, 0, nodeCount)
	chain := fake.NewBlockchain().WithBlockTimeOut(10 * time.Millisecond)

	mocks := createMocks(&validators)

	var nodes []fake.Node
	for i, valAddr := range stringGen.Take(nodeCount) {
		validator := staking.Validator{
			OperatorAddress: sdk.ValAddress(valAddr),
			Tokens:          sdk.TokensFromConsensusPower(testutils.RandIntBetween(100, 1000)),
			Status:          sdk.Bonded,
		}
		validators = append(validators, validator)
		nodes = append(nodes, newNode("node"+strconv.Itoa(i), validator.OperatorAddress, mocks, chain))
		chain.AddNodes(nodes[i])
	}
	// Check to suppress any nil warnings from IDEs
	if nodes == nil {
		panic("need at least one node")
	}

	chain.Start()
	return chain, validators, mocks, nodes
}

// registerProxies registers
func registerProxies(chain *fake.BlockChain,
	validators []staking.Validator,
	nodeCount int,
	stringGen *testutils.RandDistinctStringGen,
	t *testing.T) {
	for i := 0; i < nodeCount; i++ {
		res := <-chain.Submit(broadcastTypes.MsgRegisterProxy{
			Principal: validators[i].OperatorAddress,
			Proxy:     sdk.AccAddress(stringGen.Next()),
		})
		assert.NoError(t, res.Error)
	}

}

// takeSnapshot takes a snapshot of the current validators
func takeSnapshot(chain *fake.BlockChain, validators []staking.Validator, nodeCount int64, t *testing.T) {
	res := <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender(validators, nodeCount)})
	assert.NoError(t, res.Error)
}

// setTssdMock sets up tssd mock for btc keygen
func generateKey() *ecdsa.PrivateKey {
	// set up tssd mock for btc keygen
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
	// assign bitcoin master key
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
