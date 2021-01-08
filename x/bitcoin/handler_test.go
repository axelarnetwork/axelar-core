package bitcoin

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

const testReps = 100

func TestTrackAddress(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
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
	expectedAddress, err := types.ParseBtcAddress(addr.EncodeAddress(), types.Network(chaincfg.MainNetParams.Name))
	if err != nil {
		panic(err)
	}

	handler := NewHandler(k, &btcMock.VoterMock{}, &rpc, &btcMock.SignerMock{}, &btcMock.BalancerMock{})
	_, err = handler(ctx, types.NewMsgTrackAddress(sdk.AccAddress("sender"), expectedAddress, false))

	<-timeout.Done()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rpc.ImportAddressRescanCalls()))
	assert.False(t, rpc.ImportAddressRescanCalls()[0].Rescan)
	assert.Equal(t, expectedAddress.String(), rpc.ImportAddressRescanCalls()[0].Address)
}

func TestVerifyTx_InvalidHash_VoteDiscard(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"))
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

	hash, _ := chainhash.NewHashFromStr("f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")
	recipient, _ := types.ParseBtcAddress("bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", "mainnet")
	info := types.OutPointInfo{
		OutPoint:      wire.NewOutPoint(hash, 0),
		Amount:        10,
		Recipient:     recipient,
		Confirmations: 7,
	}
	if err := info.Validate(); err != nil {
		panic(err)
	}

	handler := NewHandler(k, v, &rpc, &btcMock.SignerMock{}, &btcMock.BalancerMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	_, err := handler(ctx, types.MsgVerifyTx{Sender: sdk.AccAddress("sender"), OutPointInfo: info})
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
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"))

	hash, _ := chainhash.NewHashFromStr("f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")
	recipient, _ := types.ParseBtcAddress("bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", "mainnet")
	info := types.OutPointInfo{
		OutPoint:      wire.NewOutPoint(hash, 0),
		Amount:        10,
		Recipient:     recipient,
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
	handler := NewHandler(k, v, &rpc, &btcMock.SignerMock{}, &btcMock.BalancerMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	_, err := handler(ctx, types.MsgVerifyTx{Sender: sdk.AccAddress("sender"), OutPointInfo: info})
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

func TestMasterKey_RawTx_Then_Transfer(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey(types.StoreKey))

	var sk, skNext *ecdsa.PrivateKey
	var txHash []byte
	var txID, sigID string
	signer := &btcMock.SignerMock{
		GetCurrentMasterKeyFunc: func(ctx sdk.Context, chain balance.Chain) (ecdsa.PublicKey, bool) {
			return sk.PublicKey, true
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
	b := &btcMock.BalancerMock{GetRecipientFunc: func(ctx sdk.Context, sender balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
		return balance.CrossChainAddress{}, false
	}}
	handler := NewHandler(k, v, rpc, signer, b)

	for i := 0; i < testReps; i++ {
		ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
		sk, _ = ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		skNext, _ = ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		sigID = testutils.RandString(int(testutils.RandIntBetween(5, 20)))

		rawTx, transfer := prepareMsgTransferToNewMasterKey(ctx, k, signer, b, rpc, sk, sigID)
		txID = transfer.TxID

		res, err := handler(ctx, rawTx)
		assert.NoError(t, err)
		txHash = res.Data

		_, err = handler(ctx, transfer)
		assert.NoError(t, err)

		assert.Equal(t, i+1, len(signer.GetKeyForSigIDCalls()))
		assert.Equal(t, sigID, signer.GetKeyForSigIDCalls()[i].SigID)
	}
}

func prepareMsgTransferToNewMasterKey(ctx sdk.Context, k keeper.Keeper, signer *btcMock.SignerMock, b *btcMock.BalancerMock, rpc *btcMock.RPCClientMock, sk *ecdsa.PrivateKey, sigID string) (types.MsgRawTx, types.MsgSendTx) {
	hash, err := chainhash.NewHash([]byte(testutils.RandString(chainhash.HashSize)))
	if err != nil {
		panic(err)
	}

	txID := hash.String()
	btcPk := btcec.PublicKey(sk.PublicKey)
	addr, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(btcPk.SerializeCompressed()), rpc.Network().Params())
	if err != nil {
		panic(err)
	}
	amount := btcutil.Amount(testutils.RandIntBetween(1, 100000000))
	k.SetUnverifiedOutpoint(ctx, txID, types.OutPointInfo{
		OutPoint:  wire.NewOutPoint(hash, uint32(testutils.RandIntBetween(0, 10))),
		Amount:    amount,
		Recipient: types.BtcAddress{Network: rpc.Network(), EncodedString: addr.EncodeAddress()},
	})
	_ = k.ProcessVerificationResult(ctx, txID, true)

	sender := sdk.AccAddress(testutils.RandString(int(testutils.RandIntBetween(5, 50))))

	query := keeper.NewQuerier(k, b, signer, rpc)
	bz, err := query(ctx, []string{keeper.QueryRawTx, txID, strconv.Itoa(int(amount)) + denom.Sat}, abci.RequestQuery{})
	if err != nil {
		panic(err)
	}
	var rawTx *wire.MsgTx
	testutils.Codec().MustUnmarshalJSON(bz, &rawTx)
	msgRawTx := types.NewMsgRawTx(sender, txID, rawTx)

	transfer := types.NewMsgSendTx(sender, txID, sigID)
	return msgRawTx, transfer
}
