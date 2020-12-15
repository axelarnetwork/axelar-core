package mock

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ types.Voter = &TestVoter{}

type TestVoter struct {
	InitPollCalled      bool
	Poll                exported.PollMeta
	VoteCalledCorrectly bool
	RecordedVote        exported.MsgVote
	ResultMock          func(sdk.Context, exported.PollMeta) exported.VotingData
}

func (t *TestVoter) InitPoll(_ sdk.Context, poll exported.PollMeta) error {
	t.InitPollCalled = true
	t.Poll = poll
	return nil
}

func (t *TestVoter) Vote(_ sdk.Context, vote exported.MsgVote) error {
	t.VoteCalledCorrectly = t.Poll.String() == vote.Poll().String()
	t.RecordedVote = vote
	return nil
}

func (t *TestVoter) TallyVote(_ sdk.Context, _ exported.MsgVote) error {
	panic("implement me")
}

func (t *TestVoter) Result(ctx sdk.Context, poll exported.PollMeta) exported.VotingData {
	return t.ResultMock(ctx, poll)
}

var _ types.RPCClient = &TestRPC{}

type TestRPC struct {
	TrackedAddress string ``
	Cancel         context.CancelFunc
	RawTxs         map[string]*btcjson.TxRawResult
}

func (t *TestRPC) ImportAddressRescan(address string, _ string, _ bool) error {
	t.TrackedAddress = address
	t.Cancel()
	return nil
}

func (t *TestRPC) ImportAddress(address string) error {
	t.TrackedAddress = address
	t.Cancel()
	return nil
}

func (t TestRPC) GetRawTransactionVerbose(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
	if res, ok := t.RawTxs[hash.String()]; !ok {
		return nil, fmt.Errorf("tx not found")
	} else {
		return res, nil
	}
}

func (t TestRPC) SendRawTransaction(tx *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
	hash := tx.TxHash()
	return &hash, nil
}

type TestSigner struct {
	GetCurrentMasterKeyMock func(sdk.Context, string) (ecdsa.PublicKey, bool)
	GetNextMasterKeyMock    func(sdk.Context, string) (ecdsa.PublicKey, bool)
	GetSigMock              func(sdk.Context, string) (tss.Signature, bool)
	GetKeyMock              func(sdk.Context, string) (ecdsa.PublicKey, bool)
}

func (t TestSigner) GetSig(ctx sdk.Context, sigID string) (tss.Signature, bool) {
	return t.GetSigMock(ctx, sigID)
}

func (t TestSigner) GetKey(ctx sdk.Context, keyID string) (ecdsa.PublicKey, bool) {
	return t.GetKeyMock(ctx, keyID)
}

func (t TestSigner) GetCurrentMasterKey(ctx sdk.Context, chain string) (ecdsa.PublicKey, bool) {
	return t.GetCurrentMasterKeyMock(ctx, chain)
}

func (t TestSigner) GetNextMasterKey(ctx sdk.Context, chain string) (ecdsa.PublicKey, bool) {
	return t.GetNextMasterKeyMock(ctx, chain)
}
