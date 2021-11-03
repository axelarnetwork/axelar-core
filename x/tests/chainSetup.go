package tests

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/ethereum/go-ethereum/common"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm"
	"github.com/axelarnetwork/axelar-core/x/nexus"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	rewardKeeper "github.com/axelarnetwork/axelar-core/x/reward/keeper"
	rewardTypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	rewardMock "github.com/axelarnetwork/axelar-core/x/reward/types/mock"
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
	Keygen      *tssMock.TofndKeyGenClientMock
	Sign        *tssMock.TofndSignClientMock
	Staker      *snapshotTypesMock.StakingKeeperMock
	Tofnd       *tssMock.TofndClientMock
	Slasher     *snapshotExportedMock.SlasherMock
	Tss         *snapshotExportedMock.TssMock
	Banker      *rewardMock.BankerMock
	Distributor *rewardMock.DistributorMock
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

	rewardSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "reward")
	rewardKeeper := rewardKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(rewardTypes.StoreKey), rewardSubspace, mocks.Banker, mocks.Distributor, mocks.Staker)

	snapSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snap")
	snapKeeper := snapshotKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(snapshotTypes.StoreKey), snapSubspace, mocks.Staker, mocks.Slasher, mocks.Tss)
	snapKeeper.SetParams(ctx, snapshotTypes.DefaultParams())
	voter := voteKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(voteTypes.StoreKey), snapKeeper, mocks.Staker, rewardKeeper)

	btcSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	bitcoinKeeper := btcKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(btcTypes.StoreKey), btcSubspace)
	btcParams := btcTypes.DefaultParams()
	bitcoinKeeper.SetParams(ctx, btcParams)

	paramsK := paramsKeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"))
	EVMKeeper := evmKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(evmTypes.StoreKey), paramsK)
	evmParams := evmTypes.DefaultParams()
	for i := range evmParams {
		evmParams[i].VotingThreshold = utils.OneThreshold
	}
	EVMKeeper.SetParams(ctx, evmParams...)

	tssSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), tssTypes.DefaultParamspace)
	signer := tssKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(tssTypes.StoreKey), tssSubspace, mocks.Slasher, rewardKeeper)

	signer.SetParams(ctx, tssTypes.DefaultParams())

	nexusSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("balanceKey"), sdk.NewKVStoreKey("tbalanceKey"), "balance")
	nexusK := nexusKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(nexusTypes.StoreKey), nexusSubspace)
	nexusK.SetParams(ctx, nexusTypes.DefaultParams())

	voter.SetDefaultVotingThreshold(ctx, voteTypes.DefaultGenesisState().VotingThreshold)

	tssRouter := tssTypes.NewRouter()
	tssRouter = tssRouter.AddRoute(evmTypes.ModuleName, evmKeeper.NewTssHandler(EVMKeeper, nexusK, signer)).
		AddRoute(btcTypes.ModuleName, btcKeeper.NewTssHandler(bitcoinKeeper, signer))
	signer.SetRouter(tssRouter)

	router := fake.NewRouter()

	snapshotHandler := snapshot.NewHandler(snapKeeper)
	btcHandler := bitcoin.NewHandler(bitcoinKeeper, voter, signer, nexusK, snapKeeper)
	ethHandler := evm.NewHandler(EVMKeeper, mocks.Tss, voter, signer, nexusK, snapKeeper)
	tssHandler := tss.NewHandler(signer, snapKeeper, nexusK, voter, &tssMock.StakingKeeperMock{
		GetLastTotalPowerFunc: mocks.Staker.GetLastTotalPowerFunc,
		ValidatorFunc:         mocks.Staker.ValidatorFunc,
	}, rewardKeeper)
	nexusHandler := nexus.NewHandler(nexusK, snapKeeper)

	router = router.
		AddRoute(sdk.NewRoute(snapshotTypes.RouterKey, snapshotHandler)).
		AddRoute(sdk.NewRoute(btcTypes.RouterKey, btcHandler)).
		AddRoute(sdk.NewRoute(evmTypes.RouterKey, ethHandler)).
		AddRoute(sdk.NewRoute(tssTypes.RouterKey, tssHandler)).
		AddRoute(sdk.NewRoute(nexusTypes.RouterKey, nexusHandler))

	queriers := map[string]sdk.Querier{
		btcTypes.QuerierRoute: btcKeeper.NewQuerier(bitcoinKeeper, signer, nexusK),
		evmTypes.QuerierRoute: evmKeeper.NewQuerier(EVMKeeper, signer, nexusK),
	}

	node := fake.NewNode(moniker, ctx, router, queriers).
		WithEndBlockers(
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return vote.EndBlocker(ctx, req, voter)
			},
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return bitcoin.EndBlocker(ctx, req, bitcoinKeeper, signer)
			},
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return evm.EndBlocker(ctx, req, EVMKeeper, nexusK, signer)
			},
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return tss.EndBlocker(ctx, req, signer, voter, nexusK, snapKeeper)
			},
			func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
				return nexus.EndBlocker(ctx, req, nexusK, mocks.Staker)
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
				totalPower = totalPower.AddRaw(val.ConsensusPower(sdk.DefaultPowerReduction))
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
		PowerReductionFunc: func(_ sdk.Context) sdk.Int {
			return sdk.DefaultPowerReduction
		},
	}

	tssK := &snapshotExportedMock.TssMock{
		GetMaxMissedBlocksPerWindowFunc: func(sdk.Context) utils.Threshold {
			return tssTypes.DefaultParams().MaxMissedBlocksPerWindow
		},
		GetTssSuspendedUntilFunc: func(sdk.Context, sdk.ValAddress) int64 { return 0 },
		IsOperatorAvailableFunc: func(_ sdk.Context, v sdk.ValAddress, keyIDs ...tssExported.KeyID) bool {
			for _, validator := range validators {
				if validator.GetOperator().String() == v.String() {
					return true
				}
			}
			return false
		},
	}

	return testMocks{
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

	// if we give different amounts of tokens to each validator, we
	// introduce a non-deterministic test failure
	// TODO: investigate why there is a non-deterministic test failure
	tokens := sdk.TokensFromConsensusPower(rand.I64Between(100, 1000), sdk.DefaultPowerReduction)
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
	// register listener for heartbeat event
	n.Node.RegisterEventListener(func(event abci.Event) bool {
		if event.Type != tssTypes.EventTypeHeartBeat {
			return false
		}

		m := mapifyAttributes(event)

		if m[sdk.AttributeKeyAction] != tssTypes.AttributeValueSend {
			return false
		}

		var keyInfos []tssTypes.KeyInfo
		_ = json.Unmarshal([]byte(m[tssTypes.AttributeKeyKeyInfos]), &keyInfos)

		var present []tssExported.KeyID
		for _, keyInfo := range keyInfos {
			if t.HasKey(string(keyInfo.KeyID)) {
				present = append(present, keyInfo.KeyID)
			}
		}

		_ = submitMsg(tssTypes.NewHeartBeatRequest(n.Proxy, present))

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
		groupRecoverInfo := []byte(tssTypes.EventTypeKeygen + tssTypes.AttributeValueStart + m[tssTypes.AttributeKeyKeyID])
		privateRecoverInfo := rand.BytesBetween(1, 100)
		result := &tofnd.MessageOut_KeygenResult{
			KeygenResultData: &tofnd.MessageOut_KeygenResult_Data{
				Data: &tofnd.KeygenOutput{
					PubKey:             pk,
					GroupRecoverInfo:   groupRecoverInfo,
					PrivateRecoverInfo: privateRecoverInfo,
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
	keygenDone        <-chan abci.Event
	signDone          <-chan abci.Event
	btcDone           <-chan abci.Event
	ethDepositDone    <-chan abci.Event
	ethTokenDone      <-chan abci.Event
	chainActivated    <-chan abci.Event
	ackRequested      <-chan abci.Event
	consolidationDone <-chan abci.Event
}

func registerWaitEventListeners(n nodeData) listeners {
	// register listener for keygen completion
	chainActivated := n.Node.RegisterEventListener(func(event abci.Event) bool {
		attributes := mapifyAttributes(event)
		return event.Type == nexusTypes.EventTypeChain &&
			attributes[sdk.AttributeKeyAction] == nexusTypes.AttributeValueActivated
	})

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

	// register listener for ack request
	ackRequested := n.Node.RegisterEventListener(func(event abci.Event) bool {
		attributes := mapifyAttributes(event)
		return event.Type == tssTypes.EventTypeHeartBeat &&
			attributes[sdk.AttributeKeyAction] == tssTypes.AttributeValueSend
	})

	// register listener for consolidation done
	condolidationDone := n.Node.RegisterEventListener(func(event abci.Event) bool {
		attributes := mapifyAttributes(event)
		return event.Type == btcTypes.EventTypeConsolidationTx &&
			attributes[sdk.AttributeKeyAction] == btcTypes.AttributeValueSigned
	})

	return listeners{
		keygenDone:        keygenDone,
		signDone:          signDone,
		btcDone:           btcConfirmationDone,
		ethDepositDone:    ethDepositDone,
		ethTokenDone:      ethTokenDone,
		chainActivated:    chainActivated,
		ackRequested:      ackRequested,
		consolidationDone: condolidationDone,
	}
}

func waitFor(eventDone <-chan abci.Event, repeats int) error {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	for i := 0; i < repeats; i++ {
		select {
		case <-eventDone:
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
