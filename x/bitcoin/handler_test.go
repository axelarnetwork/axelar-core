package bitcoin

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/utils/denom"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

const testReps = 100

func TestLink_NoMasterKey(t *testing.T) {
	cdc := testutils.Codec()
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k.SetParams(ctx, types.DefaultParams())

	recipient := balance.CrossChainAddress{Address: "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7", Chain: balance.Ethereum}

	s := &btcMock.SignerMock{
		GetCurrentMasterKeyFunc: func(ctx sdk.Context, chain balance.Chain) (ecdsa.PublicKey, bool) {
			return ecdsa.PublicKey{}, false
		},
	}

	handler := NewHandler(k, &btcMock.VoterMock{}, &btcMock.RPCClientMock{}, s, &btcMock.SnapshotterMock{}, &btcMock.BalancerMock{})
	_, err := handler(ctx, types.MsgLink{Sender: sdk.AccAddress("sender"), Recipient: recipient})

	assert.Error(t, err)
	assert.Equal(t, 1, len(s.GetCurrentMasterKeyCalls()))
}

func TestLink_Success(t *testing.T) {
	cdc := testutils.Codec()
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k.SetParams(ctx, types.DefaultParams())

	recipient := balance.CrossChainAddress{Address: "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7", Chain: balance.Ethereum}
	privKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	btcAddr, err := k.GetAddress(ctx, btcec.PublicKey(privKey.PublicKey), recipient)
	if err != nil {
		panic(err)

	}
	sender := balance.CrossChainAddress{Address: btcAddr.EncodeAddress(), Chain: balance.Bitcoin}

	b := &btcMock.BalancerMock{
		LinkAddressesFunc: func(ctx sdk.Context, s balance.CrossChainAddress, r balance.CrossChainAddress) {},
	}

	s := &btcMock.SignerMock{
		GetCurrentMasterKeyFunc: func(ctx sdk.Context, chain balance.Chain) (ecdsa.PublicKey, bool) {
			return privKey.PublicKey, true
		},
	}

	handler := NewHandler(k, &btcMock.VoterMock{}, &btcMock.RPCClientMock{}, s, &btcMock.SnapshotterMock{}, b)
	_, err = handler(ctx, types.MsgLink{Sender: sdk.AccAddress("sender"), Recipient: recipient})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(b.LinkAddressesCalls()))
	assert.Equal(t, sender, b.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, b.LinkAddressesCalls()[0].Recipient)
	assert.Equal(t, 1, len(s.GetCurrentMasterKeyCalls()))
}

func TestTrackAddress(t *testing.T) {
	cdc := testutils.Codec()
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k.SetParams(ctx, types.DefaultParams())
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	rpc := btcMock.RPCClientMock{ImportAddressRescanFunc: func(address string, _ string, rescan bool) error {
		cancel()
		return nil
	}}

	sk, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	pk := btcec.PublicKey(sk.PublicKey)
	addr, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk.SerializeCompressed()), &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}

	handler := NewHandler(k, &btcMock.VoterMock{}, &rpc, &btcMock.SignerMock{}, &btcMock.SnapshotterMock{}, &btcMock.BalancerMock{})
	_, err = handler(ctx, types.NewMsgTrackAddress(sdk.AccAddress("sender"), addr.EncodeAddress(), false))

	<-timeout.Done()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rpc.ImportAddressRescanCalls()))
	assert.False(t, rpc.ImportAddressRescanCalls()[0].Rescan)
	assert.Equal(t, addr.String(), rpc.ImportAddressRescanCalls()[0].Address)
}

func TestVerifyTx_InvalidHash_VoteDiscard(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())
	rpc := btcMock.RPCClientMock{
		GetOutPointInfoFunc: func(out *wire.OutPoint) (types.OutPointInfo, error) {
			return types.OutPointInfo{}, fmt.Errorf("not found")
		},
	}
	var poll exported.PollMeta
	v := &btcMock.VoterMock{
		InitPollFunc:   func(_ sdk.Context, p exported.PollMeta) error { poll = p; return nil },
		RecordVoteFunc: func(ctx sdk.Context, vote exported.MsgVote) error { return nil },
	}

	hash, err := chainhash.NewHashFromStr("f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")
	if err != nil {
		panic(err)
	}
	info := types.OutPointInfo{
		OutPoint:      wire.NewOutPoint(hash, 0),
		Amount:        10,
		Recipient:     "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
		Confirmations: 7,
	}
	if err := info.Validate(); err != nil {
		panic(err)
	}

	handler := NewHandler(k, v, &rpc, &btcMock.SignerMock{}, &btcMock.SnapshotterMock{}, &btcMock.BalancerMock{})

	_, err = handler(ctx, types.MsgVerifyTx{Sender: sdk.AccAddress("sender"), OutPointInfo: info})
	assert.Nil(t, err)

	assert.Equal(t, 1, len(v.InitPollCalls()))
	assert.Equal(t, hash.String(), v.InitPollCalls()[0].Poll.ID)
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
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	hash, err := chainhash.NewHashFromStr("f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")
	if err != nil {
		panic(err)
	}
	info := types.OutPointInfo{
		OutPoint:      wire.NewOutPoint(hash, 0),
		Amount:        10,
		Recipient:     "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
		Confirmations: 7,
	}
	if err := info.Validate(); err != nil {
		panic(err)
	}

	rpc := btcMock.RPCClientMock{
		GetOutPointInfoFunc: func(out *wire.OutPoint) (types.OutPointInfo, error) {
			if hash.IsEqual(&out.Hash) {
				return info, nil
			}

			return types.OutPointInfo{}, fmt.Errorf("not found")
		},
	}

	var poll exported.PollMeta
	v := &btcMock.VoterMock{
		InitPollFunc: func(_ sdk.Context, p exported.PollMeta) error { poll = p; return nil },
		RecordVoteFunc: func(ctx sdk.Context, vote exported.MsgVote) error {
			return nil
		},
	}
	handler := NewHandler(k, v, &rpc, &btcMock.SignerMock{}, &btcMock.SnapshotterMock{}, &btcMock.BalancerMock{})

	_, err = handler(ctx, types.MsgVerifyTx{Sender: sdk.AccAddress("sender"), OutPointInfo: info})
	assert.Nil(t, err)

	assert.Equal(t, 1, len(v.InitPollCalls()))
	assert.Equal(t, hash.String(), v.InitPollCalls()[0].Poll.ID)
	assert.Equal(t, types.MsgVerifyTx{}.Type(), v.InitPollCalls()[0].Poll.Type)
	assert.Equal(t, types.ModuleName, v.InitPollCalls()[0].Poll.Module)

	assert.Equal(t, 1, len(v.RecordVoteCalls()))
	assert.Equal(t, poll, v.RecordVoteCalls()[0].Vote.Poll())
	assert.Equal(t, true, v.RecordVoteCalls()[0].Vote.Data())

	actualOutPoint, ok := k.GetUnverifiedOutPoint(ctx, hash.String())
	assert.True(t, ok)
	assert.True(t, info.Equals(actualOutPoint))
}

func TestVoteVerifiedTx_NoUnverifiedOutPointWithVoteResult(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	v := &btcMock.VoterMock{
		TallyVoteFunc:  func(ctx sdk.Context, vote exported.MsgVote) error { return nil },
		ResultFunc:     func(ctx sdk.Context, poll exported.PollMeta) exported.VotingData { return true },
		DeletePollFunc: func(ctx sdk.Context, poll exported.PollMeta) {},
	}

	handler := NewHandler(k, v, &btcMock.RPCClientMock{}, &btcMock.SignerMock{}, &btcMock.SnapshotterMock{}, &btcMock.BalancerMock{})
	poll := exported.PollMeta{Module: "bitcoin", Type: "verify", ID: "txid"}
	msg := &types.MsgVoteVerifiedTx{Sender: sdk.AccAddress("sender"), PollMeta: poll, VotingData: true}
	_, err := handler(ctx, msg)
	assert.Error(t, err)
}

func TestVoteVerifiedTx_IncompleteVote(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	hash, err := chainhash.NewHash(testutils.RandBytes(32))
	if err != nil {
		panic(err)
	}
	outpoint := &wire.OutPoint{
		Hash:  *hash,
		Index: 0,
	}
	outpointInfo := types.OutPointInfo{
		OutPoint:      outpoint,
		Amount:        btcutil.Amount(1000000),
		Recipient:     "sender",
		Confirmations: 100,
	}
	k.SetUnverifiedOutpoint(ctx, "txid", outpointInfo)

	poll := exported.PollMeta{Module: "bitcoin", Type: "verify", ID: "txid"}
	v := &btcMock.VoterMock{
		TallyVoteFunc:  func(ctx sdk.Context, vote exported.MsgVote) error { return nil },
		ResultFunc:     func(ctx sdk.Context, poll exported.PollMeta) exported.VotingData { return nil },
		DeletePollFunc: func(ctx sdk.Context, p exported.PollMeta) {},
	}

	b := &btcMock.BalancerMock{
		GetRecipientFunc: func(ctx sdk.Context, s balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
			return balance.CrossChainAddress{}, false
		},
		EnqueueForTransferFunc: func(ctx sdk.Context, s balance.CrossChainAddress, amount sdk.Coin) error {
			return nil
		},
	}

	handler := NewHandler(k, v, &btcMock.RPCClientMock{}, &btcMock.SignerMock{}, &btcMock.SnapshotterMock{}, b)
	msg := &types.MsgVoteVerifiedTx{Sender: sdk.AccAddress("sender"), PollMeta: poll, VotingData: true}
	_, err = handler(ctx, msg)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(v.DeletePollCalls()))
	assert.Equal(t, 0, len(b.EnqueueForTransferCalls()))
}

func TestVoteVerifiedTx_SucessNoTransfer(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	hash, err := chainhash.NewHash(testutils.RandBytes(32))
	if err != nil {
		panic(err)
	}
	outpoint := &wire.OutPoint{
		Hash:  *hash,
		Index: 0,
	}
	outpointInfo := types.OutPointInfo{
		OutPoint:      outpoint,
		Amount:        btcutil.Amount(1000000),
		Recipient:     "sender",
		Confirmations: 100,
	}
	k.SetUnverifiedOutpoint(ctx, "txid", outpointInfo)

	poll := exported.PollMeta{Module: "bitcoin", Type: "verify", ID: "txid"}
	v := &btcMock.VoterMock{
		TallyVoteFunc:  func(ctx sdk.Context, vote exported.MsgVote) error { return nil },
		ResultFunc:     func(ctx sdk.Context, poll exported.PollMeta) exported.VotingData { return true },
		DeletePollFunc: func(ctx sdk.Context, p exported.PollMeta) {},
	}

	b := &btcMock.BalancerMock{
		GetRecipientFunc: func(ctx sdk.Context, s balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
			return balance.CrossChainAddress{}, false
		},
		EnqueueForTransferFunc: func(ctx sdk.Context, s balance.CrossChainAddress, amount sdk.Coin) error { return nil },
	}

	handler := NewHandler(k, v, &btcMock.RPCClientMock{}, &btcMock.SignerMock{}, &btcMock.SnapshotterMock{}, b)
	msg := &types.MsgVoteVerifiedTx{Sender: sdk.AccAddress("sender"), PollMeta: poll, VotingData: true}
	_, err = handler(ctx, msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(v.DeletePollCalls()))
	assert.Equal(t, poll, v.DeletePollCalls()[0].Poll)
	assert.Equal(t, 0, len(b.EnqueueForTransferCalls()))
}

func TestVoteVerifiedTx_SucessAndTransfer(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	hash, err := chainhash.NewHash(testutils.RandBytes(32))
	if err != nil {
		panic(err)
	}
	outpoint := &wire.OutPoint{
		Hash:  *hash,
		Index: 0,
	}
	outpointInfo := types.OutPointInfo{
		OutPoint:      outpoint,
		Amount:        btcutil.Amount(1000000),
		Recipient:     "sender",
		Confirmations: 100,
	}
	k.SetUnverifiedOutpoint(ctx, "txid", outpointInfo)

	poll := exported.PollMeta{Module: "bitcoin", Type: "verify", ID: "txid"}
	v := &btcMock.VoterMock{
		TallyVoteFunc:  func(ctx sdk.Context, v exported.MsgVote) error { return nil },
		ResultFunc:     func(ctx sdk.Context, p exported.PollMeta) exported.VotingData { return true },
		DeletePollFunc: func(ctx sdk.Context, p exported.PollMeta) {},
	}

	sender := balance.CrossChainAddress{Address: "sender", Chain: balance.Bitcoin}
	recipient := balance.CrossChainAddress{Address: "recipient", Chain: balance.Ethereum}

	b := &btcMock.BalancerMock{
		GetRecipientFunc: func(ctx sdk.Context, s balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
			if s.Address == sender.Address {
				return recipient, true
			}
			return balance.CrossChainAddress{}, false
		},

		EnqueueForTransferFunc: func(ctx sdk.Context, s balance.CrossChainAddress, amount sdk.Coin) error {
			if s.Address != sender.Address {
				return fmt.Errorf("sender not linked to a recipient")
			}
			return nil
		},
	}

	handler := NewHandler(k, v, &btcMock.RPCClientMock{}, &btcMock.SignerMock{}, &btcMock.SnapshotterMock{}, b)
	msg := &types.MsgVoteVerifiedTx{Sender: sdk.AccAddress("sender"), PollMeta: poll, VotingData: true}
	_, err = handler(ctx, msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(v.DeletePollCalls()))
	assert.Equal(t, poll, v.DeletePollCalls()[0].Poll)
	assert.Equal(t, 1, len(b.EnqueueForTransferCalls()))
	assert.Equal(t, sender, b.EnqueueForTransferCalls()[0].Sender)
}

func TestMasterKey_RawTx_Then_Transfer(t *testing.T) {
	cdc := testutils.Codec()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	btcSubspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "btc")
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"), btcSubspace)
	k.SetParams(ctx, types.DefaultParams())

	var sk, skNext *ecdsa.PrivateKey
	var txHash []byte
	var txID, sigID string
	signer := &btcMock.SignerMock{
		GetCurrentMasterKeyIDFunc: func(ctx sdk.Context, chain balance.Chain) (string, bool) {
			return "mkID", true
		},
		StartSignFunc: func(ctx sdk.Context, keyID string, sID string, msg []byte, validators []snapshot.Validator) error {
			sigID = sID
			return nil
		},
		GetNextMasterKeyFunc: func(ctx sdk.Context, chain balance.Chain) (ecdsa.PublicKey, bool) {
			return skNext.PublicKey, true
		},
		GetSigFunc: func(ctx sdk.Context, sID string) (tss.Signature, bool) {
			if sID == sigID {
				r, s, err := ecdsa.Sign(rand.Reader, sk, txHash)
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
	v := &btcMock.VoterMock{ResultFunc: func(s sdk.Context, pollMeta exported.PollMeta) exported.VotingData {
		return pollMeta.ID == txID
	}}
	rpc := &btcMock.RPCClientMock{
		SendRawTransactionFunc: func(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error) {
			hash := tx.TxHash()
			return &hash, nil
		},
		NetworkFunc: func() types.Network {
			return types.Network(chaincfg.MainNetParams.Name)
		}}
	b := &btcMock.BalancerMock{}
	snap := &btcMock.SnapshotterMock{GetSnapshotFunc: func(ctx sdk.Context, round int64) (snapshot.Snapshot, bool) {
		return snapshot.Snapshot{}, true
	}}
	handler := NewHandler(k, v, rpc, signer, snap, b)
	querier := keeper.NewQuerier(k, signer, rpc)

	for i := 0; i < testReps; i++ {
		sk, _ = ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		skNext, _ = ecdsa.GenerateKey(btcec.S256(), rand.Reader)

		signTx := prepareMsgSign(ctx, k, querier, sk)

		res, err := handler(ctx, signTx)
		assert.NoError(t, err)
		txHash = res.Data

		_, err = querier(ctx, []string{keeper.SendTx, signTx.TxID}, abci.RequestQuery{})
		assert.NoError(t, err)

		assert.Equal(t, i+1, len(signer.GetKeyForSigIDCalls()))
		assert.Equal(t, sigID, signer.GetKeyForSigIDCalls()[i].SigID)
	}
}

func prepareMsgSign(ctx sdk.Context, k keeper.Keeper, querier sdk.Querier, sk *ecdsa.PrivateKey) types.MsgSignTx {
	hash, err := chainhash.NewHash([]byte(testutils.RandString(chainhash.HashSize)))
	if err != nil {
		panic(err)
	}

	txID := hash.String()
	btcPk := btcec.PublicKey(sk.PublicKey)
	addr, err := k.GetAddress(ctx, btcPk, balance.CrossChainAddress{})
	if err != nil {
		panic(err)
	}
	amount := btcutil.Amount(testutils.RandIntBetween(1, 100000000))
	err = k.SetUnverifiedOutpoint(ctx, txID, types.OutPointInfo{
		OutPoint:      wire.NewOutPoint(hash, uint32(testutils.RandIntBetween(0, 10))),
		Amount:        amount,
		Recipient:     addr.EncodeAddress(),
		Confirmations: uint64(testutils.RandIntBetween(7, 1000)),
	})
	if err != nil {
		panic(err)
	}
	err = k.ProcessVerificationResult(ctx, txID, true)
	if err != nil {
		panic(err)
	}
	sender := sdk.AccAddress(testutils.RandString(int(testutils.RandIntBetween(5, 50))))

	qParams := types.RawParams{TxID: txID, Satoshi: sdk.NewInt64Coin(denom.Satoshi, int64(amount))}
	bz, err := querier(ctx, []string{keeper.QueryRawTx}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(qParams)})
	if err != nil {
		panic(err)
	}
	var rawTx *wire.MsgTx
	testutils.Codec().MustUnmarshalJSON(bz, &rawTx)
	msgRawTx := types.NewMsgSignTx(sender, txID, rawTx)

	return msgRawTx
}
