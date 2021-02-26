package bitcoin

import (
	"crypto/ecdsa"
	cryptoRand "crypto/rand"
	"fmt"
	mathRand "math/rand"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"golang.org/x/crypto/ripemd160"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/utils/denom"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestLink_NoMasterKey(t *testing.T) {
	cdc := testutils.Codec()
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k.SetParams(ctx, types.DefaultParams())

	recipient := nexus.CrossChainAddress{Address: "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7", Chain: eth.Ethereum}

	s := &mock.SignerMock{GetCurrentMasterKeyIDFunc: func(sdk.Context, nexus.Chain) (string, bool) { return "", false }}

	handler := NewHandler(k, &mock.VoterMock{}, &mock.RPCClientMock{}, s, &mock.NexusMock{})
	_, err := handler(ctx, types.MsgLink{Sender: sdk.AccAddress("sender"), RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
	assert.Equal(t, 1, len(s.GetCurrentMasterKeyIDCalls()))
}

func TestLink_Success(t *testing.T) {
	cdc := testutils.Codec()
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k.SetParams(ctx, types.DefaultParams())

	recipient := nexus.CrossChainAddress{Address: "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7", Chain: eth.Ethereum}
	privKey, err := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
	if err != nil {
		panic(err)
	}

	redeemScript, err := types.CreateCrossChainRedeemScript(btcec.PublicKey(privKey.PublicKey), recipient)
	if err != nil {
		panic(err)
	}
	btcAddr, err := types.CreateDepositAddress(k.GetNetwork(ctx), redeemScript)
	if err != nil {
		panic(err)
	}
	sender := nexus.CrossChainAddress{Address: btcAddr.EncodeAddress(), Chain: exported.Bitcoin}

	chains := map[string]nexus.Chain{exported.Bitcoin.Name: exported.Bitcoin, eth.Ethereum.Name: eth.Ethereum}
	n := &mock.NexusMock{
		LinkAddressesFunc: func(ctx sdk.Context, s nexus.CrossChainAddress, r nexus.CrossChainAddress) {},
		GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
	}

	s := &mock.SignerMock{
		GetKeyFunc: func(ctx sdk.Context, keyID string) (ecdsa.PublicKey, bool) {
			return privKey.PublicKey, true
		},
		GetCurrentMasterKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain) (string, bool) { return "testkey", true },
	}

	handler := NewHandler(k, &mock.VoterMock{}, &mock.RPCClientMock{}, s, n)
	_, err = handler(ctx, types.MsgLink{Sender: sdk.AccAddress("sender"), RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(n.LinkAddressesCalls()))
	assert.Equal(t, sender, n.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, n.LinkAddressesCalls()[0].Recipient)
	assert.Equal(t, 1, len(s.GetKeyCalls()))
}

func TestVerifyTx_InvalidHash_VoteDiscard(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())
	rpc := mock.RPCClientMock{
		GetOutPointInfoFunc: func(*chainhash.Hash, *wire.OutPoint) (types.OutPointInfo, error) {
			return types.OutPointInfo{}, fmt.Errorf("not found")
		},
	}
	var poll vote.PollMeta
	v := &mock.VoterMock{
		InitPollFunc:   func(_ sdk.Context, p vote.PollMeta) error { poll = p; return nil },
		RecordVoteFunc: func(vote vote.MsgVote) {},
	}

	txHash, err := chainhash.NewHashFromStr("f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")
	if err != nil {
		panic(err)
	}
	blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	outpoint := wire.NewOutPoint(txHash, 0)
	info := types.OutPointInfo{
		OutPoint:      outpoint,
		BlockHash:     blockHash,
		Amount:        10,
		Address:       "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
		Confirmations: 7,
	}
	if err := info.Validate(); err != nil {
		panic(err)
	}

	handler := NewHandler(k, v, &rpc, &mock.SignerMock{}, &mock.NexusMock{})

	_, err = handler(ctx, types.MsgVerifyTx{Sender: sdk.AccAddress("sender"), OutPointInfo: info})
	assert.Nil(t, err)

	assert.Equal(t, 1, len(v.InitPollCalls()))
	assert.Equal(t, outpoint.String(), v.InitPollCalls()[0].Poll.ID)
	assert.Equal(t, types.MsgVerifyTx{}.Type(), v.InitPollCalls()[0].Poll.Type)
	assert.Equal(t, types.ModuleName, v.InitPollCalls()[0].Poll.Module)

	assert.Equal(t, 1, len(v.RecordVoteCalls()))
	assert.Equal(t, poll, v.RecordVoteCalls()[0].Vote.Poll())
	assert.Equal(t, false, v.RecordVoteCalls()[0].Vote.Data())
}

func TestVerifyTx_ValidUTXO(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	txHash, err := chainhash.NewHashFromStr("f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")
	if err != nil {
		panic(err)
	}
	blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	outPoint := wire.NewOutPoint(txHash, 0)
	info := types.OutPointInfo{
		OutPoint:      outPoint,
		BlockHash:     blockHash,
		Amount:        10,
		Address:       "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
		Confirmations: 7,
	}
	if err := info.Validate(); err != nil {
		panic(err)
	}

	rpc := mock.RPCClientMock{
		GetOutPointInfoFunc: func(*chainhash.Hash, *wire.OutPoint) (types.OutPointInfo, error) { return info, nil },
	}

	var poll vote.PollMeta
	v := &mock.VoterMock{
		InitPollFunc:   func(_ sdk.Context, p vote.PollMeta) error { poll = p; return nil },
		RecordVoteFunc: func(vote vote.MsgVote) {},
	}
	handler := NewHandler(k, v, &rpc, &mock.SignerMock{}, &mock.NexusMock{})

	_, err = handler(ctx, types.MsgVerifyTx{Sender: sdk.AccAddress("sender"), OutPointInfo: info})
	assert.Nil(t, err)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(rpc.GetOutPointInfoCalls()))
	assert.Equal(t, info.BlockHash, rpc.GetOutPointInfoCalls()[0].BlockHash)
	assert.Equal(t, info.OutPoint.String(), rpc.GetOutPointInfoCalls()[0].Out.String())

	assert.Equal(t, 1, len(v.InitPollCalls()))
	assert.Equal(t, outPoint.String(), v.InitPollCalls()[0].Poll.ID)
	assert.Equal(t, types.MsgVerifyTx{}.Type(), v.InitPollCalls()[0].Poll.Type)
	assert.Equal(t, types.ModuleName, v.InitPollCalls()[0].Poll.Module)

	assert.Equal(t, 1, len(v.RecordVoteCalls()))
	assert.Equal(t, poll, v.RecordVoteCalls()[0].Vote.Poll())
	assert.Equal(t, true, v.RecordVoteCalls()[0].Vote.Data())

	actualOutPoint, ok := k.GetUnverifiedOutPointInfo(ctx, info.OutPoint)
	assert.True(t, ok)
	assert.True(t, info.Equals(actualOutPoint))
}

func TestVoteVerifiedTx_NoUnverifiedOutPointWithVoteResult(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	v := &mock.VoterMock{
		TallyVoteFunc:  func(ctx sdk.Context, vote vote.MsgVote) error { return nil },
		ResultFunc:     func(ctx sdk.Context, poll vote.PollMeta) vote.VotingData { return true },
		DeletePollFunc: func(ctx sdk.Context, poll vote.PollMeta) {},
	}

	handler := NewHandler(k, v, &mock.RPCClientMock{}, &mock.SignerMock{}, &mock.NexusMock{})
	poll := vote.PollMeta{Module: "bitcoin", Type: "verify", ID: "txid"}
	msg := &types.MsgVoteVerifiedTx{Sender: sdk.AccAddress("sender"), PollMeta: poll, VotingData: true}
	_, err := handler(ctx, msg)
	assert.Error(t, err)
}

func TestVoteVerifiedTx_IncompleteVote(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	txHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	outpoint := &wire.OutPoint{
		Hash:  *txHash,
		Index: 0,
	}
	outpointInfo := types.OutPointInfo{
		OutPoint:      outpoint,
		BlockHash:     blockHash,
		Amount:        btcutil.Amount(1000000),
		Address:       "sender",
		Confirmations: 100,
	}
	k.SetUnverifiedOutpointInfo(ctx, outpointInfo)

	poll := vote.PollMeta{Module: "bitcoin", Type: "verify", ID: outpoint.String()}
	v := &mock.VoterMock{
		TallyVoteFunc:  func(ctx sdk.Context, vote vote.MsgVote) error { return nil },
		ResultFunc:     func(ctx sdk.Context, poll vote.PollMeta) vote.VotingData { return nil },
		DeletePollFunc: func(ctx sdk.Context, p vote.PollMeta) {},
	}

	b := &mock.NexusMock{
		GetRecipientFunc: func(ctx sdk.Context, s nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, false
		},
		EnqueueForTransferFunc: func(ctx sdk.Context, s nexus.CrossChainAddress, amount sdk.Coin) error {
			return nil
		},
	}

	handler := NewHandler(k, v, &mock.RPCClientMock{}, &mock.SignerMock{}, b)
	msg := &types.MsgVoteVerifiedTx{Sender: sdk.AccAddress("sender"), PollMeta: poll, VotingData: true}
	_, err = handler(ctx, msg)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(v.DeletePollCalls()))
	assert.Equal(t, 0, len(b.EnqueueForTransferCalls()))
}

func TestVoteVerifiedTx_KeyIDNotFound(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	txHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	outpoint := &wire.OutPoint{
		Hash:  *txHash,
		Index: 0,
	}
	outpointInfo := types.OutPointInfo{
		OutPoint:      outpoint,
		BlockHash:     blockHash,
		Amount:        btcutil.Amount(1000000),
		Address:       "sender",
		Confirmations: 100,
	}
	k.SetUnverifiedOutpointInfo(ctx, outpointInfo)

	poll := vote.PollMeta{Module: "bitcoin", Type: "verify", ID: outpoint.String()}
	v := &mock.VoterMock{
		TallyVoteFunc:  func(ctx sdk.Context, vote vote.MsgVote) error { return nil },
		ResultFunc:     func(ctx sdk.Context, poll vote.PollMeta) vote.VotingData { return true },
		DeletePollFunc: func(ctx sdk.Context, p vote.PollMeta) {},
	}

	b := &mock.NexusMock{
		GetRecipientFunc: func(ctx sdk.Context, s nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, false
		},
		EnqueueForTransferFunc: func(ctx sdk.Context, s nexus.CrossChainAddress, amount sdk.Coin) error { return nil },
	}

	handler := NewHandler(k, v, &mock.RPCClientMock{}, &mock.SignerMock{}, b)
	msg := &types.MsgVoteVerifiedTx{Sender: sdk.AccAddress("sender"), PollMeta: poll, VotingData: true}
	_, err = handler(ctx, msg)
	assert.Error(t, err)
	assert.Equal(t, 1, len(v.DeletePollCalls()))
	assert.Equal(t, poll, v.DeletePollCalls()[0].Poll)
	assert.Equal(t, 0, len(b.EnqueueForTransferCalls()))
}

func TestVoteVerifiedTx_Success_NotLinked(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	txHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	outpoint := &wire.OutPoint{
		Hash:  *txHash,
		Index: 0,
	}
	sender := randomAddress()
	outpointInfo := types.OutPointInfo{
		OutPoint:      outpoint,
		BlockHash:     blockHash,
		Amount:        btcutil.Amount(1000000),
		Address:       sender.EncodeAddress(),
		Confirmations: 100,
	}
	k.SetUnverifiedOutpointInfo(ctx, outpointInfo)

	k.SetKeyIDByAddress(ctx, sender, "testkey")

	poll := vote.PollMeta{Module: "bitcoin", Type: "verify", ID: outpoint.String()}
	v := &mock.VoterMock{
		TallyVoteFunc:  func(ctx sdk.Context, vote vote.MsgVote) error { return nil },
		ResultFunc:     func(ctx sdk.Context, poll vote.PollMeta) vote.VotingData { return true },
		DeletePollFunc: func(ctx sdk.Context, p vote.PollMeta) {},
	}

	b := &mock.NexusMock{
		GetRecipientFunc: func(ctx sdk.Context, s nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, false
		},
		EnqueueForTransferFunc: func(ctx sdk.Context, s nexus.CrossChainAddress, amount sdk.Coin) error {
			return fmt.Errorf("not linked")
		},
	}

	handler := NewHandler(k, v, &mock.RPCClientMock{}, &mock.SignerMock{}, b)
	msg := &types.MsgVoteVerifiedTx{Sender: sdk.AccAddress("sender"), PollMeta: poll, VotingData: true}
	_, err = handler(ctx, msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(v.DeletePollCalls()))
	assert.Equal(t, poll, v.DeletePollCalls()[0].Poll)
	assert.Equal(t, 1, len(b.EnqueueForTransferCalls()))
}

func TestVoteVerifiedTx_SucessAndTransfer(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	txHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	outpoint := &wire.OutPoint{
		Hash:  *txHash,
		Index: 0,
	}
	btcSender := randomAddress()
	outpointInfo := types.OutPointInfo{
		OutPoint:      outpoint,
		BlockHash:     blockHash,
		Amount:        btcutil.Amount(1000000),
		Address:       btcSender.EncodeAddress(),
		Confirmations: 100,
	}
	k.SetUnverifiedOutpointInfo(ctx, outpointInfo)
	k.SetKeyIDByAddress(ctx, btcSender, "testkey")

	poll := vote.PollMeta{Module: "bitcoin", Type: "verify", ID: outpoint.String()}
	v := &mock.VoterMock{
		TallyVoteFunc:  func(ctx sdk.Context, v vote.MsgVote) error { return nil },
		ResultFunc:     func(ctx sdk.Context, p vote.PollMeta) vote.VotingData { return true },
		DeletePollFunc: func(ctx sdk.Context, p vote.PollMeta) {},
	}

	sender := nexus.CrossChainAddress{Address: btcSender.EncodeAddress(), Chain: exported.Bitcoin}
	recipient := nexus.CrossChainAddress{Address: "recipient", Chain: eth.Ethereum}

	b := &mock.NexusMock{
		GetRecipientFunc: func(ctx sdk.Context, s nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return recipient, true
		},

		EnqueueForTransferFunc: func(ctx sdk.Context, s nexus.CrossChainAddress, amount sdk.Coin) error {
			if s.Address != sender.Address {
				return fmt.Errorf("sender not linked to a recipient")
			}
			return nil
		},
	}

	handler := NewHandler(k, v, &mock.RPCClientMock{}, &mock.SignerMock{}, b)
	msg := &types.MsgVoteVerifiedTx{Sender: sdk.AccAddress("btcSender"), PollMeta: poll, VotingData: true}
	_, err = handler(ctx, msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(v.DeletePollCalls()))
	assert.Equal(t, poll, v.DeletePollCalls()[0].Poll)
	assert.Equal(t, 1, len(b.EnqueueForTransferCalls()))
	assert.Equal(t, sender, b.EnqueueForTransferCalls()[0].Sender)
}

type mocks struct {
	*mock.RPCClientMock
	*mock.VoterMock
	*mock.SignerMock
	*mock.NexusMock
}
type expectedResult struct {
	depositCount  int
	transferCount int
	hasError      bool
}

func TestNewHandler_SignPendingTransfers(t *testing.T) {
	var (
		ctx  sdk.Context
		k    keeper.Keeper
		m    mocks
		h    sdk.Handler
		sigs []btcec.Signature
	)

	init := func() {
		cdc := testutils.Codec()
		btcSubspace := params.NewSubspace(cdc, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), "btc")
		ctx = sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
		k = keeper.NewKeeper(cdc, sdk.NewKVStoreKey("btc"), btcSubspace)
		k.SetParams(ctx, types.DefaultParams())

		sigs = make([]btcec.Signature, 0)
		sk, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
		chains := map[string]nexus.Chain{exported.Bitcoin.Name: exported.Bitcoin, eth.Ethereum.Name: eth.Ethereum}
		m = mocks{
			&mock.RPCClientMock{},
			&mock.VoterMock{
				InitPollFunc:   func(ctx sdk.Context, poll vote.PollMeta) error { return nil },
				TallyVoteFunc:  func(sdk.Context, vote.MsgVote) error { return nil },
				ResultFunc:     func(sdk.Context, vote.PollMeta) vote.VotingData { return true },
				DeletePollFunc: func(ctx sdk.Context, poll vote.PollMeta) {},
				RecordVoteFunc: func(vote vote.MsgVote) {}},
			&mock.SignerMock{
				GetKeyFunc: func(sdk.Context, string) (ecdsa.PublicKey, bool) { return sk.PublicKey, true },
				GetCurrentMasterKeyFunc: func(sdk.Context, nexus.Chain) (ecdsa.PublicKey, bool) {
					return sk.PublicKey, true
				},
				GetCurrentMasterKeyIDFunc: func(sdk.Context, nexus.Chain) (string, bool) {
					return testutils.RandStringBetween(5, 20), true
				},
				GetNextMasterKeyIDFunc: func(sdk.Context, nexus.Chain) (string, bool) { return "", false },
				GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
					return testutils.RandPosInt(), true
				},
				StartSignFunc: func(_ sdk.Context, _ string, _ string, msg []byte) error {
					r, s, _ := ecdsa.Sign(cryptoRand.Reader, sk, msg)
					sigs = append(sigs, btcec.Signature{R: r, S: s})
					return nil
				},
			},
			&mock.NexusMock{
				LinkAddressesFunc:          func(sdk.Context, nexus.CrossChainAddress, nexus.CrossChainAddress) {},
				EnqueueForTransferFunc:     func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error { return nil },
				ArchivePendingTransferFunc: func(ctx sdk.Context, transfer nexus.CrossChainTransfer) {},
				GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
					c, ok := chains[chain]
					return c, ok
				},
			},
		}
		h = NewHandler(k, m.VoterMock, m.RPCClientMock, m.SignerMock, m.NexusMock)
	}

	testCases := []struct {
		label   string
		prepare func(sdk.Handler, sdk.Context, mocks) (sdk.Msg, expectedResult)
	}{
		{"nothing pending", prepareMsgSignPendingTransfersDoNothing},
		{"not enough deposits", prepareMsgSignPendingTransfersNotEnoughDeposits},
		{"successful completion", prepareMsgSignPendingTransfersSuccessful},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			init()
			msg, expected := testCase.prepare(h, ctx, m)
			_, err := h(ctx, msg)

			if expected.hasError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, expected.depositCount, len(m.SignerMock.StartSignCalls()))
			assert.Equal(t, expected.transferCount, len(m.NexusMock.ArchivePendingTransferCalls()))
			if expected.transferCount > 0 {
				_, err = k.AssembleBtcTx(ctx, k.GetRawTx(ctx), sigs)
				assert.NoError(t, err)
			}
		})
	}
}

func prepareMsgSignPendingTransfersSuccessful(h sdk.Handler, ctx sdk.Context, m mocks) (sdk.Msg, expectedResult) {
	var transfers []nexus.CrossChainTransfer
	totalAmount := sdk.ZeroInt()
	transferCount := int(testutils.RandIntBetween(1, 100))
	for i := 0; i < transferCount; i++ {
		transfer := randomTransfer()
		totalAmount = totalAmount.Add(transfer.Asset.Amount)
		transfers = append(transfers, transfer)
	}

	fee := btcutil.Amount(testutils.RandPosInt())

	totalDeposits := sdk.ZeroInt()
	depositCount := 0
	for ; totalDeposits.SubRaw(int64(fee)).LT(totalAmount); depositCount++ {
		res, _ := h(ctx, randomMsgLink())
		msgVerifyTx := randomMsgVerifyTx(string(res.Data))
		totalDeposits = totalDeposits.AddRaw(int64(msgVerifyTx.OutPointInfo.Amount))
		m.RPCClientMock.GetOutPointInfoFunc = func(*chainhash.Hash, *wire.OutPoint) (types.OutPointInfo, error) {
			return msgVerifyTx.OutPointInfo, nil
		}

		_, _ = h(ctx, msgVerifyTx)
		_, _ = h(ctx, getMsgVoteVerifyTx(msgVerifyTx, true))
	}

	m.NexusMock.GetPendingTransfersForChainFunc = func(ctx sdk.Context, chain nexus.Chain) []nexus.CrossChainTransfer {
		return transfers
	}

	return types.NewMsgSignPendingTransfers(sdk.AccAddress(testutils.RandStringBetween(5, 20)), fee),
		expectedResult{
			depositCount:  depositCount,
			transferCount: transferCount,
			hasError:      false,
		}
}

func prepareMsgSignPendingTransfersNotEnoughDeposits(h sdk.Handler, ctx sdk.Context, m mocks) (sdk.Msg, expectedResult) {
	totalDeposits := sdk.ZeroInt()
	depositCount := int(testutils.RandIntBetween(1, 100))
	for i := 0; i < depositCount; i++ {
		res, _ := h(ctx, randomMsgLink())
		msgVerifyTx := randomMsgVerifyTx(string(res.Data))
		totalDeposits = totalDeposits.AddRaw(int64(msgVerifyTx.OutPointInfo.Amount))
		m.RPCClientMock.GetOutPointInfoFunc = func(*chainhash.Hash, *wire.OutPoint) (types.OutPointInfo, error) {
			return msgVerifyTx.OutPointInfo, nil
		}
		_, _ = h(ctx, msgVerifyTx)
		_, _ = h(ctx, getMsgVoteVerifyTx(msgVerifyTx, true))
	}

	fee := btcutil.Amount(testutils.RandPosInt())

	var transfers []nexus.CrossChainTransfer
	totalAmount := sdk.ZeroInt()
	for totalAmount.AddRaw(int64(fee)).LTE(totalDeposits) {
		transfer := randomTransfer()
		totalAmount = totalAmount.Add(transfer.Asset.Amount)
		transfers = append(transfers, transfer)
	}
	m.NexusMock.GetPendingTransfersForChainFunc = func(ctx sdk.Context, chain nexus.Chain) []nexus.CrossChainTransfer {
		return transfers
	}

	return types.NewMsgSignPendingTransfers(sdk.AccAddress(testutils.RandStringBetween(5, 20)), fee),
		expectedResult{
			depositCount:  0,
			transferCount: 0,
			hasError:      true,
		}
}

func randomMsgLink() types.MsgLink {
	return types.MsgLink{
		Sender:         sdk.AccAddress(testutils.RandStringBetween(5, 20)),
		RecipientAddr:  testutils.RandStringBetween(5, 100),
		RecipientChain: eth.Ethereum.Name,
	}
}

func getMsgVoteVerifyTx(msgVerifyTx types.MsgVerifyTx, result bool) *types.MsgVoteVerifiedTx {
	return &types.MsgVoteVerifiedTx{
		Sender: sdk.AccAddress(testutils.RandStringBetween(5, 20)),
		PollMeta: vote.PollMeta{
			Module: types.ModuleName,
			Type:   msgVerifyTx.Type(),
			ID:     msgVerifyTx.OutPointInfo.OutPoint.String(),
		},
		VotingData: result,
	}
}

func randomMsgVerifyTx(addr string) types.MsgVerifyTx {
	txHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	conf := mathRand.Uint64()
	if conf == 0 {
		conf += 1
	}
	return types.NewMsgVerifyTx(sdk.AccAddress(testutils.RandStringBetween(5, 20)), types.OutPointInfo{
		OutPoint:      wire.NewOutPoint(txHash, mathRand.Uint32()),
		Amount:        btcutil.Amount(testutils.RandPosInt()),
		BlockHash:     blockHash,
		Address:       addr,
		Confirmations: conf,
	})
}

func randomTransfer() nexus.CrossChainTransfer {
	return nexus.CrossChainTransfer{
		Recipient: nexus.CrossChainAddress{Chain: exported.Bitcoin, Address: randomAddress().EncodeAddress()},
		Asset:     sdk.NewInt64Coin(denom.Satoshi, testutils.RandPosInt()),
		ID:        mathRand.Uint64(),
	}
}

func prepareMsgSignPendingTransfersDoNothing(_ sdk.Handler, _ sdk.Context, m mocks) (sdk.Msg, expectedResult) {
	m.NexusMock.GetPendingTransfersForChainFunc = func(ctx sdk.Context, chain nexus.Chain) []nexus.CrossChainTransfer {
		return nil
	}

	return types.NewMsgSignPendingTransfers(
			sdk.AccAddress(testutils.RandStringBetween(5, 20)),
			btcutil.Amount(testutils.RandPosInt()),
		), expectedResult{
			depositCount:  0,
			transferCount: 0,
			hasError:      true,
		}
}

func randomAddress() btcutil.Address {
	addr, err := btcutil.NewAddressScriptHashFromHash(testutils.RandBytes(ripemd160.Size), types.DefaultParams().Network.Params)
	if err != nil {
		panic(err)
	}
	return addr
}
