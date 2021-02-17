package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"testing"
	"time"

	"github.com/axelarnetwork/axelar-core/testutils"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	ethKeeper "github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
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

	// take snapshot
	res := <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender(validators, nodeCount)})
	assert.NoError(t, res.Error)

	// create master keys for btc
	btcMasterKey := generateKey()
	mocks.Keygen.SendFunc = func(_ *tssd.MessageIn) error {
		return nil
	}
	mocks.Keygen.CloseSendFunc = func() error {
		return nil
	}
	mocks.Keygen.RecvFunc = func() (*tssd.MessageOut, error) {
		pk, _ := convert.PubkeyToBytes(btcMasterKey.PublicKey)
		return &tssd.MessageOut{
			Data: &tssd.MessageOut_KeygenResult{KeygenResult: pk}}, nil
	}
	res, btcMasterKeyID := createMasterKeyID(chain, validators, nodeCount, &stringGen, mocks)
	assert.NoError(t, res.Error)
	assert.Equal(t, nodeCount, len(mocks.Keygen.SendCalls()))
	assert.Equal(t, nodeCount, len(mocks.Keygen.CloseSendCalls()))

	// create master keys for eth
	ethMasterKey := generateKey()
	mocks.Keygen.RecvFunc = func() (*tssd.MessageOut, error) {
		pk, _ := convert.PubkeyToBytes(ethMasterKey.PublicKey)
		return &tssd.MessageOut{
			Data: &tssd.MessageOut_KeygenResult{KeygenResult: pk}}, nil
	}
	res, ethMasterKeyID := createMasterKeyID(chain, validators, nodeCount, &stringGen, mocks)
	assert.NoError(t, res.Error)
	assert.Equal(t, 2*nodeCount, len(mocks.Keygen.SendCalls()))
	assert.Equal(t, 2*nodeCount, len(mocks.Keygen.CloseSendCalls()))

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// assign bitcoin master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: randomSender(validators, nodeCount),
		Chain:  balance.Bitcoin,
		KeyID:  btcMasterKeyID,
	})
	assert.NoError(t, res.Error)

	// rotate to the first btc master key
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender(validators, nodeCount),
		Chain:  balance.Bitcoin,
	})
	assert.NoError(t, res.Error)

	// assign key as ethereum master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: randomSender(validators, nodeCount),
		Chain:  balance.Ethereum,
		KeyID:  ethMasterKeyID,
	})
	assert.NoError(t, res.Error)

	// rotate to the first eth master key
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender(validators, nodeCount),
		Chain:  balance.Ethereum,
	})
	assert.NoError(t, res.Error)

	// steps followed as per https://github.com/axelarnetwork/axelarate#mint-erc20-wrapped-bitcoin-tokens-on-ethereum

	// 1. Get a deposit address for an Ethereum recipient address
	// we don't provide an actual recipient address, so it is created automatically
	crosschainAddr := balance.CrossChainAddress{Chain: balance.Ethereum, Address: testutils.RandStringBetween(5, 20)}
	res = <-chain.Submit(btcTypes.NewMsgLink(randomSender(validators, nodeCount), crosschainAddr))
	assert.NoError(t, res.Error)
	depositAddr := string(res.Data)

	// 2. Send BTC to the deposit address and wait until confirmed
	blockHash, expectedOut, _ := sendBTCtoDepositAddress(depositAddr, mocks)

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
	// commandID := signPendingTransfersTx(chain, validators, nodeCount, mocks, ethMasterKeyID, ethMasterKey, t)
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

	// wait for voting to be done
	// Q: Why do we have to wait for 22 blocks instead of 12?
	chain.WaitNBlocks(22)

	// 7. Submit the minting command from an externally controlled address to AxelarGateway
	// Q: Does SendAndSign need to check anything?
	mocks.ETH.SendAndSignTransactionFunc = func(_ context.Context, _ goEth.CallMsg) (string, error) {
		return "", nil
	}

	sender := randomSender(validators, nodeCount)
	contractAddress := randomSender(validators, nodeCount)

	_, err = nodes[0].Query(
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
