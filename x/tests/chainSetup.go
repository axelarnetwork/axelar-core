package tests

import (
	"context"
	"strconv"
	"time"

	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/x/ethereum"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"

	"github.com/axelarnetwork/axelar-core/store"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	"github.com/axelarnetwork/axelar-core/x/broadcast"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
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
)

func randomSender() sdk.AccAddress {
	return testutils.RandBytes(int(testutils.RandIntBetween(5, 50)))
}

type testMocks struct {
	BTC     *btcMock.RPCClientMock
	ETH     *ethMock.RPCClientMock
	Keygen  *tssdMock.TSSDKeyGenClientMock
	Sign    *tssdMock.TSSDSignClientMock
	Staker  *snapMock.StakingKeeperMock
	TSSD    *tssdMock.TSSDClientMock
	Slasher *snapMock.SlasherMock
}

type nodeData struct {
	Node      *fake.Node
	Validator staking.Validator
	Mocks     testMocks
}

func newNode(moniker string, validator sdk.ValAddress, mocks testMocks, chain *fake.BlockChain) *fake.Node {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	broadcaster := fake.NewBroadcaster(testutils.Codec(), validator, chain.Submit)

	snapSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snap")
	snapKeeper := snapshotKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(snapTypes.StoreKey), snapSubspace, mocks.Staker, mocks.Slasher)
	snapKeeper.SetParams(ctx, snapTypes.DefaultParams())
	voter := voteKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(voteTypes.StoreKey), store.NewSubjectiveStore(), snapKeeper, broadcaster)

	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	bitcoinKeeper := btcKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(btcTypes.StoreKey), btcSubspace)
	btcParams := btcTypes.DefaultParams()
	btcParams.Network = mocks.BTC.Network()
	bitcoinKeeper.SetParams(ctx, btcParams)

	ethSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "eth")
	ethereumKeeper := ethKeeper.NewEthKeeper(testutils.Codec(), sdk.NewKVStoreKey(ethTypes.StoreKey), ethSubspace)
	ethereumKeeper.SetParams(ctx, ethTypes.DefaultParams())

	signer := tssKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(tssTypes.StoreKey), mocks.TSSD,
		params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), tssTypes.DefaultParamspace),
		voter, broadcaster,
	)
	signer.SetParams(ctx, tssTypes.DefaultParams())

	nexusSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("balanceKey"), sdk.NewKVStoreKey("tbalanceKey"), "balance")
	nexusK := nexusKeeper.NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(nexusTypes.StoreKey), nexusSubspace)
	nexusK.SetParams(ctx, nexusTypes.DefaultParams())

	voter.SetVotingInterval(ctx, voteTypes.DefaultGenesisState().VotingInterval)
	voter.SetVotingThreshold(ctx, voteTypes.DefaultGenesisState().VotingThreshold)

	router := fake.NewRouter()

	broadcastHandler := broadcast.NewHandler(broadcaster)
	btcHandler := bitcoin.NewHandler(bitcoinKeeper, voter, mocks.BTC, signer, snapKeeper, nexusK)
	ethHandler := ethereum.NewHandler(ethereumKeeper, mocks.ETH, voter, signer, snapKeeper, nexusK)
	snapHandler := snapshot.NewHandler(snapKeeper)
	tssHandler := tss.NewHandler(signer, snapKeeper, nexusK, voter, mocks.Staker)
	voteHandler := vote.NewHandler()

	router = router.
		AddRoute(broadcastTypes.RouterKey, broadcastHandler).
		AddRoute(btcTypes.RouterKey, btcHandler).
		AddRoute(ethTypes.RouterKey, ethHandler).
		AddRoute(snapTypes.RouterKey, snapHandler).
		AddRoute(voteTypes.RouterKey, voteHandler).
		AddRoute(tssTypes.RouterKey, tssHandler)

	queriers := map[string]sdk.Querier{
		btcTypes.QuerierRoute: btcKeeper.NewQuerier(bitcoinKeeper, signer, nexusK, mocks.BTC),
		ethTypes.QuerierRoute: ethKeeper.NewQuerier(mocks.ETH, ethereumKeeper, signer),
	}

	node := fake.NewNode(moniker, ctx, router, queriers).
		WithEndBlockers(func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
			return vote.EndBlocker(ctx, req, voter)
		})
	return node
}

func createMocks(validators []staking.Validator) testMocks {

	slasher := &snapMock.SlasherMock{
		GetValidatorSigningInfoFunc: func(ctx sdk.Context, address sdk.ConsAddress) (snapTypes.ValidatorInfo, bool) {
			newInfo := slashingTypes.NewValidatorSigningInfo(
				address,
				int64(0),        // height at which validator was first a candidate OR was unjailed
				int64(3),        // index offset into signed block bit array. TODO: check if needs to be set correctly.
				time.Unix(0, 0), // jailed until
				false,           // tomstoned
				int64(0),        // missed blocks
			)
			return snapTypes.ValidatorInfo{ValidatorSigningInfo: newInfo}, true
		},
	}

	stakingKeeper := &snapMock.StakingKeeperMock{
		IterateLastValidatorsFunc: func(ctx sdk.Context, fn func(index int64, validator sdkExported.ValidatorI) (stop bool)) {
			for j, val := range validators {
				if fn(int64(j), val) {
					break
				}
			}
		},
		GetLastTotalPowerFunc: func(ctx sdk.Context) sdk.Int {
			totalPower := sdk.ZeroInt()
			for _, val := range validators {
				totalPower = totalPower.AddRaw(val.ConsensusPower())
			}
			return totalPower
		},
	}

	btcClient := &btcMock.RPCClientMock{
		SendRawTransactionFunc: func(tx *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
			hash := tx.TxHash()
			return &hash, nil
		},
		NetworkFunc: func() btcTypes.Network { return btcTypes.Mainnet }}

	ethClient := &ethMock.RPCClientMock{
		// TODO add functions when needed
	}

	keygen := &tssdMock.TSSDKeyGenClientMock{
		SendFunc:      func(*tssd.MessageIn) error { return nil },
		CloseSendFunc: func() error { return nil }}
	sign := &tssdMock.TSSDSignClientMock{}
	tssdClient := &tssdMock.TSSDClientMock{
		KeygenFunc: func(context.Context, ...grpc.CallOption) (tssd.GG18_KeygenClient, error) { return keygen, nil },
		SignFunc:   func(context.Context, ...grpc.CallOption) (tssd.GG18_SignClient, error) { return sign, nil },
	}
	return testMocks{
		BTC:     btcClient,
		ETH:     ethClient,
		TSSD:    tssdClient,
		Keygen:  keygen,
		Sign:    sign,
		Staker:  stakingKeeper,
		Slasher: slasher,
	}
}

// initChain Creates a chain with given number of validators
func initChain(nodeCount int) (*fake.BlockChain, []nodeData) {
	stringGen := testutils.RandStrings(5, 50).Distinct()
	defer stringGen.Stop()

	// create nodes
	var validators []staking.Validator
	for _, valAddr := range stringGen.Take(nodeCount) {
		// assign validators
		validator := staking.Validator{
			OperatorAddress: sdk.ValAddress(valAddr),
			Tokens:          sdk.TokensFromConsensusPower(testutils.RandIntBetween(100, 1000)),
			Status:          sdk.Bonded,
			ConsPubKey:      ed25519.GenPrivKey().PubKey(),
		}
		validators = append(validators, validator)
	}
	// create a chain
	chain := fake.NewBlockchain().WithBlockTimeOut(10 * time.Millisecond)
	var data []nodeData
	for i, validator := range validators {
		// create mocks
		mocks := createMocks(validators)

		// assign nodes
		node := newNode("node"+strconv.Itoa(i), validator.OperatorAddress, mocks, chain)
		chain.AddNodes(node)
		data = append(data, nodeData{Node: node, Validator: validator, Mocks: mocks})
	}

	// start chain
	chain.Start()

	return chain, data
}
