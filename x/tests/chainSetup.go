package tests

import (
	"context"
	"fmt"
	mathRand "math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
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

	"github.com/axelarnetwork/axelar-core/app"
	eth2 "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmMock "github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	snapshotExportedMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	snapshotKeeper "github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	snapshotTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	snapshotTypesMock "github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
	"github.com/axelarnetwork/axelar-core/x/tss"
	tssExported "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssKeeper "github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	tssMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	"github.com/axelarnetwork/axelar-core/x/vote"
	voteKeeper "github.com/axelarnetwork/axelar-core/x/vote/keeper"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

func randomSender() sdk.AccAddress {
	return rand.AccAddr()
}

type testMocks struct {
	ETH     *evmMock.RPCClientMock
	BTC     *btcMock.RPCClientMock
	Keygen  *tssMock.TofndKeyGenClientMock
	Sign    *tssMock.TofndSignClientMock
	Staker  *snapshotTypesMock.StakingKeeperMock
	Tofnd   *tssMock.TofndClientMock
	Slasher *snapshotExportedMock.SlasherMock
	Tss     *snapshotExportedMock.TssMock
}

type nodeData struct {
	Node      *fake.Node
	Validator stakingtypes.Validator
	Proxy     sdk.AccAddress
	Mocks     testMocks
}

func newNode(moniker string, mocks testMocks) *fake.Node {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger().With("node", moniker))
	encCfg := app.MakeEncodingConfig()

	snapSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snap")
	snapKeeper := snapshotKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(snapshotTypes.StoreKey), snapSubspace, mocks.Staker, mocks.Slasher, mocks.Tss)
	snapKeeper.SetParams(ctx, snapshotTypes.DefaultParams())
	voter := voteKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(voteTypes.StoreKey), snapKeeper)

	btcSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	bitcoinKeeper := btcKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(btcTypes.StoreKey), btcSubspace)
	btcParams := btcTypes.DefaultParams()
	bitcoinKeeper.SetParams(ctx, btcParams)

	paramsK := paramsKeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"))
	EVMKeeper := evmKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(evmTypes.StoreKey), paramsK)
	EVMKeeper.SetParams(ctx, evmTypes.DefaultParams()...)

	tssSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), tssTypes.DefaultParamspace)
	signer := tssKeeper.NewKeeper(encCfg.Amino, sdk.NewKVStoreKey(tssTypes.StoreKey), tssSubspace, mocks.Slasher)

	// set the acknowledgment window just enough for all nodes to be able to submit their acks in time
	tssParams := tssTypes.DefaultParams()
	signer.SetParams(ctx, tssParams)

	nexusSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("balanceKey"), sdk.NewKVStoreKey("tbalanceKey"), "balance")
	nexusK := nexusKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(nexusTypes.StoreKey), nexusSubspace)
	nexusK.SetParams(ctx, nexusTypes.DefaultParams())

	voter.SetDefaultVotingThreshold(ctx, voteTypes.DefaultGenesisState().VotingThreshold)

	router := fake.NewRouter()

	snapshotHandler := snapshot.NewHandler(snapKeeper)
	btcHandler := bitcoin.NewHandler(bitcoinKeeper, voter, signer, nexusK, snapKeeper)
	ethHandler := evm.NewHandler(EVMKeeper, mocks.Tss, voter, signer, nexusK, snapKeeper)
	tssHandler := tss.NewHandler(signer, snapKeeper, nexusK, voter, &tssMock.StakingKeeperMock{
		GetLastTotalPowerFunc: mocks.Staker.GetLastTotalPowerFunc,
	})

	router = router.
		AddRoute(sdk.NewRoute(snapshotTypes.RouterKey, snapshotHandler)).
		AddRoute(sdk.NewRoute(btcTypes.RouterKey, btcHandler)).
		AddRoute(sdk.NewRoute(evmTypes.RouterKey, ethHandler)).
		AddRoute(sdk.NewRoute(tssTypes.RouterKey, tssHandler))

	evmMap := make(map[string]evmTypes.RPCClient)
	evmMap["ethereum"] = mocks.ETH

	queriers := map[string]sdk.Querier{
		btcTypes.QuerierRoute: btcKeeper.NewQuerier(mocks.BTC, bitcoinKeeper, signer, nexusK),
		evmTypes.QuerierRoute: evmKeeper.NewQuerier(evmMap, EVMKeeper, signer, nexusK),
	}

	node := fake.NewNode(moniker, ctx, router, queriers).
		WithEndBlockers(
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return vote.EndBlocker(ctx, req, voter)
			},
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return bitcoin.EndBlocker(ctx, req, bitcoinKeeper, signer, voter, snapKeeper)
			},
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return evm.EndBlocker(ctx, req, EVMKeeper, nexusK, signer)
			},
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return tss.EndBlocker(ctx, req, signer, voter, snapKeeper)
			},
		)
	return node
}

func createMocks(validators []stakingtypes.Validator) testMocks {
	slasher := &snapshotExportedMock.SlasherMock{
		GetValidatorSigningInfoFunc: func(ctx sdk.Context, address sdk.ConsAddress) (slashingtypes.ValidatorSigningInfo, bool) {
			newInfo := slashingtypes.NewValidatorSigningInfo(
				address,
				int64(0),        // height at which validator was first a candidate OR was unjailed
				int64(3),        // index offset into signed block bit array. TODO: check if needs to be set correctly.
				time.Unix(0, 0), // jailed until
				false,           // tomstoned
				int64(0),        // missed blocks
			)
			return newInfo, true
		},
		SignedBlocksWindowFunc: func(sdk.Context) int64 { return 100 },
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
		ValidatorFunc: func(_ sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI {
			addrStr := addr.String()
			for _, val := range validators {
				if val.OperatorAddress == addrStr {
					return val
				}
			}
			return nil
		},
	}

	tssK := &snapshotExportedMock.TssMock{
		GetMaxMissedBlocksPerWindowFunc: func(sdk.Context) utils.Threshold {
			return tssTypes.DefaultParams().MaxMissedBlocksPerWindow
		},
		GetTssSuspendedUntilFunc: func(sdk.Context, sdk.ValAddress) int64 { return 0 },
		OperatorIsAvailableForCounterFunc: func(_ sdk.Context, _ int64, v sdk.ValAddress) bool {

			// we cannot evaluate the counter number, but for the context of the unit tests,
			// we can assume the validators always send their acknowledgments
			for _, validator := range validators {
				if validator.GetOperator().String() == v.String() {
					return true
				}
			}
			return false
		},
	}

	ethClient := &evmMock.RPCClientMock{
		SendAndSignTransactionFunc: func(context.Context, geth.CallMsg) (string, error) {
			return "", nil
		},
		PendingNonceAtFunc: func(context.Context, common.Address) (uint64, error) {
			return mathRand.Uint64(), nil
		},
		SendTransactionFunc: func(context.Context, *gethTypes.Transaction) error { return nil },
	}

	btcClient := &btcMock.RPCClientMock{}

	return testMocks{
		BTC:     btcClient,
		ETH:     ethClient,
		Staker:  stakingKeeper,
		Slasher: slasher,
		Tss:     tssK,
	}
}

// initChain Creates a chain with given number of validators
func initChain(nodeCount int, test string) (*fake.BlockChain, []nodeData) {
	protoPK, err := cryptocodec.FromTmPubKeyInterface(ed25519.GenPrivKey().PubKey())
	if err != nil {
		panic(err)
	}
	consPK, err := codectypes.NewAnyWithValue(protoPK)
	if err != nil {
		panic(err)
	}

	// the recovery infos we store on chain must match the number of shares. However, currently we
	// do not have a way to properly mock the number of recovery infos in a way that matches the
	// number of shares that we assign during initialization. To circuvent this issue, we force
	// validators to always have one single key share (even when distributing them by stake), and
	// we mock a single recovery info. This way we force the amount of recovery infos to match
	// the number of shares.

	// TODO: find a way to correctly mock the amount of recovery infos from the number of shares
	// held by any  given validators,  even if we assign an arbitrary number of tokens to each
	tokens := sdk.TokensFromConsensusPower(rand.I64Between(100, 1000))
	var validators []stakingtypes.Validator
	for i := 0; i < nodeCount; i++ {
		// assign validators
		validator := stakingtypes.Validator{
			OperatorAddress: rand.ValAddr().String(),
			Tokens:          tokens,
			Status:          stakingtypes.Bonded,
			ConsensusPubkey: consPK,
		}
		validators = append(validators, validator)
	}
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Tokens.GT(validators[j].Tokens)
	})
	// create a chain
	chain := fake.NewBlockchain().WithBlockSize(50 * nodeCount).WithBlockTimeOut(20 * time.Millisecond)

	t := fake.NewTofnd()
	var data []nodeData
	for i, validator := range validators {
		// create mocks
		mocks := createMocks(validators)

		node := newNode(test+strconv.Itoa(i), mocks)
		chain.AddNodes(node)
		n := nodeData{Node: node, Validator: validator, Mocks: mocks, Proxy: rand.AccAddr()}

		registerTSSEventListeners(n, t, chain.Submit)
		registerBTCEventListener(n, chain.Submit)
		registerETHEventListener(n, chain.Submit)
		data = append(data, n)
	}

	// start chain
	chain.Start()

	return chain, data
}

func registerBTCEventListener(n nodeData, submitMsg func(msg sdk.Msg) (result <-chan *fake.Result)) {
	encCfg := app.MakeEncodingConfig()

	// register listener for confirmation
	n.Node.RegisterEventListener(func(event abci.Event) bool {
		if event.Type != btcTypes.EventTypeOutpointConfirmation {
			return false
		}

		m := mapifyAttributes(event)
		if m[sdk.AttributeKeyAction] != btcTypes.AttributeValueStart {
			return false
		}

		var pollKey voting.PollKey
		encCfg.Amino.MustUnmarshalJSON([]byte(m[btcTypes.AttributeKeyPoll]), &pollKey)

		var out btcTypes.OutPointInfo
		encCfg.Amino.MustUnmarshalJSON([]byte(m[btcTypes.AttributeKeyOutPointInfo]), &out)
		_ = submitMsg(btcTypes.NewVoteConfirmOutpointRequest(n.Proxy, pollKey, out.GetOutPoint(), true))

		return true
	})
}

func registerETHEventListener(n nodeData, submitMsg func(msg sdk.Msg) (result <-chan *fake.Result)) {
	encCfg := app.MakeEncodingConfig()
	// register listener for deposit confirmation
	n.Node.RegisterEventListener(func(event abci.Event) bool {
		if event.Type != evmTypes.EventTypeDepositConfirmation {
			return false
		}

		m := mapifyAttributes(event)
		if m[sdk.AttributeKeyAction] != evmTypes.AttributeValueStart {
			return false
		}

		var pollKey voting.PollKey
		encCfg.Amino.MustUnmarshalJSON([]byte(m[evmTypes.AttributeKeyPoll]), &pollKey)

		_ = submitMsg(&evmTypes.VoteConfirmDepositRequest{
			Sender:      n.Proxy,
			Chain:       m[evmTypes.AttributeKeyChain],
			PollKey:     pollKey,
			Confirmed:   true,
			TxID:        types.Hash(common.HexToHash(m[evmTypes.AttributeKeyTxID])),
			BurnAddress: types.Address(common.HexToAddress(m[evmTypes.AttributeKeyBurnAddress])),
		})

		return true
	})

	// register listener for token deploy confirmation
	n.Node.RegisterEventListener(func(event abci.Event) bool {
		if event.Type != evmTypes.EventTypeTokenConfirmation {
			return false
		}

		m := mapifyAttributes(event)
		if m[sdk.AttributeKeyAction] != evmTypes.AttributeValueStart {
			return false
		}

		var pollKey voting.PollKey
		encCfg.Amino.MustUnmarshalJSON([]byte(m[evmTypes.AttributeKeyPoll]), &pollKey)

		_ = submitMsg(
			&evmTypes.VoteConfirmTokenRequest{
				Sender:    n.Proxy,
				Chain:     m[evmTypes.AttributeKeyChain],
				PollKey:   pollKey,
				Confirmed: true,
				TxID:      evmTypes.Hash(common.HexToHash(m[evmTypes.AttributeKeyTxID])),
				Asset:     m[evmTypes.AttributeKeyAsset],
			})

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
		Amount:   btcutil.Amount(rand.I64Between(1, 10000000)),
		Address:  recipient,
	}
}

func registerTSSEventListeners(n nodeData, t *fake.Tofnd, submitMsg func(msg sdk.Msg) (result <-chan *fake.Result)) {
	// register listener for tofnd acknowledgment
	n.Node.RegisterEventListener(func(event abci.Event) bool {
		if event.Type != tssTypes.EventTypeAck {
			return false
		}

		m := mapifyAttributes(event)
		var ackType tssExported.AckType
		var ID string

		switch m[sdk.AttributeKeyAction] {
		case tssTypes.AttributeValueKeygen:
			ackType = tssExported.AckType_Keygen
			ID = m[tssTypes.AttributeKeyKeyID]
		case tssTypes.AttributeValueSign:
			ackType = tssExported.AckType_Sign
			ID = m[tssTypes.AttributeKeySigID]
		default:
			return false
		}

		if m[tssTypes.AttributeKeyKeyID] == "" {
			return false
		}

		height, err := strconv.ParseInt(m[tssTypes.AttributeKeyHeight], 10, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot convert string to int64: %s", err.Error()))
		}

		_ = submitMsg(&tssTypes.AckRequest{
			Sender:  n.Proxy,
			ID:      ID,
			AckType: ackType,
			Height:  height})

		return true
	})

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

		// the recovery infos we store on chain must match the number of shares. However, currently we
		// do not have a way to properly mock the number of recovery infos in a way that matches the
		// number of shares that we assign during initialization. To circuvent this issue, we force
		// validators to always have one single key share (even when distributing them by stake), and
		// we mock a single recovery info. This way we force the amount of recovery infos to match
		// the number of shares.

		// TODO: find a way to correctly mock the amount of group infos and party infos from the number of shares
		// held by any  given validators,  even if we assign an arbitrary number of tokens to each
		groupInfo := []byte{1}
		partyInfo := [][]byte{{1}}
		result := &tofnd.MessageOut_KeygenResult{
			KeygenResultData: &tofnd.MessageOut_KeygenResult_Data{
				Data: &tofnd.KeygenOutput{
					PubKey: pk, GroupInfo: groupInfo, RecoveryInfo: partyInfo,
				},
			},
		}

		_ = submitMsg(&tssTypes.VotePubKeyRequest{
			Sender:  n.Proxy,
			Result:  result,
			PollKey: voting.NewPollKey(tssTypes.ModuleName, m[tssTypes.AttributeKeyKeyID])})

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

		_ = submitMsg(&tssTypes.VoteSigRequest{
			Sender: n.Proxy,
			Result: &tofnd.MessageOut_SignResult{SignResultData: &tofnd.MessageOut_SignResult_Signature{Signature: sig}},
			PollKey: voting.NewPollKey(
				tssTypes.ModuleName,
				m[tssTypes.AttributeKeySigID],
			)})
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
		attributes := mapifyAttributes(event)
		return event.Type == tssTypes.EventTypeKeygen &&
			attributes[sdk.AttributeKeyAction] == tssTypes.AttributeValueDecided
	})

	// register btc listener for outpoint confirmation
	btcConfirmationDone := n.Node.RegisterEventListener(func(event abci.Event) bool {
		attributes := mapifyAttributes(event)
		return event.Type == btcTypes.EventTypeOutpointConfirmation &&
			(attributes[sdk.AttributeKeyAction] == btcTypes.AttributeValueConfirm ||
				attributes[sdk.AttributeKeyAction] == btcTypes.AttributeValueReject)
	})

	// register evm listener for confirmation
	ethDepositDone := n.Node.RegisterEventListener(func(event abci.Event) bool {
		attributes := mapifyAttributes(event)
		return event.Type == evmTypes.EventTypeDepositConfirmation &&
			(attributes[sdk.AttributeKeyAction] == evmTypes.AttributeValueConfirm ||
				attributes[sdk.AttributeKeyAction] == evmTypes.AttributeValueReject)
	})

	// register evm listener for confirmation
	ethTokenDone := n.Node.RegisterEventListener(func(event abci.Event) bool {
		attributes := mapifyAttributes(event)
		return event.Type == evmTypes.EventTypeTokenConfirmation &&
			(attributes[sdk.AttributeKeyAction] == evmTypes.AttributeValueConfirm ||
				attributes[sdk.AttributeKeyAction] == evmTypes.AttributeValueReject)
	})

	// register listener for sign completion
	signDone := n.Node.RegisterEventListener(func(event abci.Event) bool {
		attributes := mapifyAttributes(event)
		return event.Type == tssTypes.EventTypeSign &&
			attributes[sdk.AttributeKeyAction] == tssTypes.AttributeValueDecided
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
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
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
			logs = append(logs, &goEthTypes.Log{Address: gateway, Data: data, Topics: []common.Hash{eth2.ERC20TokenDeploymentSig}})
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
