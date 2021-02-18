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
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

const testReps = 100

func TestLink_NoMasterKey(t *testing.T) {
	cdc := testutils.Codec()
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k.SetParams(ctx, types.DefaultParams())

	recipient := balance.CrossChainAddress{Address: "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7", Chain: eth.Ethereum}

	s := &mock.SignerMock{GetCurrentMasterKeyIDFunc: func(sdk.Context, balance.Chain) (string, bool) { return "", false }}

	handler := NewHandler(k, &mock.VoterMock{}, &mock.RPCClientMock{}, s, &mock.SnapshotterMock{}, &mock.BalancerMock{})
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

	recipient := balance.CrossChainAddress{Address: "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7", Chain: eth.Ethereum}
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
	sender := balance.CrossChainAddress{Address: btcAddr.EncodeAddress(), Chain: exported.Bitcoin}

	chains := map[string]balance.Chain{exported.Bitcoin.Name: exported.Bitcoin, eth.Ethereum.Name: eth.Ethereum}
	b := &mock.BalancerMock{
		LinkAddressesFunc: func(ctx sdk.Context, s balance.CrossChainAddress, r balance.CrossChainAddress) {},
		GetChainFunc: func(ctx sdk.Context, chain string) (balance.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
	}

	s := &mock.SignerMock{
		GetKeyFunc: func(ctx sdk.Context, keyID string) (ecdsa.PublicKey, bool) {
			return privKey.PublicKey, true
		},
		GetCurrentMasterKeyIDFunc: func(ctx sdk.Context, chain balance.Chain) (string, bool) { return "testkey", true },
	}

	handler := NewHandler(k, &mock.VoterMock{}, &mock.RPCClientMock{}, s, &mock.SnapshotterMock{}, b)
	_, err = handler(ctx, types.MsgLink{Sender: sdk.AccAddress("sender"), RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(b.LinkAddressesCalls()))
	assert.Equal(t, sender, b.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, b.LinkAddressesCalls()[0].Recipient)
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
		RecordVoteFunc: func(ctx sdk.Context, vote vote.MsgVote) error { return nil },
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

	handler := NewHandler(k, v, &rpc, &mock.SignerMock{}, &mock.SnapshotterMock{}, &mock.BalancerMock{})

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
		InitPollFunc: func(_ sdk.Context, p vote.PollMeta) error { poll = p; return nil },
		RecordVoteFunc: func(ctx sdk.Context, vote vote.MsgVote) error {
			return nil
		},
	}
	handler := NewHandler(k, v, &rpc, &mock.SignerMock{}, &mock.SnapshotterMock{}, &mock.BalancerMock{})

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

	handler := NewHandler(k, v, &mock.RPCClientMock{}, &mock.SignerMock{}, &mock.SnapshotterMock{}, &mock.BalancerMock{})
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

	poll := vote.PollMeta{Module: "bitcoin", Type: "verify", ID: "txid"}
	v := &mock.VoterMock{
		TallyVoteFunc:  func(ctx sdk.Context, vote vote.MsgVote) error { return nil },
		ResultFunc:     func(ctx sdk.Context, poll vote.PollMeta) vote.VotingData { return nil },
		DeletePollFunc: func(ctx sdk.Context, p vote.PollMeta) {},
	}

	b := &mock.BalancerMock{
		GetRecipientFunc: func(ctx sdk.Context, s balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
			return balance.CrossChainAddress{}, false
		},
		EnqueueForTransferFunc: func(ctx sdk.Context, s balance.CrossChainAddress, amount sdk.Coin) error {
			return nil
		},
	}

	handler := NewHandler(k, v, &mock.RPCClientMock{}, &mock.SignerMock{}, &mock.SnapshotterMock{}, b)
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

	b := &mock.BalancerMock{
		GetRecipientFunc: func(ctx sdk.Context, s balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
			return balance.CrossChainAddress{}, false
		},
		EnqueueForTransferFunc: func(ctx sdk.Context, s balance.CrossChainAddress, amount sdk.Coin) error { return nil },
	}

	handler := NewHandler(k, v, &mock.RPCClientMock{}, &mock.SignerMock{}, &mock.SnapshotterMock{}, b)
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

	b := &mock.BalancerMock{
		GetRecipientFunc: func(ctx sdk.Context, s balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
			return balance.CrossChainAddress{}, false
		},
		EnqueueForTransferFunc: func(ctx sdk.Context, s balance.CrossChainAddress, amount sdk.Coin) error {
			return fmt.Errorf("not linked")
		},
	}

	handler := NewHandler(k, v, &mock.RPCClientMock{}, &mock.SignerMock{}, &mock.SnapshotterMock{}, b)
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

	sender := balance.CrossChainAddress{Address: btcSender.EncodeAddress(), Chain: exported.Bitcoin}
	recipient := balance.CrossChainAddress{Address: "recipient", Chain: eth.Ethereum}

	b := &mock.BalancerMock{
		GetRecipientFunc: func(ctx sdk.Context, s balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
			return recipient, true
		},

		EnqueueForTransferFunc: func(ctx sdk.Context, s balance.CrossChainAddress, amount sdk.Coin) error {
			if s.Address != sender.Address {
				return fmt.Errorf("sender not linked to a recipient")
			}
			return nil
		},
	}

	handler := NewHandler(k, v, &mock.RPCClientMock{}, &mock.SignerMock{}, &mock.SnapshotterMock{}, b)
	msg := &types.MsgVoteVerifiedTx{Sender: sdk.AccAddress("btcSender"), PollMeta: poll, VotingData: true}
	_, err = handler(ctx, msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(v.DeletePollCalls()))
	assert.Equal(t, poll, v.DeletePollCalls()[0].Poll)
	assert.Equal(t, 1, len(b.EnqueueForTransferCalls()))
	assert.Equal(t, sender, b.EnqueueForTransferCalls()[0].Sender)
}

func TestSignTx(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	var sk, skNext *ecdsa.PrivateKey
	var txHash []byte
	var txID, sigID string
	signer := &mock.SignerMock{
		GetCurrentMasterKeyIDFunc: func(ctx sdk.Context, chain balance.Chain) (string, bool) {
			return "mkID", true
		},
		StartSignFunc: func(ctx sdk.Context, keyID string, sID string, msg []byte, validators []snapshot.Validator) error {
			sigID = sID
			txHash = msg
			return nil
		},
		GetNextMasterKeyFunc: func(ctx sdk.Context, chain balance.Chain) (ecdsa.PublicKey, bool) {
			return skNext.PublicKey, true
		},
		GetSigFunc: func(ctx sdk.Context, sID string) (tss.Signature, bool) {
			if sID == sigID {
				r, s, err := ecdsa.Sign(cryptoRand.Reader, sk, txHash)
				if err != nil {
					panic(err)
				}
				return tss.Signature{R: r, S: s}, true
			}
			return tss.Signature{}, false
		},
		GetKeyForSigIDFunc: func(ctx sdk.Context, sigID string) (ecdsa.PublicKey, bool) {
			return sk.PublicKey, true
		},
		GetSnapshotRoundForKeyIDFunc: func(ctx sdk.Context, keyID string) (int64, bool) {
			return testutils.RandIntBetween(0, 100000), true
		},
	}
	v := &mock.VoterMock{ResultFunc: func(s sdk.Context, pollMeta vote.PollMeta) vote.VotingData {
		return pollMeta.ID == txID
	}}
	rpc := &mock.RPCClientMock{
		SendRawTransactionFunc: func(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error) {
			hash := tx.TxHash()
			return &hash, nil
		},
		NetworkFunc: func() types.Network { return types.Mainnet }}
	b := &mock.BalancerMock{}
	snap := &mock.SnapshotterMock{GetSnapshotFunc: func(ctx sdk.Context, round int64) (snapshot.Snapshot, bool) {
		return snapshot.Snapshot{}, true
	}}
	handler := NewHandler(k, v, rpc, signer, snap, b)
	querier := keeper.NewQuerier(k, signer, b, rpc)

	for _, recpAddr := range testutils.RandStrings(5, 20).Take(testReps) {
		sk, _ = ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
		skNext, _ = ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
		recipient := balance.CrossChainAddress{Chain: eth.Ethereum,
			Address: recpAddr,
		}
		b.GetRecipientFunc = func(ctx sdk.Context, sender balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
			return recipient, true
		}

		signTx := prepareMsgSign(ctx, k, querier, sk, recipient)

		_, err := handler(ctx, signTx)
		assert.NoError(t, err)

		_, err = querier(ctx, []string{keeper.SendTx}, abci.RequestQuery{Data: cdc.MustMarshalJSON(signTx.Outpoint)})
		assert.NoError(t, err)
	}
}

type mocks struct {
	*mock.RPCClientMock
	*mock.VoterMock
	*mock.SignerMock
	*mock.SnapshotterMock
	*mock.BalancerMock
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
		chains := map[string]balance.Chain{exported.Bitcoin.Name: exported.Bitcoin, eth.Ethereum.Name: eth.Ethereum}
		m = mocks{
			&mock.RPCClientMock{},
			&mock.VoterMock{
				InitPollFunc:   func(ctx sdk.Context, poll vote.PollMeta) error { return nil },
				TallyVoteFunc:  func(sdk.Context, vote.MsgVote) error { return nil },
				ResultFunc:     func(sdk.Context, vote.PollMeta) vote.VotingData { return true },
				DeletePollFunc: func(ctx sdk.Context, poll vote.PollMeta) {},
				RecordVoteFunc: func(sdk.Context, vote.MsgVote) error { return nil }},
			&mock.SignerMock{
				GetKeyFunc: func(sdk.Context, string) (ecdsa.PublicKey, bool) { return sk.PublicKey, true },
				GetCurrentMasterKeyFunc: func(sdk.Context, balance.Chain) (ecdsa.PublicKey, bool) {
					return sk.PublicKey, true
				},
				GetCurrentMasterKeyIDFunc: func(sdk.Context, balance.Chain) (string, bool) {
					return testutils.RandStringBetween(5, 20), true
				},
				GetSnapshotRoundForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
					return testutils.RandPosInt(), true
				},
				StartSignFunc: func(_ sdk.Context, _ string, _ string, msg []byte, _ []snapshot.Validator) error {
					r, s, _ := ecdsa.Sign(cryptoRand.Reader, sk, msg)
					sigs = append(sigs, btcec.Signature{R: r, S: s})
					return nil
				},
			},
			&mock.SnapshotterMock{
				GetSnapshotFunc: func(sdk.Context, int64) (snapshot.Snapshot, bool) {
					return snapshot.Snapshot{}, true
				},
			},
			&mock.BalancerMock{
				LinkAddressesFunc:          func(sdk.Context, balance.CrossChainAddress, balance.CrossChainAddress) {},
				EnqueueForTransferFunc:     func(sdk.Context, balance.CrossChainAddress, sdk.Coin) error { return nil },
				ArchivePendingTransferFunc: func(ctx sdk.Context, transfer balance.CrossChainTransfer) {},
				GetChainFunc: func(ctx sdk.Context, chain string) (balance.Chain, bool) {
					c, ok := chains[chain]
					return c, ok
				},
			},
		}
		h = NewHandler(k, m.VoterMock, m.RPCClientMock, m.SignerMock, m.SnapshotterMock, m.BalancerMock)
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
			assert.Equal(t, expected.transferCount, len(m.BalancerMock.ArchivePendingTransferCalls()))
			if expected.transferCount > 0 {
				_, err = k.AssembleBtcTx(ctx, k.GetRawConsolidationTx(ctx), sigs)
				assert.NoError(t, err)
			}
		})
	}
}

func prepareMsgSignPendingTransfersSuccessful(h sdk.Handler, ctx sdk.Context, m mocks) (sdk.Msg, expectedResult) {
	var transfers []balance.CrossChainTransfer
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

	m.BalancerMock.GetPendingTransfersForChainFunc = func(ctx sdk.Context, chain balance.Chain) []balance.CrossChainTransfer {
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

	var transfers []balance.CrossChainTransfer
	totalAmount := sdk.ZeroInt()
	for totalAmount.AddRaw(int64(fee)).LTE(totalDeposits) {
		transfer := randomTransfer()
		totalAmount = totalAmount.Add(transfer.Asset.Amount)
		transfers = append(transfers, transfer)
	}
	m.BalancerMock.GetPendingTransfersForChainFunc = func(ctx sdk.Context, chain balance.Chain) []balance.CrossChainTransfer {
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

func randomTransfer() balance.CrossChainTransfer {
	return balance.CrossChainTransfer{
		Recipient: balance.CrossChainAddress{Chain: exported.Bitcoin, Address: randomAddress().EncodeAddress()},
		Asset:     sdk.NewInt64Coin(denom.Satoshi, testutils.RandPosInt()),
		ID:        mathRand.Uint64(),
	}
}

func prepareMsgSignPendingTransfersDoNothing(_ sdk.Handler, _ sdk.Context, m mocks) (sdk.Msg, expectedResult) {
	m.BalancerMock.GetPendingTransfersForChainFunc = func(ctx sdk.Context, chain balance.Chain) []balance.CrossChainTransfer {
		return nil
	}

	return types.NewMsgSignPendingTransfers(
			sdk.AccAddress(testutils.RandStringBetween(5, 20)),
			btcutil.Amount(testutils.RandPosInt()),
		), expectedResult{
			depositCount:  0,
			transferCount: 0,
			hasError:      false,
		}
}

func prepareMsgSign(ctx sdk.Context, k keeper.Keeper, querier sdk.Querier, sk *ecdsa.PrivateKey, recipient balance.CrossChainAddress) types.MsgSignTx {
	hash, err := chainhash.NewHash([]byte(testutils.RandString(chainhash.HashSize)))
	if err != nil {
		panic(err)
	}

	btcPk := btcec.PublicKey(sk.PublicKey)
	script, err := types.CreateCrossChainRedeemScript(btcPk, recipient)
	if err != nil {
		panic(err)
	}
	addr, err := types.CreateDepositAddress(k.GetNetwork(ctx), script)
	if err != nil {
		panic(err)
	}
	k.SetRedeemScriptByAddress(ctx, addr, script)
	amount := btcutil.Amount(testutils.RandIntBetween(1, 100000000))
	outPoint := wire.NewOutPoint(hash, uint32(testutils.RandIntBetween(0, 10)))
	k.SetUnverifiedOutpointInfo(ctx, types.OutPointInfo{
		OutPoint:      outPoint,
		Amount:        amount,
		Address:       addr.EncodeAddress(),
		Confirmations: uint64(testutils.RandIntBetween(7, 1000)),
	})
	k.ProcessVerificationResult(ctx, outPoint.String(), true)
	sender := sdk.AccAddress(testutils.RandString(int(testutils.RandIntBetween(5, 50))))

	qParams := types.RawTxParams{OutPoint: outPoint, Satoshi: sdk.NewInt64Coin(denom.Satoshi, int64(amount)), DepositAddr: addr.EncodeAddress()}
	bz, err := querier(ctx, []string{keeper.QueryRawTx}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(qParams)})
	if err != nil {
		panic(err)
	}
	var rawTx *wire.MsgTx
	testutils.Codec().MustUnmarshalJSON(bz, &rawTx)
	msgRawTx := types.NewMsgSignTx(sender, outPoint, rawTx)

	k.SetKeyIDByOutpoint(ctx, outPoint, "testkey")

	return msgRawTx
}

func randomAddress() btcutil.Address {
	addr, err := btcutil.NewAddressScriptHashFromHash(testutils.RandBytes(ripemd160.Size), types.DefaultParams().Network.Params)
	if err != nil {
		panic(err)
	}
	return addr
}
