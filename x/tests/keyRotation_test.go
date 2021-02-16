package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/utils/denom"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const nodeCount = 10

// globally available storage variables to control the behaviour of the mocks
var (
	// set of validators known to the staking keeper
	validators = make([]staking.Validator, 0, nodeCount)
)

type testMocks struct {
	BTC     *btcMock.RPCClientMock
	Keygen  *tssdMock.TSSDKeyGenClientMock
	Sign    *tssdMock.TSSDSignClientMock
	Staker  *snapMock.StakingKeeperMock
	TSSD    *tssdMock.TSSDClientMock
	Slasher *snapMock.SlasherMock
}

// Testing the key rotation functionality.
// (0. Register proxies for all validators)
//  1. Create an initial validator snapshot
//  2. Create a key (wait for vote)
//  3. Designate that key to be the first master key for bitcoin
//  4. Rotate to the designated master key
//  5. Simulate bitcoin deposit to the current master key
//  6. Query deposit tx info
//  7. Verify the deposit is confirmed on bitcoin (wait for vote)
//  8. Create a second snapshot
//  9. Create a new key with the second snapshot's validator set (wait for vote)
// 10. Designate that key to be the next master key for bitcoin
// 11. Sign a consolidation transaction (wait for vote)
// 12. Send the signed transaction to bitcoin
// 13. Query transfer tx info
// 14. Verify the fund transfer is confirmed on bitcoin (wait for vote)
// 15. Rotate to the new master key
func TestKeyRotation(t *testing.T) {

	// set up chain
	const nodeCount = 10

	stringGen := testutils.RandStrings(5, 50).Distinct()
	defer stringGen.Stop()

	chain, validators, mocks, nodes := createChain(nodeCount, &stringGen)

	// register proxies for all validators
	registerProxies(chain, validators, nodeCount, &stringGen, t)
	takeSnapshot(chain, validators, nodeCount, t)

	// create master key for btc
	masterKeyID, masterKey := createMasterKeyID(chain, validators, nodeCount, &stringGen, mocks, t)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// assign bitcoin master key
	assignMasterKey(chain, validators, nodeCount, masterKeyID, balance.Bitcoin, t)
	// rotate to the first btc master key
	rotateMasterKey(chain, validators, nodeCount, balance.Bitcoin, t)

	// get deposit address for ethereum transfer
	depositAddr, _ := getCrossChainAddress(balance.CrossChainAddress{}, balance.Ethereum, chain, validators, nodeCount, t)

	// simulate deposit to master key address
	blockHash, expectedOut, outPointInfo := sendBTCtoDepositAddress(depositAddr, mocks)

	// query for deposit info
	info := queryOutPointInfo(nodes, blockHash, expectedOut, t)

	// verify deposit to master key
	verifyTx(chain, validators, nodeCount, info, t)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// second snapshot
	takeSnapshot(chain, validators, nodeCount, t)

	// create another master key
	keyID2, _ := createMasterKeyID(chain, validators, nodeCount, &stringGen, mocks, t)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// assign second key to be the second master key
	assignMasterKey(chain, validators, nodeCount, keyID2, balance.Bitcoin, t)

	// sign transaction
	signTx(chain, validators, expectedOut, rawTx, nodeCount, mocks, masterKeyID, masterKey, t)

	// wait for voting to be done
	chain.WaitNBlocks(22)

	// send tx to Bitcoin
	bz, err = nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.SendTx},
		abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(expectedOut)})
	assert.NoError(t, err)

	// set up btc mock to return the new tx
	var transferHash *chainhash.Hash
	testutils.Codec().MustUnmarshalJSON(bz, &transferHash)
	blockHash, err = chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	voutIdx := uint32(0)
	confirmations := uint64(testutils.RandIntBetween(1, 10000))
	transferOut := wire.NewOutPoint(transferHash, voutIdx)
	mocks.BTC.GetOutPointInfoFunc = func(_ *chainhash.Hash, out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
		if out.String() == transferOut.String() {
			return btcTypes.OutPointInfo{
				OutPoint:      transferOut,
				BlockHash:     blockHash,
				Amount:        amount,
				Address:       consAddr,
				Confirmations: confirmations,
			}, nil
		}

		return btcTypes.OutPointInfo{}, fmt.Errorf("tx %s not found", out.String())
	}

	// query for transfer info
	info = queryOutPointInfo(nodes, blockHash, transferOut, t)

	// verify master key transfer
	verifyTx(chain, validators, nodeCount, info, t)

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// rotate master key to key 2
	rotateMasterKey(chain, validators, nodeCount, balance.Bitcoin, t)
}

func queryRawTx(nodes []fake.Node, expectedOut *wire.OutPoint, consAddr string, prevAmount int64, t *testing.T) (*wire.MsgTx, btcutil.Amount) {
	amount := btcutil.Amount(prevAmount - testutils.RandIntBetween(1, prevAmount-1))

	bz, err := nodes[0].Query(
		[]string{btcTypes.QuerierRoute, btcKeeper.QueryRawTx},
		abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(
			btcTypes.RawTxParams{
				OutPoint:    expectedOut,
				DepositAddr: consAddr,
				Satoshi:     sdk.NewInt64Coin(denom.Sat, int64(amount)),
			})},
	)
	assert.NoError(t, err)
	var rawTx *wire.MsgTx
	testutils.Codec().MustUnmarshalJSON(bz, &rawTx)
	return rawTx, amount
}

func signTx(
	chain *fake.BlockChain,
	validators []staking.Validator,
	expectedOut *wire.OutPoint,
	rawTx *wire.MsgTx,
	nodeCount int,
	mocks testMocks,
	masterKeyID string,
	masterKey *ecdsa.PrivateKey,
	t *testing.T) {
	// set up tssd mock for signing
	msgToSign := make(chan []byte, nodeCount)
	mocks.Sign.SendFunc = func(messageIn *tssd.MessageIn) error {
		assert.Equal(t, masterKeyID, messageIn.GetSignInit().KeyUid)
		msgToSign <- messageIn.GetSignInit().MessageToSign
		return nil
	}
	sigChan := make(chan []byte, 1)
	go func() {
		r, s, err := ecdsa.Sign(rand.Reader, masterKey, <-msgToSign)
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

	// ensure and .CloseSend() is called
	closeTimeout, closeCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	mocks.Sign.CloseSendFunc = func() error {
		if len(mocks.Sign.CloseSendCalls()) == nodeCount {
			closeCancel()
		}
		return nil
	}

	// sign transfer tx
	res := <-chain.Submit(btcTypes.NewMsgSignTx(
		randomSender(validators, int64(nodeCount)),
		expectedOut,
		rawTx))
	assert.NoError(t, res.Error)
	// assert tssd was properly called
	<-closeTimeout.Done()
	assert.Equal(t, nodeCount, len(mocks.Sign.CloseSendCalls()))
}
