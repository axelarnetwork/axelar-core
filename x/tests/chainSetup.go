package tests

import (
	"context"
	"fmt"
	mathRand "math/rand"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	goEthTypes "github.com/ethereum/go-ethereum/core/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/cmd/vald/eth"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/ethereum"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/broadcast"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	ethKeeper "github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	ethMock "github.com/axelarnetwork/axelar-core/x/ethereum/types/mock"
	"github.com/axelarnetwork/axelar-core/x/snapshot"
	snapshotExported "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapshotExportedMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	snapshotKeeper "github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	snapshotTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	snapshotTypesMock "github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
	"github.com/axelarnetwork/axelar-core/x/tss"
	tssKeeper "github.com/axelarnetwork/axelar-core/x/tss/keeper"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	tssMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	"github.com/axelarnetwork/axelar-core/x/vote"
	voteKeeper "github.com/axelarnetwork/axelar-core/x/vote/keeper"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

func randomSender() sdk.AccAddress {
	return rand.Bytes(int(rand.I64Between(5, 50)))
}
func randomEthSender() common.Address {
	return common.BytesToAddress(rand.Bytes(common.AddressLength))
}

type testMocks struct {
	ETH     *ethMock.RPCClientMock
	Keygen  *tssMock.TofndKeyGenClientMock
	Sign    *tssMock.TofndSignClientMock
	Staker  *snapshotTypesMock.StakingKeeperMock
	Tofnd   *tssMock.TofndClientMock
	Slasher *snapshotExportedMock.SlasherMock
	Tss     *snapshotExportedMock.TssMock
}

type nodeData struct {
	Node        *fake.Node
	Validator   stakingtypes.Validator
	Mocks       testMocks
	Broadcaster fake.Broadcaster
}

func newNode(moniker string, broadcaster fake.Broadcaster, mocks testMocks) *fake.Node {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger().With("node", moniker))
	encCfg := testutils.MakeEncodingConfig()

	snapSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snap")
	snapKeeper := snapshotKeeper.NewKeeper(encCfg.Amino, sdk.NewKVStoreKey(snapshotTypes.StoreKey), snapSubspace, broadcaster, mocks.Staker, mocks.Slasher, mocks.Tss)
	snapKeeper.SetParams(ctx, snapshotTypes.DefaultParams())
	voter := voteKeeper.NewKeeper(encCfg.Amino, sdk.NewKVStoreKey(voteTypes.StoreKey), snapKeeper, broadcaster)

	btcSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	bitcoinKeeper := btcKeeper.NewKeeper(encCfg.Amino, sdk.NewKVStoreKey(btcTypes.StoreKey), btcSubspace)
	btcParams := btcTypes.DefaultParams()
	bitcoinKeeper.SetParams(ctx, btcParams)

	ethSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "eth")
	ethereumKeeper := ethKeeper.NewEthKeeper(encCfg.Amino, sdk.NewKVStoreKey(ethTypes.StoreKey), ethSubspace)
	ethereumKeeper.SetParams(ctx, ethTypes.DefaultParams())

	tssSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), tssTypes.DefaultParamspace)
	signer := tssKeeper.NewKeeper(encCfg.Amino, sdk.NewKVStoreKey(tssTypes.StoreKey), tssSubspace, mocks.Slasher)
	signer.SetParams(ctx, tssTypes.DefaultParams())

	nexusSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("balanceKey"), sdk.NewKVStoreKey("tbalanceKey"), "balance")
	nexusK := nexusKeeper.NewKeeper(encCfg.Amino, sdk.NewKVStoreKey(nexusTypes.StoreKey), nexusSubspace)
	nexusK.SetParams(ctx, nexusTypes.DefaultParams())

	voter.SetVotingInterval(ctx, voteTypes.DefaultGenesisState().VotingInterval)
	voter.SetVotingThreshold(ctx, voteTypes.DefaultGenesisState().VotingThreshold)

	router := fake.NewRouter()

	broadcastHandler := broadcast.NewHandler(broadcaster)
	btcHandler := bitcoin.NewHandler(bitcoinKeeper, voter, signer, nexusK, snapKeeper)
	ethHandler := ethereum.NewHandler(ethereumKeeper, voter, signer, nexusK, snapKeeper)
	snapHandler := snapshot.NewHandler()
	tssHandler := tss.NewHandler(signer, snapKeeper, nexusK, voter, &tssMock.StakingKeeperMock{
		GetLastTotalPowerFunc: mocks.Staker.GetLastTotalPowerFunc,
	}, broadcaster)
	voteHandler := vote.NewHandler()

	router = router.
		AddRoute(sdk.NewRoute(broadcastTypes.RouterKey, broadcastHandler)).
		AddRoute(sdk.NewRoute(btcTypes.RouterKey, btcHandler)).
		AddRoute(sdk.NewRoute(ethTypes.RouterKey, ethHandler)).
		AddRoute(sdk.NewRoute(snapshotTypes.RouterKey, snapHandler)).
		AddRoute(sdk.NewRoute(voteTypes.RouterKey, voteHandler)).
		AddRoute(sdk.NewRoute(tssTypes.RouterKey, tssHandler))

	queriers := map[string]sdk.Querier{
		btcTypes.QuerierRoute: btcKeeper.NewQuerier(bitcoinKeeper, signer, nexusK),
		ethTypes.QuerierRoute: ethKeeper.NewQuerier(mocks.ETH, ethereumKeeper, signer),
	}

	node := fake.NewNode(moniker, ctx, router, queriers).
		WithEndBlockers(
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return vote.EndBlocker(ctx, req, voter)
			},
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return bitcoin.EndBlocker(ctx, req, bitcoinKeeper, signer)
			},
		)
	return node
}

func createMocks(validators []stakingtypes.Validator) testMocks {
	slasher := &snapshotExportedMock.SlasherMock{
		GetValidatorSigningInfoFunc: func(ctx sdk.Context, address sdk.ConsAddress) (snapshotExported.ValidatorInfo, bool) {
			newInfo := slashingtypes.NewValidatorSigningInfo(
				address,
				int64(0),        // height at which validator was first a candidate OR was unjailed
				int64(3),        // index offset into signed block bit array. TODO: check if needs to be set correctly.
				time.Unix(0, 0), // jailed until
				false,           // tomstoned
				int64(0),        // missed blocks
			)
			return snapshotExported.ValidatorInfo{ValidatorSigningInfo: newInfo}, true
		},
	}

	stakingKeeper := &snapshotTypesMock.StakingKeeperMock{
		IterateBondedValidatorsByPowerFunc: func(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {
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

	tssK := &snapshotExportedMock.TssMock{
		GetValidatorDeregisteredBlockHeightFunc: func(ctx sdk.Context, valAddr sdk.ValAddress) int64 {
			return 0
		},
	}

	ethClient := &ethMock.RPCClientMock{
		SendAndSignTransactionFunc: func(context.Context, geth.CallMsg) (string, error) {
			return "", nil
		},
		PendingNonceAtFunc: func(context.Context, common.Address) (uint64, error) {
			return mathRand.Uint64(), nil
		},
		SendTransactionFunc: func(context.Context, *gethTypes.Transaction) error { return nil },
	}

	return testMocks{
		ETH:     ethClient,
		Staker:  stakingKeeper,
		Slasher: slasher,
		Tss:     tssK,
	}
}

// initChain Creates a chain with given number of validators
func initChain(nodeCount int, test string) (*fake.BlockChain, []nodeData) {
	stringGen := rand.Strings(5, 50).Distinct()
	encCfg := testutils.MakeEncodingConfig()

	protoPK, err := cryptocodec.FromTmPubKeyInterface(ed25519.GenPrivKey().PubKey())
	if err != nil {
		panic(err)
	}
	consPK, err := codectypes.NewAnyWithValue(protoPK)
	if err != nil {
		panic(err)
	}

	var validators []stakingtypes.Validator
	for _, valAddr := range stringGen.Take(nodeCount) {
		// assign validators
		validator := stakingtypes.Validator{
			OperatorAddress: valAddr,
			Tokens:          sdk.TokensFromConsensusPower(rand.I64Between(100, 1000)),
			Status:          stakingtypes.Bonded,
			ConsensusPubkey: consPK,
		}
		validators = append(validators, validator)
	}
	// create a chain
	chain := fake.NewBlockchain().WithBlockTimeOut(10 * time.Millisecond)

	t := fake.NewTofnd()
	var data []nodeData
	for i, validator := range validators {
		// create mocks
		mocks := createMocks(validators)

		// assign nodes
		oppAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {
			panic(err)
		}
		broadcaster := fake.NewBroadcaster(encCfg.Amino, oppAddr, chain.Submit)

		node := newNode(test+strconv.Itoa(i), broadcaster, mocks)
		chain.AddNodes(node)
		n := nodeData{Node: node, Validator: validator, Mocks: mocks, Broadcaster: broadcaster}

		registerTSSEventListeners(n, t)
		registerBTCEventListener(n)
		registerETHEventListener(n)
		data = append(data, n)
	}

	// start chain
	chain.Start()

	return chain, data
}

func registerBTCEventListener(n nodeData) {
	encCfg := testutils.MakeEncodingConfig()

	// register listener for confirmation
	n.Node.RegisterEventListener(func(event abci.Event) bool {
		if event.Type != btcTypes.EventTypeOutpointConfirmation {
			return false
		}

		m := mapifyAttributes(event)
		if m[sdk.AttributeKeyAction] != btcTypes.AttributeValueStart {
			return false
		}

		var poll voting.PollMeta
		encCfg.Amino.MustUnmarshalJSON([]byte(m[btcTypes.AttributeKeyPoll]), &poll)

		var out btcTypes.OutPointInfo
		encCfg.Amino.MustUnmarshalJSON([]byte(m[btcTypes.AttributeKeyOutPointInfo]), &out)
		err := n.Broadcaster.Broadcast(n.Node.Ctx,
			&btcTypes.MsgVoteConfirmOutpoint{
				Sender:    n.Broadcaster.GetProxy(n.Node.Ctx, n.Broadcaster.LocalPrincipal).String(),
				Poll:      poll,
				Confirmed: true,
				OutPoint:  out.OutPoint,
			})
		if err != nil {
			panic(err)
		}

		return true
	})
}

func registerETHEventListener(n nodeData) {
	encCfg := testutils.MakeEncodingConfig()
	// register listener for deposit confirmation
	n.Node.RegisterEventListener(func(event abci.Event) bool {
		if event.Type != ethTypes.EventTypeDepositConfirmation {
			return false
		}

		m := mapifyAttributes(event)
		if m[sdk.AttributeKeyAction] != ethTypes.AttributeValueStart {
			return false
		}

		var poll voting.PollMeta
		encCfg.Amino.MustUnmarshalJSON([]byte(m[ethTypes.AttributeKeyPoll]), &poll)

		err := n.Broadcaster.Broadcast(n.Node.Ctx,
			&ethTypes.MsgVoteConfirmDeposit{
				Sender:    n.Broadcaster.GetProxy(n.Node.Ctx, n.Broadcaster.LocalPrincipal),
				Poll:      poll,
				Confirmed: true,
				TxID:      m[ethTypes.AttributeKeyTxID],
				BurnAddr:  m[ethTypes.AttributeKeyBurnAddress],
			})
		if err != nil {
			panic(err)
		}

		return true
	})

	// register listener for token deploy confirmation
	n.Node.RegisterEventListener(func(event abci.Event) bool {
		if event.Type != ethTypes.EventTypeTokenConfirmation {
			return false
		}

		m := mapifyAttributes(event)
		if m[sdk.AttributeKeyAction] != ethTypes.AttributeValueStart {
			return false
		}

		var poll voting.PollMeta
		encCfg.Amino.MustUnmarshalJSON([]byte(m[ethTypes.AttributeKeyPoll]), &poll)

		err := n.Broadcaster.Broadcast(n.Node.Ctx,
			&ethTypes.MsgVoteConfirmToken{
				Sender:    n.Broadcaster.GetProxy(n.Node.Ctx, n.Broadcaster.LocalPrincipal),
				Poll:      poll,
				Confirmed: true,
				TxID:      m[ethTypes.AttributeKeyTxID],
				Symbol:    m[ethTypes.AttributeKeySymbol],
			})
		if err != nil {
			panic(err)
		}

		return true
	})
}

func randomOutpointInfo(recipient string) btcTypes.OutPointInfo {
	txHash, err := chainhash.NewHash(rand.Bytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}

	voutIdx := uint32(rand.I64Between(0, 100))
	return btcTypes.OutPointInfo{
		OutPoint: wire.NewOutPoint(txHash, voutIdx).String(),
		Amount:   rand.I64Between(1, 10000000),
		Address:  recipient,
	}
}

func registerTSSEventListeners(n nodeData, t *fake.Tofnd) {
	// register listener for keygen start
	n.Node.RegisterEventListener(func(event abci.Event) bool {
		if event.Type != tssTypes.EventTypeKeygen {
			return false
		}

		m := mapifyAttributes(event)
		if m[sdk.AttributeKeyAction] != tssTypes.AttributeValueStart {
			return false
		}
		if m[tssTypes.AttributeKeyKeyID] == "" {
			return false
		}

		pk := t.KeyGen(m[tssTypes.AttributeKeyKeyID]) // simulate correct keygen + vote
		err := n.Broadcaster.Broadcast(n.Node.Ctx,
			&tssTypes.MsgVotePubKey{
				Sender:      n.Broadcaster.GetProxy(n.Node.Ctx, n.Broadcaster.LocalPrincipal).String(),
				PubKeyBytes: pk,
				PollMeta:    voting.NewPollMeta(tssTypes.ModuleName, m[tssTypes.AttributeKeyKeyID])})
		if err != nil {
			panic(err)
		}

		return true
	})

	// register listener for sign start
	n.Node.RegisterEventListener(func(event abci.Event) bool {
		if event.Type != tssTypes.EventTypeSign {
			return false
		}

		m := mapifyAttributes(event)
		if m[sdk.AttributeKeyAction] != tssTypes.AttributeValueStart {
			return false
		}
		if m[tssTypes.AttributeKeySigID] == "" {
			return false
		}

		sig := t.Sign(m[tssTypes.AttributeKeySigID], m[tssTypes.AttributeKeyKeyID], []byte(m[tssTypes.AttributeKeyPayload]))

		err := n.Broadcaster.Broadcast(n.Node.Ctx,
			&tssTypes.MsgVoteSig{
				Sender:   n.Broadcaster.GetProxy(n.Node.Ctx, n.Broadcaster.LocalPrincipal).String(),
				SigBytes: sig,
				PollMeta: voting.NewPollMeta(
					tssTypes.ModuleName,
					m[tssTypes.AttributeKeySigID],
				)})
		if err != nil {
			panic(err)
		}

		return true
	})
}

type listeners struct {
	keygenDone     <-chan abci.Event
	signDone       <-chan abci.Event
	btcDone        <-chan abci.Event
	ethDepositDone <-chan abci.Event
	ethTokenDone   <-chan abci.Event
}

func registerWaitEventListeners(n nodeData) listeners {
	// register listener for keygen completion
	keygenDone := n.Node.RegisterEventListener(func(event abci.Event) bool {
		return event.Type == tssTypes.EventTypePubKeyDecided
	})

	// register btc listener for outpoint confirmation
	btcConfirmationDone := n.Node.RegisterEventListener(func(event abci.Event) bool {
		attributes := mapifyAttributes(event)
		return event.Type == btcTypes.EventTypeOutpointConfirmation &&
			(attributes[sdk.AttributeKeyAction] == btcTypes.AttributeValueConfirm ||
				attributes[sdk.AttributeKeyAction] == btcTypes.AttributeValueReject)
	})

	// register eth listener for confirmation
	ethDepositDone := n.Node.RegisterEventListener(func(event abci.Event) bool {
		attributes := mapifyAttributes(event)
		return event.Type == ethTypes.EventTypeDepositConfirmation &&
			(attributes[sdk.AttributeKeyAction] == ethTypes.AttributeValueConfirm ||
				attributes[sdk.AttributeKeyAction] == ethTypes.AttributeValueReject)
	})

	// register eth listener for confirmation
	ethTokenDone := n.Node.RegisterEventListener(func(event abci.Event) bool {
		attributes := mapifyAttributes(event)
		return event.Type == ethTypes.EventTypeTokenConfirmation &&
			(attributes[sdk.AttributeKeyAction] == ethTypes.AttributeValueConfirm ||
				attributes[sdk.AttributeKeyAction] == ethTypes.AttributeValueReject)
	})

	// register listener for sign completion
	signDone := n.Node.RegisterEventListener(func(event abci.Event) bool {
		return event.Type == tssTypes.EventTypeSigDecided
	})

	return listeners{
		keygenDone:     keygenDone,
		signDone:       signDone,
		btcDone:        btcConfirmationDone,
		ethDepositDone: ethDepositDone,
		ethTokenDone:   ethTokenDone,
	}
}

func waitFor(eventDone <-chan abci.Event, repeats int) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Duration(repeats)*2*time.Second)
	defer cancel()
	for i := 0; i < repeats; i++ {
		select {
		case <-eventDone:
			break
		case <-timeout.Done():
			return fmt.Errorf("timeout at %d of %d", i, repeats-1)
		}
	}
	return nil
}

func mapifyAttributes(event abci.Event) map[string]string {
	m := map[string]string{}
	for _, attribute := range sdk.StringifyEvent(event).Attributes {
		m[attribute.Key] = attribute.Value
	}
	return m
}

func createTokenDeployLogs(gateway, addr common.Address) []*goEthTypes.Log {
	numLogs := rand.I64Between(1, 100)
	pos := rand.I64Between(0, numLogs)
	var logs []*goEthTypes.Log

	for i := int64(0); i < numLogs; i++ {
		stringType, err := abi.NewType("string", "string", nil)
		if err != nil {
			panic(err)
		}
		addressType, err := abi.NewType("address", "address", nil)
		if err != nil {
			panic(err)
		}
		args := abi.Arguments{{Type: stringType}, {Type: addressType}}

		if i == pos {
			data, err := args.Pack("satoshi", addr)
			if err != nil {
				panic(err)
			}
			logs = append(logs, &goEthTypes.Log{Address: gateway, Data: data, Topics: []common.Hash{eth.ERC20TokenDeploySig}})
			continue
		}

		randDenom := rand.Str(4)
		randGateway := common.BytesToAddress(rand.Bytes(common.AddressLength))
		randAddr := common.BytesToAddress(rand.Bytes(common.AddressLength))
		randData, err := args.Pack(randDenom, randAddr)
		randTopic := common.BytesToHash(rand.Bytes(common.HashLength))
		if err != nil {
			panic(err)
		}
		logs = append(logs, &goEthTypes.Log{Address: randGateway, Data: randData, Topics: []common.Hash{randTopic}})
	}

	return logs
}
