package app_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/btcsuite/btcd/btcec/v2"
	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	auxiliarytypes "github.com/axelarnetwork/axelar-core/x/auxiliary/types"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	multisigtypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	permissionexported "github.com/axelarnetwork/axelar-core/x/permission/exported"
	permissionkeeper "github.com/axelarnetwork/axelar-core/x/permission/keeper"
	permissiontypes "github.com/axelarnetwork/axelar-core/x/permission/types"
	rewardtypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	voteexported "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/funcs"
)

const (
	testChainID = "axelar-emergency-test"
	feeDenom    = "uaxl" // the only fee denom feepolicy allows by default
)

// testChain runs a single-validator AxelarApp in-process: InitChain, FinalizeBlock, Commit and CheckTx are driven
// directly through ABCI with real signed transactions, so the full ante handler and msg service router are exercised
type testChain struct {
	t      *testing.T
	app    *app.AxelarApp
	db     dbm.DB
	home   string
	encCfg params.EncodingConfig
	valSet *cmttypes.ValidatorSet
	time   time.Time

	useProposals bool

	validatorOwner *testAccount
	accessControl  *testAccount
	chainManager   *testAccount
	user           *testAccount
}

type testAccount struct {
	priv cryptotypes.PrivKey
	addr sdk.AccAddress
}

func newTestAccount() *testAccount {
	priv := secp256k1.GenPrivKey()
	return &testAccount{priv: priv, addr: sdk.AccAddress(priv.PubKey().Address())}
}

func newTestChain(t *testing.T) *testChain {
	app.WasmEnabled, app.IBCWasmHooksEnabled = "false", "false"

	c := &testChain{
		t:              t,
		db:             dbm.NewMemDB(),
		home:           t.TempDir(),
		encCfg:         app.MakeEncodingConfig(),
		time:           time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		validatorOwner: newTestAccount(),
		accessControl:  newTestAccount(),
		chainManager:   newTestAccount(),
		user:           newTestAccount(),
	}
	c.app = c.newApp(nil)

	privVal := mock.NewPV()
	pubKey := funcs.Must(privVal.GetPubKey())
	c.valSet = cmttypes.NewValidatorSet([]*cmttypes.Validator{cmttypes.NewValidator(pubKey, 1)})

	accounts := []*testAccount{c.validatorOwner, c.accessControl, c.chainManager, c.user}
	var genAccs []authtypes.GenesisAccount
	var balances []banktypes.Balance
	for i, acc := range accounts {
		genAccs = append(genAccs, authtypes.NewBaseAccount(acc.addr, acc.priv.PubKey(), uint64(i), 0))
		balances = append(balances, banktypes.Balance{
			Address: acc.addr.String(),
			Coins: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(1_000_000_000_000)),
				sdk.NewCoin(feeDenom, sdkmath.NewInt(1_000_000_000_000)),
			),
		})
	}

	genesis := app.GetModuleBasics().DefaultGenesis(c.encCfg.Codec)
	genesis = funcs.Must(simtestutil.GenesisStateWithValSet(c.encCfg.Codec, genesis, c.valSet, genAccs, balances...))
	genesis[permissiontypes.ModuleName] = c.encCfg.Codec.MustMarshalJSON(&permissiontypes.GenesisState{
		Params: permissiontypes.DefaultParams(),
		GovAccounts: []permissiontypes.GovAccount{
			permissiontypes.NewGovAccount(c.accessControl.addr, permissionexported.ROLE_ACCESS_CONTROL),
			permissiontypes.NewGovAccount(c.chainManager.addr, permissionexported.ROLE_CHAIN_MANAGEMENT),
		},
	})

	var govGenesis govv1.GenesisState
	c.encCfg.Codec.MustUnmarshalJSON(genesis[govtypes.ModuleName], &govGenesis)
	votingPeriod, expeditedVotingPeriod := 10*time.Second, 5*time.Second
	govGenesis.Params.VotingPeriod = &votingPeriod
	govGenesis.Params.ExpeditedVotingPeriod = &expeditedVotingPeriod
	genesis[govtypes.ModuleName] = c.encCfg.Codec.MustMarshalJSON(&govGenesis)

	consensusParams := *simtestutil.DefaultConsensusParams
	block := *consensusParams.Block
	block.MaxGas = -1
	consensusParams.Block = &block

	_, err := c.app.InitChain(&abci.RequestInitChain{
		ChainId:         testChainID,
		Time:            c.time,
		ConsensusParams: &consensusParams,
		AppStateBytes:   funcs.Must(json.Marshal(genesis)),
		InitialHeight:   1,
	})
	require.NoError(t, err)

	c.nextBlock()

	return c
}

func (c *testChain) newApp(opts simtestutil.AppOptionsMap) *app.AxelarApp {
	appOpts := simtestutil.AppOptionsMap{flags.FlagHome: c.home}
	for k, v := range opts {
		appOpts[k] = v
	}

	return app.NewAxelarApp(log.NewNopLogger(), c.db, nil, true, c.encCfg, appOpts, []wasm.Option{}, baseapp.SetChainID(testChainID))
}

// fork starts a second node on a copy of this node's committed state
func (c *testChain) fork(opts simtestutil.AppOptionsMap) *testChain {
	db := dbm.NewMemDB()
	it := funcs.Must(c.db.Iterator(nil, nil))
	for ; it.Valid(); it.Next() {
		require.NoError(c.t, db.Set(it.Key(), it.Value()))
	}
	require.NoError(c.t, it.Close())

	node := *c
	node.db, node.home = db, c.t.TempDir()
	node.app = node.newApp(opts)

	return &node
}

func (c *testChain) height() int64 {
	return c.app.LastBlockHeight()
}

// nextBlock finalizes and commits a block with the given txs and returns the tx results
func (c *testChain) nextBlock(txs ...[]byte) []*abci.ExecTxResult {
	if c.useProposals {
		txs = c.prepareProposal(txs...)
		require.Equal(c.t, abci.ResponseProcessProposal_ACCEPT, c.processProposal(txs))
	}

	res, err := c.finalizeBlock(txs...)
	require.NoError(c.t, err)

	_, err = c.app.Commit()
	require.NoError(c.t, err)

	return res.TxResults
}

func (c *testChain) prepareProposal(txs ...[]byte) [][]byte {
	res, err := c.app.PrepareProposal(&abci.RequestPrepareProposal{
		MaxTxBytes:      1 << 20,
		Txs:             txs,
		Height:          c.height() + 1,
		Time:            c.time.Add(time.Second),
		ProposerAddress: c.valSet.Proposer.Address,
	})
	require.NoError(c.t, err)

	return res.Txs
}

func (c *testChain) processProposal(txs [][]byte) abci.ResponseProcessProposal_ProposalStatus {
	res, err := c.app.ProcessProposal(&abci.RequestProcessProposal{
		Txs:             txs,
		Height:          c.height() + 1,
		Time:            c.time.Add(time.Second),
		ProposerAddress: c.valSet.Proposer.Address,
	})
	require.NoError(c.t, err)

	return res.Status
}

func (c *testChain) finalizeBlock(txs ...[]byte) (*abci.ResponseFinalizeBlock, error) {
	c.time = c.time.Add(time.Second)

	return c.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:          c.height() + 1,
		Time:            c.time,
		Txs:             txs,
		ProposerAddress: c.valSet.Proposer.Address,
	})
}

func (c *testChain) ctx() sdk.Context {
	return c.app.NewUncachedContext(false, cmtproto.Header{Height: c.height(), Time: c.time, ChainID: testChainID})
}

// tx signs a tx with the signer's committed account number and sequence
func (c *testChain) tx(signer *testAccount, msgs ...sdk.Msg) []byte {
	return c.txWithFee(signer, sdk.Coins{}, msgs...)
}

func (c *testChain) txWithFee(signer *testAccount, fee sdk.Coins, msgs ...sdk.Msg) []byte {
	acc := getKeeper[authkeeper.AccountKeeper](c).GetAccount(c.ctx(), signer.addr)
	require.NotNil(c.t, acc)

	tx := funcs.Must(simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		c.encCfg.TxConfig,
		msgs,
		fee,
		10_000_000,
		testChainID,
		[]uint64{acc.GetAccountNumber()},
		[]uint64{acc.GetSequence()},
		signer.priv,
	))

	return funcs.Must(c.encCfg.TxConfig.TxEncoder()(tx))
}

func (c *testChain) checkTx(tx []byte) *abci.ResponseCheckTx {
	return funcs.Must(c.app.CheckTx(&abci.RequestCheckTx{Tx: tx, Type: abci.CheckTxType_New}))
}

// deliver delivers a single tx in its own block and returns its result
func (c *testChain) deliver(signer *testAccount, msgs ...sdk.Msg) *abci.ExecTxResult {
	results := c.nextBlock(c.tx(signer, msgs...))
	// the tx comes after a pause skip marker, if there is one
	return results[len(results)-1]
}

func (c *testChain) pauseLevel() permissionexported.PauseLevel {
	return getKeeper[permissionkeeper.Keeper](c).GetPauseLevel(c.ctx())
}

func (c *testChain) balance(addr sdk.AccAddress) sdkmath.Int {
	return c.balanceOf(addr, sdk.DefaultBondDenom)
}

func (c *testChain) balanceOf(addr sdk.AccAddress, denom string) sdkmath.Int {
	return getKeeper[bankkeeper.BaseKeeper](c).GetBalance(c.ctx(), addr, denom).Amount
}

func (c *testChain) pause(level permissionexported.PauseLevel) {
	res := c.deliver(c.accessControl, permissiontypes.NewPauseRequest(c.accessControl.addr, level, "incident"))
	require.Equal(c.t, uint32(0), res.Code, res.Log)
	require.Equal(c.t, level, c.pauseLevel())
}

func (c *testChain) passProposal(msgs ...sdk.Msg) govv1.Proposal {
	proposalID := c.submitProposal(msgs...)

	for range 12 {
		c.nextBlock()
	}

	return c.proposal(proposalID)
}

// submitProposal submits a gov proposal and votes yes with the validator
func (c *testChain) submitProposal(msgs ...sdk.Msg) uint64 {
	govKeeper := getKeeper[govkeeper.Keeper](c)
	minDeposit := funcs.Must(govKeeper.Params.Get(c.ctx())).MinDeposit
	proposal := funcs.Must(govv1.NewMsgSubmitProposal(msgs, minDeposit, c.user.addr.String(), "", "emergency", "emergency", false))
	res := c.deliver(c.user, proposal)
	require.Equal(c.t, uint32(0), res.Code, res.Log)

	proposalID := funcs.Must(govKeeper.ProposalID.Peek(c.ctx())) - 1

	res = c.deliver(c.validatorOwner, govv1.NewMsgVote(c.validatorOwner.addr, proposalID, govv1.OptionYes, ""))
	require.Equal(c.t, uint32(0), res.Code, res.Log)

	return proposalID
}

func (c *testChain) proposal(id uint64) govv1.Proposal {
	return funcs.Must(getKeeper[govkeeper.Keeper](c).Proposals.Get(c.ctx(), id))
}

func getKeeper[T any](c *testChain) *T {
	return app.GetKeeper[T](c.app.Keepers)
}

func bankSend(from, to sdk.AccAddress) *banktypes.MsgSend {
	return banktypes.NewMsgSend(from, to, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))
}

func isChainPausedErr(codespace string, code uint32) bool {
	return codespace == permissiontypes.ErrChainPaused.Codespace() && code == permissiontypes.ErrChainPaused.ABCICode()
}

func assertRejectedAsPaused(t *testing.T, c *testChain, signer *testAccount, msgs ...sdk.Msg) {
	t.Helper()

	checkRes := c.checkTx(c.tx(signer, msgs...))
	assert.True(t, isChainPausedErr(checkRes.Codespace, checkRes.Code), "CheckTx: expected chain paused error, got %d %s: %s", checkRes.Code, checkRes.Codespace, checkRes.Log)

	res := c.deliver(signer, msgs...)
	assert.True(t, isChainPausedErr(res.Codespace, res.Code), "FinalizeBlock: expected chain paused error, got %d %s: %s", res.Code, res.Codespace, res.Log)
}

func assertNotRejectedAsPaused(t *testing.T, c *testChain, signer *testAccount, msgs ...sdk.Msg) *abci.ExecTxResult {
	t.Helper()

	checkRes := c.checkTx(c.tx(signer, msgs...))
	assert.False(t, isChainPausedErr(checkRes.Codespace, checkRes.Code), "CheckTx: unexpected chain paused error: %s", checkRes.Log)
	assert.NotContains(t, checkRes.Log, "chain is paused")

	res := c.deliver(signer, msgs...)
	assert.False(t, isChainPausedErr(res.Codespace, res.Code), "FinalizeBlock: unexpected chain paused error: %s", res.Log)
	assert.NotContains(t, res.Log, "chain is paused")

	return res
}

func TestEmergencyPause_OnlyAccessControlCanPauseAndUnpause(t *testing.T) {
	c := newTestChain(t)

	for _, signer := range []*testAccount{c.chainManager, c.user, c.validatorOwner} {
		res := c.deliver(signer, permissiontypes.NewPauseRequest(signer.addr, permissionexported.PAUSE_LEVEL_ALL, "incident"))
		assert.NotEqual(t, uint32(0), res.Code)
		assert.Contains(t, res.Log, "is not authorized")
		assert.False(t, c.pauseLevel().IsPaused())
	}

	c.pause(permissionexported.PAUSE_LEVEL_ALL)

	for _, signer := range []*testAccount{c.chainManager, c.user, c.validatorOwner} {
		res := c.deliver(signer, permissiontypes.NewUnpauseRequest(signer.addr, "resolved"))
		assert.NotEqual(t, uint32(0), res.Code)
		assert.Contains(t, res.Log, "is not authorized")
		assert.True(t, c.pauseLevel().IsPaused())
	}

	res := c.deliver(c.accessControl, permissiontypes.NewUnpauseRequest(c.accessControl.addr, "resolved"))
	assert.Equal(t, uint32(0), res.Code, res.Log)
	assert.False(t, c.pauseLevel().IsPaused())
}

func TestEmergencyPause_UserTransferFailsWhilePausedAndSucceedsAfterUnpause(t *testing.T) {
	c := newTestChain(t)
	recipient := newTestAccount().addr

	res := c.deliver(c.user, bankSend(c.user.addr, recipient))
	require.Equal(t, uint32(0), res.Code, res.Log)
	require.Equal(t, int64(1), c.balance(recipient).Int64())

	for _, level := range []permissionexported.PauseLevel{permissionexported.PAUSE_LEVEL_ALL, permissionexported.PAUSE_LEVEL_EXCEPT_VALIDATORS} {
		before := c.balance(recipient)

		c.pause(level)
		assertRejectedAsPaused(t, c, c.user, bankSend(c.user.addr, recipient))
		// privileged signers are paused too
		assertRejectedAsPaused(t, c, c.accessControl, bankSend(c.accessControl.addr, recipient))
		assert.Equal(t, before, c.balance(recipient))

		res = c.deliver(c.accessControl, permissiontypes.NewUnpauseRequest(c.accessControl.addr, "resolved"))
		require.Equal(t, uint32(0), res.Code, res.Log)

		res = c.deliver(c.user, bankSend(c.user.addr, recipient))
		require.Equal(t, uint32(0), res.Code, res.Log)
	}
	assert.Equal(t, int64(3), c.balance(recipient).Int64())
}

func TestEmergencyPause_AllowlistedMessagesStillPass(t *testing.T) {
	t.Run("level all: changing the level, halting and unpausing", func(t *testing.T) {
		c := newTestChain(t)
		c.pause(permissionexported.PAUSE_LEVEL_ALL)

		res := assertNotRejectedAsPaused(t, c, c.accessControl, permissiontypes.NewPauseRequest(c.accessControl.addr, permissionexported.PAUSE_LEVEL_EXCEPT_VALIDATORS, "loosen"))
		assert.Equal(t, uint32(0), res.Code, res.Log)
		assert.Equal(t, permissionexported.PAUSE_LEVEL_EXCEPT_VALIDATORS, c.pauseLevel())

		res = assertNotRejectedAsPaused(t, c, c.accessControl, permissiontypes.NewPauseRequest(c.accessControl.addr, permissionexported.PAUSE_LEVEL_ALL, "tighten"))
		assert.Equal(t, uint32(0), res.Code, res.Log)
		assert.Equal(t, permissionexported.PAUSE_LEVEL_ALL, c.pauseLevel())

		res = assertNotRejectedAsPaused(t, c, c.accessControl, permissiontypes.NewUnpauseRequest(c.accessControl.addr, "resolved"))
		assert.Equal(t, uint32(0), res.Code, res.Log)
		assert.False(t, c.pauseLevel().IsPaused())

		c.pause(permissionexported.PAUSE_LEVEL_ALL)
		res = c.deliver(c.accessControl, permissiontypes.NewHaltChainRequest(c.accessControl.addr, "escalate", false))
		assert.Equal(t, uint32(0), res.Code, res.Log)
	})

	t.Run("level all: validator and key rotation messages are rejected", func(t *testing.T) {
		c := newTestChain(t)
		c.pause(permissionexported.PAUSE_LEVEL_ALL)

		assertRejectedAsPaused(t, c, c.user, rewardtypes.NewRefundMsgRequest(c.user.addr, submitSignature(c.user.addr)))
		assertRejectedAsPaused(t, c, c.user, valdVote(c.user.addr))
		assertRejectedAsPaused(t, c, c.accessControl, permissiontypes.NewRegisterControllerRequest(c.accessControl.addr, newTestAccount().addr))
	})

	t.Run("level except validators: liveness and key rotation messages", func(t *testing.T) {
		c := newTestChain(t)
		c.pause(permissionexported.PAUSE_LEVEL_EXCEPT_VALIDATORS)

		// these fail later for unrelated reasons (no signing session), but must not be stopped by the pause
		assertNotRejectedAsPaused(t, c, c.user, rewardtypes.NewRefundMsgRequest(c.user.addr, submitSignature(c.user.addr)))
		assertNotRejectedAsPaused(t, c, c.user, auxiliarytypes.NewBatchRequest(c.user.addr, []sdk.Msg{
			rewardtypes.NewRefundMsgRequest(c.user.addr, submitSignature(c.user.addr)),
		}))
		assertNotRejectedAsPaused(t, c, c.user, valdVote(c.user.addr))

		controller := newTestAccount().addr
		res := assertNotRejectedAsPaused(t, c, c.accessControl, permissiontypes.NewRegisterControllerRequest(c.accessControl.addr, controller))
		assert.Equal(t, uint32(0), res.Code, res.Log)
		res = assertNotRejectedAsPaused(t, c, c.accessControl, permissiontypes.NewDeregisterControllerRequest(c.accessControl.addr, controller))
		assert.Equal(t, uint32(0), res.Code, res.Log)
	})
}

func TestEmergencyPause_NestedMessagesAreRejected(t *testing.T) {
	t.Run("authz MsgExec", func(t *testing.T) {
		c := newTestChain(t)
		recipient := newTestAccount().addr

		// the user grants the chain manager the right to send on its behalf before the pause
		grant := funcs.Must(authz.NewMsgGrant(c.user.addr, c.chainManager.addr, authz.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})), nil))
		res := c.deliver(c.user, grant)
		require.Equal(t, uint32(0), res.Code, res.Log)

		exec := authz.NewMsgExec(c.chainManager.addr, []sdk.Msg{bankSend(c.user.addr, recipient)})
		for _, level := range []permissionexported.PauseLevel{permissionexported.PAUSE_LEVEL_ALL, permissionexported.PAUSE_LEVEL_EXCEPT_VALIDATORS} {
			c.pause(level)
			assertRejectedAsPaused(t, c, c.chainManager, &exec)

			// bypass the ante handler and execute through the router like a nested message would be
			_, err := c.app.MsgServiceRouter().Handler(&exec)(c.ctx(), &exec)
			assert.ErrorIs(t, err, permissiontypes.ErrChainPaused)
		}
		assert.True(t, c.balance(recipient).IsZero())

		res = c.deliver(c.accessControl, permissiontypes.NewUnpauseRequest(c.accessControl.addr, "resolved"))
		require.Equal(t, uint32(0), res.Code, res.Log)
		res = c.deliver(c.chainManager, &exec)
		require.Equal(t, uint32(0), res.Code, res.Log)
		assert.Equal(t, int64(1), c.balance(recipient).Int64())
	})

	t.Run("BatchRequest", func(t *testing.T) {
		c := newTestChain(t)
		recipient := newTestAccount().addr
		batch := auxiliarytypes.NewBatchRequest(c.user.addr, []sdk.Msg{bankSend(c.user.addr, recipient)})

		c.pause(permissionexported.PAUSE_LEVEL_ALL)
		assertRejectedAsPaused(t, c, c.user, batch)

		// the batch itself is allowed at this level, but its inner message is not
		c.pause(permissionexported.PAUSE_LEVEL_EXCEPT_VALIDATORS)
		assertRejectedAsPaused(t, c, c.user, batch)

		// bypass the ante handler: the circuit breaker rejects the inner message, the batch records the failure
		ctx := c.ctx()
		res, err := c.app.MsgServiceRouter().Handler(batch)(ctx, batch)
		require.NoError(t, err)
		var batchRes auxiliarytypes.BatchResponse
		require.NoError(t, c.encCfg.Codec.Unmarshal(res.MsgResponses[0].Value, &batchRes))
		require.Len(t, batchRes.Responses, 1)
		assert.Contains(t, batchRes.Responses[0].GetErr(), "chain is paused")
		assert.True(t, c.balance(recipient).IsZero())
	})

	t.Run("RefundMsgRequest", func(t *testing.T) {
		c := newTestChain(t)
		refund := rewardtypes.NewRefundMsgRequest(c.user.addr, submitSignature(c.user.addr))

		c.pause(permissionexported.PAUSE_LEVEL_ALL)
		assertRejectedAsPaused(t, c, c.user, refund)

		_, err := c.app.MsgServiceRouter().Handler(refund)(c.ctx(), refund)
		assert.ErrorIs(t, err, permissiontypes.ErrChainPaused)
	})

	t.Run("gov proposal execution", func(t *testing.T) {
		for _, paused := range []bool{false, true} {
			t.Run(fmt.Sprintf("paused %t", paused), func(t *testing.T) {
				c := newTestChain(t)
				govAddr := authtypes.NewModuleAddress(govtypes.ModuleName)
				setSendEnabled := banktypes.NewMsgSetSendEnabled(govAddr.String(), []*banktypes.SendEnabled{banktypes.NewSendEnabled("testdenom", false)}, nil)

				govKeeper := getKeeper[govkeeper.Keeper](c)
				minDeposit := funcs.Must(govKeeper.Params.Get(c.ctx())).MinDeposit
				proposal := funcs.Must(govv1.NewMsgSubmitProposal([]sdk.Msg{setSendEnabled}, minDeposit, c.user.addr.String(), "", "disable testdenom", "disable testdenom", false))
				res := c.deliver(c.user, proposal)
				require.Equal(t, uint32(0), res.Code, res.Log)

				res = c.deliver(c.validatorOwner, govv1.NewMsgVote(c.validatorOwner.addr, 1, govv1.OptionYes, ""))
				require.Equal(t, uint32(0), res.Code, res.Log)

				if paused {
					c.pause(permissionexported.PAUSE_LEVEL_ALL)
				}

				// let the voting period end so the gov EndBlocker executes the proposal
				for range 12 {
					c.nextBlock()
				}

				prop := funcs.Must(govKeeper.Proposals.Get(c.ctx(), 1))
				_, executed := getKeeper[bankkeeper.BaseKeeper](c).GetSendEnabledEntry(c.ctx(), "testdenom")
				if paused {
					assert.Equal(t, govv1.StatusFailed, prop.Status)
					assert.Contains(t, prop.FailedReason, "chain is paused")
					assert.False(t, executed)
				} else {
					assert.Equal(t, govv1.StatusPassed, prop.Status, prop.FailedReason)
					assert.True(t, executed)
				}
			})
		}
	})
}

func TestEmergencyHalt(t *testing.T) {
	t.Run("only access control can halt", func(t *testing.T) {
		c := newTestChain(t)

		for _, signer := range []*testAccount{c.chainManager, c.user, c.validatorOwner} {
			res := c.deliver(signer, permissiontypes.NewHaltChainRequest(signer.addr, "incident", false))
			assert.NotEqual(t, uint32(0), res.Code)
			assert.Contains(t, res.Log, "is not authorized")
		}

		_, err := getKeeper[upgradekeeper.Keeper](c).GetUpgradePlan(c.ctx())
		assert.ErrorIs(t, err, upgradetypes.ErrNoUpgradePlanFound)
	})

	t.Run("pending plan is only replaced with override", func(t *testing.T) {
		c := newTestChain(t)
		upgradeK := getKeeper[upgradekeeper.Keeper](c)

		pending := upgradetypes.Plan{Name: "v1.6", Height: c.height() + 100}
		ctx := c.ctx()
		require.NoError(t, upgradeK.ScheduleUpgrade(ctx, pending))

		res := c.deliver(c.accessControl, permissiontypes.NewHaltChainRequest(c.accessControl.addr, "incident", false))
		assert.NotEqual(t, uint32(0), res.Code)
		assert.Contains(t, res.Log, "upgrade plan v1.6 is already scheduled")
		assert.Equal(t, pending.Name, funcs.Must(upgradeK.GetUpgradePlan(c.ctx())).Name)

		res = c.deliver(c.accessControl, permissiontypes.NewHaltChainRequest(c.accessControl.addr, "incident", true))
		require.Equal(t, uint32(0), res.Code, res.Log)
		plan := funcs.Must(upgradeK.GetUpgradePlan(c.ctx()))
		assert.Equal(t, permissionexported.EmergencyHaltPlanName(c.height()+1), plan.Name)

		event := findEvent(res.Events, permissiontypes.EventTypeEmergencyHalt)
		require.NotNil(t, event)
		assert.Equal(t, "v1.6", eventAttribute(*event, permissiontypes.AttributeKeyReplacedPlanName))
	})

	t.Run("chain halts at the next height and resumes with skip upgrades", func(t *testing.T) {
		c := newTestChain(t)
		c.pause(permissionexported.PAUSE_LEVEL_ALL)

		res := c.deliver(c.accessControl, permissiontypes.NewHaltChainRequest(c.accessControl.addr, "incident", false))
		require.Equal(t, uint32(0), res.Code, res.Log)

		haltHeight := c.height() + 1
		planName := permissionexported.EmergencyHaltPlanName(haltHeight)
		plan := funcs.Must(getKeeper[upgradekeeper.Keeper](c).GetUpgradePlan(c.ctx()))
		assert.Equal(t, upgradetypes.Plan{Name: planName, Height: haltHeight, Info: "incident"}, plan)

		event := findEvent(res.Events, permissiontypes.EventTypeEmergencyHalt)
		require.NotNil(t, event)
		assert.Equal(t, c.accessControl.addr.String(), eventAttribute(*event, permissiontypes.AttributeKeySigner))
		assert.Equal(t, "incident", eventAttribute(*event, permissiontypes.AttributeKeyReason))
		assert.Equal(t, fmt.Sprint(haltHeight), eventAttribute(*event, permissiontypes.AttributeKeyHaltHeight))

		// the node refuses to execute the halt height
		_, err := c.finalizeBlock()
		require.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("UPGRADE \"%s\" NEEDED at height: %d", planName, haltHeight))
		assert.Equal(t, haltHeight-1, c.height())

		upgradeInfo := string(funcs.Must(os.ReadFile(filepath.Join(c.home, "data", "upgrade-info.json"))))
		assert.Contains(t, upgradeInfo, planName)

		// a restarted node without the skip flag stops again
		c.app = c.newApp(nil)
		_, err = c.finalizeBlock()
		require.Error(t, err)

		// a restarted node with --unsafe-skip-upgrades resumes block production, still paused
		c.app = c.newApp(simtestutil.AppOptionsMap{server.FlagUnsafeSkipUpgrades: []int{int(haltHeight)}})
		c.nextBlock()
		c.nextBlock()
		assert.Equal(t, haltHeight+1, c.height())
		assert.Equal(t, permissionexported.PAUSE_LEVEL_ALL, c.pauseLevel())

		res = c.deliver(c.accessControl, permissiontypes.NewUnpauseRequest(c.accessControl.addr, "resolved"))
		require.Equal(t, uint32(0), res.Code, res.Log)
		res = c.deliver(c.user, bankSend(c.user.addr, newTestAccount().addr))
		require.Equal(t, uint32(0), res.Code, res.Log)
	})
}

func TestEmergencyPause_SkipHaltPause(t *testing.T) {
	c := newTestChain(t)
	recipient := newTestAccount().addr
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName)
	marker := permissiontypes.PauseSkipMarker()

	send := func() []byte { return c.tx(c.user, bankSend(c.user.addr, recipient)) }
	withFlag := simtestutil.AppOptionsMap{app.FlagUnsafeSkipHaltPause: c.height() + 3}

	// a validator with the flag only marks paused blocks from the flag height on
	c.app = c.newApp(withFlag)
	assert.NotContains(t, c.prepareProposal(send()), marker, "not paused")
	c.pause(permissionexported.PAUSE_LEVEL_ALL)
	assert.NotContains(t, c.prepareProposal(send()), marker, "below the flag height")
	c.nextBlock()

	// a validator without the flag does not mark its proposals
	c.app = c.newApp(nil)
	assert.NotContains(t, c.prepareProposal(send()), marker)

	c.app = c.newApp(withFlag)
	marked := c.prepareProposal(send())
	require.Len(t, marked, 2)
	assert.Equal(t, marker, marked[0])
	assert.Equal(t, abci.ResponseProcessProposal_ACCEPT, c.processProposal(marked))

	// validators without the flag reject marked blocks
	c.app = c.newApp(nil)
	assert.Equal(t, abci.ResponseProcessProposal_REJECT, c.processProposal(marked))

	// but once committed, a node without the flag executes the marked block like everyone else
	res, err := c.finalizeBlock(marked...)
	require.NoError(t, err)
	_, err = c.app.Commit()
	require.NoError(t, err)
	assert.NotEqual(t, uint32(0), res.TxResults[0].Code, "the marker is not a tx")
	assert.Equal(t, uint32(0), res.TxResults[1].Code, res.TxResults[1].Log)
	assert.Equal(t, int64(1), c.balance(recipient).Int64())
	assert.Equal(t, permissionexported.PAUSE_LEVEL_ALL, c.pauseLevel())

	event := findEvent(res.Events, permissiontypes.EventTypeEmergencyPauseSkipped)
	require.NotNil(t, event)
	assert.Equal(t, sdk.ConsAddress(c.valSet.Proposer.Address).String(), eventAttribute(*event, permissiontypes.AttributeKeyProposer))

	// an unmarked block still enforces the pause, and validators with the flag accept it
	c.app = c.newApp(withFlag)
	unmarked := [][]byte{send()}
	assert.Equal(t, abci.ResponseProcessProposal_ACCEPT, c.processProposal(unmarked))
	txRes := c.nextBlock(unmarked...)
	assert.True(t, isChainPausedErr(txRes[0].Codespace, txRes[0].Code), txRes[0].Log)

	// from now on blocks go through proposals, so they are marked
	c.useProposals = true

	// marked blocks reject pause and halt, from anyone
	for _, msg := range []sdk.Msg{
		permissiontypes.NewPauseRequest(c.accessControl.addr, permissionexported.PAUSE_LEVEL_EXCEPT_VALIDATORS, "re-pause"),
		permissiontypes.NewHaltChainRequest(c.accessControl.addr, "halt", false),
	} {
		txRes := c.deliver(c.accessControl, msg)
		assert.Equal(t, permissiontypes.ErrPauseSkipped.ABCICode(), txRes.Code, txRes.Log)
	}
	assert.Equal(t, permissionexported.PAUSE_LEVEL_ALL, c.pauseLevel())
	_, err = getKeeper[upgradekeeper.Keeper](c).GetUpgradePlan(c.ctx())
	assert.ErrorIs(t, err, upgradetypes.ErrNoUpgradePlanFound)

	prop := c.passProposal(permissiontypes.NewHaltChainRequest(govAddr, "halt", false))
	assert.Equal(t, govv1.StatusFailed, prop.Status)
	assert.Contains(t, prop.FailedReason, permissiontypes.ErrPauseSkipped.Error())

	// gov can submit, vote and unpause
	prop = c.passProposal(permissiontypes.NewUnpauseRequest(govAddr, "gov override"))
	require.Equal(t, govv1.StatusPassed, prop.Status, prop.FailedReason)
	assert.False(t, c.pauseLevel().IsPaused())

	// once unpaused, proposers stop adding the marker
	assert.NotContains(t, c.prepareProposal(send()), marker)

	// a validator that keeps the flag also skips a later pause, until it restarts without it
	c.pause(permissionexported.PAUSE_LEVEL_ALL)
	assert.Contains(t, c.prepareProposal(send()), marker)

	c.useProposals = false
	c.app = c.newApp(nil)
	assert.NotContains(t, c.prepareProposal(send()), marker)
	txRes = c.nextBlock(send())
	assert.True(t, isChainPausedErr(txRes[0].Codespace, txRes[0].Code), txRes[0].Log)
}

func TestEmergencyPause_SkipHaltPauseNodesAgree(t *testing.T) {
	flagged := newTestChain(t)
	flagged.pause(permissionexported.PAUSE_LEVEL_ALL)
	flagged.app = flagged.newApp(simtestutil.AppOptionsMap{app.FlagUnsafeSkipHaltPause: flagged.height()})
	unflagged := flagged.fork(nil)
	nodes := []*testChain{flagged, unflagged}

	marker := permissiontypes.PauseSkipMarker()
	recipient := newTestAccount().addr
	send := func() []byte { return flagged.tx(flagged.user, bankSend(flagged.user.addr, recipient)) }

	// commits the block on both nodes, requires the same results and app hash, and returns the results
	commit := func(txs [][]byte) []*abci.ExecTxResult {
		var responses []*abci.ResponseFinalizeBlock
		for _, node := range nodes {
			res, err := node.finalizeBlock(txs...)
			require.NoError(t, err)
			_, err = node.app.Commit()
			require.NoError(t, err)
			responses = append(responses, res)
		}

		require.Equal(t, responses[0].TxResults, responses[1].TxResults)
		require.Equal(t, responses[0].AppHash, responses[1].AppHash)

		return responses[0].TxResults
	}

	// a block proposed by the node with the flag is marked, and only that node votes for it
	marked := flagged.prepareProposal(send())
	require.Equal(t, marker, marked[0])
	assert.Equal(t, abci.ResponseProcessProposal_ACCEPT, flagged.processProposal(marked))
	assert.Equal(t, abci.ResponseProcessProposal_REJECT, unflagged.processProposal(marked))

	results := commit(marked)
	assert.Equal(t, uint32(0), results[1].Code, results[1].Log)

	// only the node with the flag accepts paused txs into its mempool
	checkRes := flagged.checkTx(send())
	assert.Equal(t, uint32(0), checkRes.Code, checkRes.Log)
	checkRes = unflagged.checkTx(send())
	assert.True(t, isChainPausedErr(checkRes.Codespace, checkRes.Code), checkRes.Log)

	// a block proposed by the node without the flag is unmarked, both vote for it, and the pause applies
	unmarked := unflagged.prepareProposal(send())
	assert.NotContains(t, unmarked, marker)
	for _, node := range nodes {
		assert.Equal(t, abci.ResponseProcessProposal_ACCEPT, node.processProposal(unmarked))
	}

	results = commit(unmarked)
	assert.True(t, isChainPausedErr(results[0].Codespace, results[0].Code), results[0].Log)

	// pause is rejected in a marked block
	marked = flagged.prepareProposal(flagged.tx(flagged.accessControl,
		permissiontypes.NewPauseRequest(flagged.accessControl.addr, permissionexported.PAUSE_LEVEL_EXCEPT_VALIDATORS, "re-pause")))
	require.Equal(t, marker, marked[0])
	results = commit(marked)
	assert.Equal(t, permissiontypes.ErrPauseSkipped.ABCICode(), results[1].Code, results[1].Log)

	// unpausing in an unmarked block ends the override on both nodes
	results = commit(unflagged.prepareProposal(flagged.tx(flagged.accessControl,
		permissiontypes.NewUnpauseRequest(flagged.accessControl.addr, "resolved"))))
	assert.Equal(t, uint32(0), results[0].Code, results[0].Log)

	unmarked = flagged.prepareProposal(send())
	assert.NotContains(t, unmarked, marker)
	assert.Equal(t, abci.ResponseProcessProposal_ACCEPT, unflagged.processProposal(unmarked))
	results = commit(unmarked)
	assert.Equal(t, uint32(0), results[0].Code, results[0].Log)

	for _, node := range nodes {
		assert.False(t, node.pauseLevel().IsPaused())
		assert.Equal(t, int64(2), node.balance(recipient).Int64())
	}
}

func TestEmergencyHalt_ResumeWithPatchedBinary(t *testing.T) {
	c := newTestChain(t)

	res := c.deliver(c.accessControl, permissiontypes.NewHaltChainRequest(c.accessControl.addr, "incident", false))
	require.Equal(t, uint32(0), res.Code, res.Log)
	haltHeight := c.height() + 1
	planName := permissionexported.EmergencyHaltPlanName(haltHeight)

	_, err := c.finalizeBlock()
	require.Error(t, err)

	// a handler for the halt plan, like an entry in chainUpgrades
	c.app = c.newApp(nil)
	getKeeper[upgradekeeper.Keeper](c).SetUpgradeHandler(planName, func(_ context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return vm, nil
	})
	c.nextBlock()
	c.nextBlock()
	assert.Equal(t, haltHeight+1, c.height())

	_, err = getKeeper[upgradekeeper.Keeper](c).GetUpgradePlan(c.ctx())
	assert.ErrorIs(t, err, upgradetypes.ErrNoUpgradePlanFound)
	name, height, err := getKeeper[upgradekeeper.Keeper](c).GetLastCompletedUpgrade(c.ctx())
	require.NoError(t, err)
	assert.Equal(t, planName, name)
	assert.Equal(t, haltHeight, height)

	// a binary without the handler refuses to continue, so the chainUpgrades entry has to stay
	c.app = c.newApp(nil)
	_, err = c.finalizeBlock()
	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("upgrade handler is missing for %s upgrade plan", planName))
}

func TestEmergencyPause_RejectedTxsPayNoFee(t *testing.T) {
	c := newTestChain(t)
	fee := sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1_000))
	recipient := newTestAccount().addr

	before := c.balanceOf(c.user.addr, feeDenom)
	res := c.nextBlock(c.txWithFee(c.user, fee, bankSend(c.user.addr, recipient)))[0]
	require.Equal(t, uint32(0), res.Code, res.Log)
	require.Equal(t, before.Sub(fee.AmountOf(feeDenom)), c.balanceOf(c.user.addr, feeDenom))

	c.pause(permissionexported.PAUSE_LEVEL_ALL)

	before = c.balanceOf(c.user.addr, feeDenom)
	tx := c.txWithFee(c.user, fee, bankSend(c.user.addr, recipient))
	checkRes := c.checkTx(tx)
	assert.True(t, isChainPausedErr(checkRes.Codespace, checkRes.Code), checkRes.Log)

	// included anyway, it fails in the ante handler and the fee deduction is reverted
	res = c.nextBlock(tx)[0]
	assert.True(t, isChainPausedErr(res.Codespace, res.Code), res.Log)
	assert.Equal(t, before, c.balanceOf(c.user.addr, feeDenom))
}

func TestEmergencyPause_ExecutedByGov(t *testing.T) {
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName)

	t.Run("pause and unpause", func(t *testing.T) {
		c := newTestChain(t)

		// both are submitted before the pause, which blocks new proposals
		pauseID := c.submitProposal(permissiontypes.NewPauseRequest(govAddr, permissionexported.PAUSE_LEVEL_ALL, "gov pause"))
		c.nextBlock()
		c.nextBlock()
		unpauseID := c.submitProposal(permissiontypes.NewUnpauseRequest(govAddr, "gov unpause"))

		for c.proposal(pauseID).Status == govv1.StatusVotingPeriod {
			c.nextBlock()
		}
		require.Equal(t, govv1.StatusPassed, c.proposal(pauseID).Status, c.proposal(pauseID).FailedReason)
		assert.Equal(t, permissionexported.PAUSE_LEVEL_ALL, c.pauseLevel())
		assertRejectedAsPaused(t, c, c.user, bankSend(c.user.addr, newTestAccount().addr))

		for c.proposal(unpauseID).Status == govv1.StatusVotingPeriod {
			c.nextBlock()
		}
		require.Equal(t, govv1.StatusPassed, c.proposal(unpauseID).Status, c.proposal(unpauseID).FailedReason)
		assert.False(t, c.pauseLevel().IsPaused())
	})

	t.Run("halt", func(t *testing.T) {
		c := newTestChain(t)
		proposalID := c.submitProposal(permissiontypes.NewHaltChainRequest(govAddr, "gov halt", false))

		var err error
		for range 20 {
			if _, err = c.finalizeBlock(); err != nil {
				break
			}
			_, err = c.app.Commit()
			require.NoError(t, err)
		}

		require.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("UPGRADE \"%s\" NEEDED", permissionexported.EmergencyHaltPlanName(c.height()+1)))
		assert.Equal(t, govv1.StatusPassed, c.proposal(proposalID).Status)
	})
}

func TestEmergencyPause_Events(t *testing.T) {
	c := newTestChain(t)

	res := c.deliver(c.accessControl, permissiontypes.NewPauseRequest(c.accessControl.addr, permissionexported.PAUSE_LEVEL_ALL, "incident"))
	require.Equal(t, uint32(0), res.Code, res.Log)
	event := findEvent(res.Events, permissiontypes.EventTypeEmergencyPause)
	require.NotNil(t, event)
	assert.Equal(t, c.accessControl.addr.String(), eventAttribute(*event, permissiontypes.AttributeKeySigner))
	assert.Equal(t, "incident", eventAttribute(*event, permissiontypes.AttributeKeyReason))
	assert.Equal(t, fmt.Sprint(c.height()), eventAttribute(*event, permissiontypes.AttributeKeyHeight))
	assert.Equal(t, permissionexported.PAUSE_LEVEL_ALL.String(), eventAttribute(*event, permissiontypes.AttributeKeyPauseLevel))

	res = c.deliver(c.accessControl, permissiontypes.NewUnpauseRequest(c.accessControl.addr, "resolved"))
	require.Equal(t, uint32(0), res.Code, res.Log)
	event = findEvent(res.Events, permissiontypes.EventTypeEmergencyUnpause)
	require.NotNil(t, event)
	assert.Equal(t, c.accessControl.addr.String(), eventAttribute(*event, permissiontypes.AttributeKeySigner))
	assert.Equal(t, "resolved", eventAttribute(*event, permissiontypes.AttributeKeyReason))
	assert.Equal(t, fmt.Sprint(c.height()), eventAttribute(*event, permissiontypes.AttributeKeyHeight))
}

// valdVote is a vote in the shape vald sends it
func valdVote(sender sdk.AccAddress) *auxiliarytypes.BatchRequest {
	vote := votetypes.NewVoteRequest(sender, voteexported.PollID(1), evmtypes.NewVoteEvents(nexus.ChainName("Ethereum")))
	return auxiliarytypes.NewBatchRequest(sender, []sdk.Msg{rewardtypes.NewRefundMsgRequest(sender, vote)})
}

func submitSignature(sender sdk.AccAddress) *multisigtypes.SubmitSignatureRequest {
	priv := funcs.Must(btcec.NewPrivateKey())
	hash := sha256.Sum256([]byte("payload"))
	return multisigtypes.NewSubmitSignatureRequest(sender, 1, ec.Sign(priv, hash[:]).Serialize())
}

func findEvent(events []abci.Event, eventType string) *abci.Event {
	for _, event := range events {
		if event.Type == eventType {
			return &event
		}
	}

	return nil
}

func eventAttribute(event abci.Event, key string) string {
	for _, attr := range event.Attributes {
		if attr.Key == key {
			return attr.Value
		}
	}

	return ""
}
