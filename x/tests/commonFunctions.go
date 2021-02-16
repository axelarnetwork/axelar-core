package tests

import (
	"strconv"
	"testing"
	"time"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
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
