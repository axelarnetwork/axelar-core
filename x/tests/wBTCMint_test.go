package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/axelarnetwork/axelar-core/store"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	balanceKeeper "github.com/axelarnetwork/axelar-core/x/balance/keeper"
	balanceTypes "github.com/axelarnetwork/axelar-core/x/balance/types"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	"github.com/axelarnetwork/axelar-core/x/broadcast"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	"github.com/axelarnetwork/axelar-core/x/ethereum"
	ethKeeper "github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	ethMock "github.com/axelarnetwork/axelar-core/x/ethereum/types/mock"
	"github.com/axelarnetwork/axelar-core/x/snapshot"
	snapshotKeeper "github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
	"github.com/axelarnetwork/axelar-core/x/tss"
	tssKeeper "github.com/axelarnetwork/axelar-core/x/tss/keeper"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	tssdMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	"github.com/axelarnetwork/axelar-core/x/vote"
	voteKeeper "github.com/axelarnetwork/axelar-core/x/vote/keeper"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
)

const nodeCount2 = 10

// globally available storage variables to control the behaviour of the mocks
var (
	// set of validators2 known to the staking keeper
	validators2 = make([]staking.Validator, 0, nodeCount2)
)

type testMocks2 struct {
	BTC    *btcMock.RPCClientMock
	ETH    *ethMock.RPCClientMock
	Keygen *tssdMock.TSSDKeyGenClientMock
	Sign   *tssdMock.TSSDSignClientMock
	Staker *snapMock.StakingKeeperMock
	TSSD   *tssdMock.TSSDClientMock
}

// 0. Create and start a chain
// 1. Get a deposit address for the given Ethereum recipient address
// 2. Track the new deposit address
// 3. Send BTC to the deposit address and wait until confirmed
// 4. Collect all information that needs to be verified about the deposit
// 5. Verify the previously received information
// 6. Wait until verification is complete
// 7. Sign all pending transfers to Ethereum
// 8. Submit the minting command from an externally controlled address to AxelarGateway

func Test_wBTC_mint(t *testing.T) {
	// 0. Create and start a chain
	chain := fake.NewBlockchain().WithBlockTimeOut(10 * time.Millisecond)

	stringGen := testutils.RandStrings(5, 50).Distinct()
	defer stringGen.Stop()

	mocks := createMocks2()

	var nodes []fake.Node
	for i, valAddr := range stringGen.Take(nodeCount2) {
		validator := staking.Validator{
			OperatorAddress: sdk.ValAddress(valAddr),
			Tokens:          sdk.TokensFromConsensusPower(testutils.RandIntBetween(100, 1000)),
			Status:          sdk.Bonded,
		}
		validators2 = append(validators2, validator)
		nodes = append(nodes, newNode2("node"+strconv.Itoa(i), validator.OperatorAddress, mocks, chain))
		chain.AddNodes(nodes[i])
	}
	// Check to suppress any nil warnings from IDEs
	if nodes == nil {
		panic("need at least one node")
	}

	chain.Start()

	// register proxies
	for i := 0; i < nodeCount2; i++ {
		res := <-chain.Submit(broadcastTypes.MsgRegisterProxy{
			Principal: validators2[i].OperatorAddress,
			Proxy:     sdk.AccAddress(stringGen.Next()),
		})
		assert.NoError(t, res.Error)
	}

	// take first validator snapshot
	res := <-chain.Submit(snapTypes.MsgSnapshot{Sender: randomSender2()})
	assert.NoError(t, res.Error)

	// set up tssd mock for first keygen
	masterKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	mocks.Keygen.RecvFunc = func() (*tssd.MessageOut, error) {
		pk, _ := convert.PubkeyToBytes(masterKey.PublicKey)
		return &tssd.MessageOut{
			Data: &tssd.MessageOut_KeygenResult{KeygenResult: pk}}, nil
	}
	// ensure all nodes call .Send() and .CloseSend()
	sendTimeout, sendCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	closeTimeout, closeCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	mocks.Keygen.SendFunc = func(_ *tssd.MessageIn) error {
		// Q: This is never true
		if len(mocks.Keygen.SendCalls()) == nodeCount2 {
			sendCancel()
		}
		return nil
	}
	mocks.Keygen.CloseSendFunc = func() error {
		if len(mocks.Keygen.CloseSendCalls()) == nodeCount2 {
			closeCancel()
		}
		return nil
	}

	// create first key
	masterKeyID := stringGen.Next()
	res = <-chain.Submit(tssTypes.MsgKeygenStart{
		Sender:    randomSender2(),
		NewKeyID:  masterKeyID,
		Threshold: int(testutils.RandIntBetween(1, int64(len(validators2)))),
	})
	assert.NoError(t, res.Error)

	// assert tssd was properly called
	<-sendTimeout.Done()
	<-closeTimeout.Done()
	assert.Equal(t, nodeCount2, len(mocks.Keygen.SendCalls()))
	assert.Equal(t, nodeCount2, len(mocks.Keygen.CloseSendCalls()))

	// wait for voting to be done
	chain.WaitNBlocks(12)

	// assign key as bitcoin master key
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: randomSender2(),
		Chain:  balance.Bitcoin,
		KeyID:  masterKeyID,
	})
	assert.NoError(t, res.Error)

	// assign key as ethereum master key
	// Q: is this correct? Or distinct masterkeys for different chains need to be created
	res = <-chain.Submit(tssTypes.MsgAssignNextMasterKey{
		Sender: randomSender2(),
		Chain:  balance.Ethereum,
		KeyID:  masterKeyID,
	})
	assert.NoError(t, res.Error)

	// rotate to the first master key
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender2(),
		Chain:  balance.Bitcoin,
	})
	assert.NoError(t, res.Error)

	// rotate to the first master key
	// Q: is this correct?
	res = <-chain.Submit(tssTypes.MsgRotateMasterKey{
		Sender: randomSender2(),
		Chain:  balance.Ethereum,
	})
	assert.NoError(t, res.Error)

	// 1. Get a deposit address for the given Ethereum recipient address
	ethAddr := balance.CrossChainAddress{Chain: balance.Ethereum, Address: testutils.RandStringBetween(5, 20)}
	res = <-chain.Submit(btcTypes.NewMsgLink(randomSender2(), ethAddr))
	assert.NoError(t, res.Error)
	depositAddr := string(res.Data)

	// 2. Track the new deposit address
	res = <-chain.Submit(btcTypes.NewMsgTrackAddress(randomSender2(), depositAddr, true))
	assert.NoError(t, res.Error)

	// 3. Send BTC to the deposit address and wait until confirmed
	txHash, err := chainhash.NewHash(testutils.RandBytes(32))
	if err != nil {
		panic(err)
	}
	voutIdx := uint32(testutils.RandIntBetween(0, 100))
	expectedOut := wire.NewOutPoint(txHash, voutIdx)
	amount := btcutil.Amount(testutils.RandIntBetween(1, 10000000))
	confirmations := uint64(testutils.RandIntBetween(1, 10000))

	mocks.BTC.GetOutPointInfoFunc = func(out *wire.OutPoint) (btcTypes.OutPointInfo, error) {
		if out.String() == expectedOut.String() {
			return btcTypes.OutPointInfo{
				OutPoint:      expectedOut,
				Amount:        amount,
				DepositAddr:   depositAddr,
				Confirmations: confirmations,
			}, nil
		}
		return btcTypes.OutPointInfo{}, fmt.Errorf("tx %s not found", out.String())
	}

	// 4. Collect all information that needs to be verified about the deposit
	bz, err := nodes[0].Query([]string{btcTypes.QuerierRoute, btcKeeper.QueryOutInfo}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(expectedOut)})
	assert.NoError(t, err)
	var info btcTypes.OutPointInfo
	testutils.Codec().MustUnmarshalJSON(bz, &info)

	// 5. Verify the previously received information
	res = <-chain.Submit(btcTypes.NewMsgVerifyTx(randomSender2(), info))
	assert.NoError(t, res.Error)

	// 6. Wait until verification is complete
	chain.WaitNBlocks(12)

	// 7. Sign all pending transfers to Ethereum
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

	closeTimeout, closeCancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	mocks.Sign.CloseSendFunc = func() error {
		if len(mocks.Sign.CloseSendCalls()) == nodeCount {
			closeCancel()
		}
		return nil
	}

	res = <-chain.Submit(ethTypes.NewMsgSignPendingTransfersTx(randomSender2()))
	assert.NoError(t, res.Error)
	// commandID := res.Data

	// // TODO: to be changed with random addresses
	// fromAdderss := "0xE3deF8C6b7E357bf38eC701Ce631f78F2532987A"
	// contractAddress := "0x73ADD47055eba3191fD26285788F8a8b3Fcf9e17"

	// // wait for voting to be done
	// chain.WaitNBlocks(12)

	// // 8. Submit the minting command from an externally controlled address to AxelarGateway
	// bz, err = nodes[0].Query(
	// 	[]string{
	// 		ethTypes.QuerierRoute,
	// 		ethKeeper.SendMintTx,
	// 		commandID,
	// 		fromAdderss,
	// 		contractAddress,
	// 	},
	// 	abci.RequestQuery{})

	// // Error here: createMintTxAndSend -> GetSig returned an error and bz is nil
	// testutils.Codec().MustUnmarshalJSON(bz, &info)

	// // 9. Ensure that minting is done
}

func randomSender2() sdk.AccAddress {
	return sdk.AccAddress(validators2[testutils.RandIntBetween(0, nodeCount2)].OperatorAddress)
}

func newNode2(moniker string, validator sdk.ValAddress, mocks testMocks2, chain *fake.BlockChain) fake.Node {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	broadcaster := fake.NewBroadcaster(testutils.Codec(), validator, chain.Submit)

	snapKeeper := snapshotKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(snapTypes.StoreKey), mocks.Staker)
	voter := voteKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(voteTypes.StoreKey), store.NewSubjectiveStore(), snapKeeper, broadcaster)

	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	bitcoinKeeper := btcKeeper.NewBtcKeeper(testutils.Codec(), sdk.NewKVStoreKey(btcTypes.StoreKey), btcSubspace)
	btcParams := btcTypes.DefaultParams()
	btcParams.Network = mocks.BTC.Network()
	bitcoinKeeper.SetParams(ctx, btcParams)

	ethSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "eth")
	ethereumKeeper := ethKeeper.NewEthKeeper(testutils.Codec(), sdk.NewKVStoreKey(ethTypes.StoreKey), ethSubspace)
	ethParams := ethTypes.DefaultParams()
	// ethParams.Network = mocks.ETH.Network()
	ethereumKeeper.SetParams(ctx, ethParams)

	signer := tssKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(tssTypes.StoreKey), mocks.TSSD,
		params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), tssTypes.DefaultParamspace),
		voter, broadcaster,
	)
	signer.SetParams(ctx, tssTypes.DefaultParams())
	balancer := balanceKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(balanceTypes.StoreKey))

	voter.SetVotingInterval(ctx, voteTypes.DefaultGenesisState().VotingInterval)
	voter.SetVotingThreshold(ctx, voteTypes.DefaultGenesisState().VotingThreshold)

	router := fake.NewRouter()

	broadcastHandler := broadcast.NewHandler(broadcaster)
	btcHandler := bitcoin.NewHandler(bitcoinKeeper, voter, mocks.BTC, signer, snapKeeper, balancer)
	ethHandler := ethereum.NewHandler(ethereumKeeper, mocks.ETH, voter, signer, snapKeeper, balancer)
	snapHandler := snapshot.NewHandler(snapKeeper)
	tssHandler := tss.NewHandler(signer, snapKeeper, voter)
	voteHandler := vote.NewHandler()

	router = router.
		AddRoute(broadcastTypes.RouterKey, broadcastHandler).
		AddRoute(btcTypes.RouterKey, btcHandler).
		AddRoute(ethTypes.RouterKey, ethHandler).
		AddRoute(snapTypes.RouterKey, snapHandler).
		AddRoute(voteTypes.RouterKey, voteHandler).
		AddRoute(tssTypes.RouterKey, tssHandler)

	queriers := map[string]sdk.Querier{
		btcTypes.QuerierRoute: btcKeeper.NewQuerier(bitcoinKeeper, signer, balancer, mocks.BTC),
		ethTypes.QuerierRoute: ethKeeper.NewQuerier(mocks.ETH, ethereumKeeper, signer),
	}

	node := fake.NewNode(moniker, ctx, router, queriers).
		WithEndBlockers(func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
			return vote.EndBlocker(ctx, req, voter)
		})
	return node
}

func createMocks2() testMocks2 {
	stakingKeeper := &snapMock.StakingKeeperMock{
		IterateLastValidatorsFunc: func(ctx sdk.Context, fn func(index int64, validator sdkExported.ValidatorI) (stop bool)) {
			for j, val := range validators2 {
				if fn(int64(j), val) {
					break
				}
			}
		},
		GetLastTotalPowerFunc: func(ctx sdk.Context) sdk.Int {
			totalPower := sdk.ZeroInt()
			for _, val := range validators2 {
				totalPower = totalPower.AddRaw(val.ConsensusPower())
			}
			return totalPower
		},
	}

	btcClient := &btcMock.RPCClientMock{
		ImportAddressRescanFunc: func(string, string, bool) error { return nil },
		SendRawTransactionFunc: func(tx *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
			hash := tx.TxHash()
			return &hash, nil
		},
		NetworkFunc: func() btcTypes.Network { return btcTypes.Mainnet }}

	ethClient := &ethMock.RPCClientMock{
		// TODO add functions
	}

	keygen := &tssdMock.TSSDKeyGenClientMock{}
	sign := &tssdMock.TSSDSignClientMock{}
	tssdClient := &tssdMock.TSSDClientMock{
		KeygenFunc: func(context.Context, ...grpc.CallOption) (tssd.GG18_KeygenClient, error) { return keygen, nil },
		SignFunc:   func(context.Context, ...grpc.CallOption) (tssd.GG18_SignClient, error) { return sign, nil },
	}
	return testMocks2{
		BTC:    btcClient,
		ETH:    ethClient,
		TSSD:   tssdClient,
		Keygen: keygen,
		Sign:   sign,
		Staker: stakingKeeper,
	}
}
