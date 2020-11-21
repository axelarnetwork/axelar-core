package mock

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

var _ types.Voter = &TestVoter{}

type TestVoter struct {
	InitPollCalled      bool
	Poll                exported.PollMeta
	VoteCalledCorrectly bool
	RecordedVote        exported.MsgVote
}

func (t *TestVoter) InitPoll(_ sdkTypes.Context, poll exported.PollMeta) error {
	t.InitPollCalled = true
	t.Poll = poll
	return nil
}

func (t *TestVoter) Vote(_ sdkTypes.Context, vote exported.MsgVote) error {
	t.VoteCalledCorrectly = t.Poll.String() == vote.Poll().String()
	t.RecordedVote = vote
	return nil
}

func (t *TestVoter) TallyVote(_ sdkTypes.Context, _ exported.MsgVote) error {
	panic("implement me")
}

func (t *TestVoter) Result(_ sdkTypes.Context, _ exported.PollMeta) exported.Vote {
	panic("implement me")
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

func (t TestRPC) SendRawTransaction(_ *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
	panic("implement me")
}
