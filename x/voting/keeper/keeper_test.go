package keeper

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/store"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/mock"
	"github.com/axelarnetwork/axelar-core/utils"
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	stExported "github.com/axelarnetwork/axelar-core/x/staking/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/types"
)

var (
	// default test data, each tests adds modifications as necessary
	poll1 = exported.PollMeta{Module: "testModule", Type: "testType", ID: "poll1"}
	poll2 = exported.PollMeta{Module: "testModule", Type: "testType", ID: "poll2"}
	poll3 = exported.PollMeta{Module: "otherModule", Type: "otherType", ID: "poll3"}

	voteForPoll1 = &mockVote{Path: poll1.Module, MsgType: poll1.Type, PollMeta: poll1, VotingData: "poll1 data"}
	voteForPoll2 = &mockVote{Path: poll2.Module, MsgType: poll2.Type, PollMeta: poll2, VotingData: "poll2 data"}
	voteForPoll3 = &mockVote{Path: poll3.Module, MsgType: poll3.Type, PollMeta: poll3, VotingData: "poll3 data"}
)

func init() {
	cdc := testutils.Codec()
	cdc.RegisterConcrete(&mockVote{}, "mockVote", nil)
	cdc.RegisterConcrete("", "string", nil)
}

func newKeeper(b bcExported.Broadcaster, validators ...stExported.Validator) Keeper {
	return NewKeeper(testutils.Codec(), mock.NewKVStoreKey("voting"), store.NewSubjectiveStore(), mock.NewTestStaker(validators...), b)
}

func newBroadcaster() (mock.Broadcaster, <-chan sdk.Msg) {
	out := make(chan sdk.Msg, 10)
	b := mock.NewBroadcaster(testutils.Codec(), sdk.AccAddress("sender"), sdk.ValAddress("validator"), out)
	return b, out
}

func TestKeeper(t *testing.T) {
	t.Run("no error on initializing new poll", func(t *testing.T) { initPollNoError(t) })
	t.Run("error when initializing poll with same id as existing poll", func(t *testing.T) { initPollSameIdReturnError(t) })

	t.Run("vote on []byte data", func(t *testing.T) { voteOnBytes(t) })

	t.Run("error when voting for unknown poll, no polls initialized", func(t *testing.T) { voteNoPollsReturnError(t) })
	t.Run("error when voting where poll id matches none of the existing polls", func(t *testing.T) { votePollIdMismatchReturnError(t) })
	t.Run("vote for existing poll is added to next ballot, exactly one message", func(t *testing.T) { voteOnNextBallot(t) })
	t.Run("votes that are part of one ballot will not be added to consecutive ballots", func(t *testing.T) { votesNotRepeatedInConsecutiveBallots(t) })
	t.Run("error when voting on same poll multiple times", func(t *testing.T) { voteMultipleTimesReturnError(t) })
	t.Run("send no ballot when there are no votes", func(t *testing.T) { noVotesNoBallot(t) })

	t.Run("error when tallying non-existing poll", func(t *testing.T) { tallyNonExistingPollReturnError(t) })
	t.Run("error when tallied vote comes from unauthorized voter", func(t *testing.T) { tallyUnknownVoterReturnError(t) })
	t.Run("error when tallying second vote from same validator", func(t *testing.T) { tallyTwoVotesFromSameValidatorReturnError(t) })
	t.Run("tally vote no winner", func(t *testing.T) { tallyNoWinner(t) })
	t.Run("tally vote with winner", func(t *testing.T) { tallyWithWinner(t) })
	t.Run("tally multiple votes until poll is decided", func(t *testing.T) { tallyMultipleVotesUntilDecision(t) })
	t.Run("tally vote for already decided vote", func(t *testing.T) { tallyForDecidedPoll(t) })
}

func initPollNoError(t *testing.T) {
	k := newKeeper(nil)
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	assert.NoError(t, k.InitPoll(ctx, poll1))
}

func initPollSameIdReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(nil)

	assert.NoError(t, k.InitPoll(ctx, poll1))
	assert.Error(t, k.InitPoll(ctx, poll1))
}

func voteOnBytes(t *testing.T) {
	b, msgs := newBroadcaster()
	k := newKeeper(b)
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	v1 := voteForPoll1
	v1.VotingData = []byte("some test data")

	assert.NoError(t, k.InitPoll(ctx, poll1))
	assert.NoError(t, k.Vote(ctx, v1))
	k.SendBallot(ctx)

	timeout, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	select {
	case <-timeout.Done():
		assert.FailNow(t, "no message received")
	case m := <-msgs:
		assert.IsType(t, types.MsgBallot{}, m)
		assert.Len(t, m.(types.MsgBallot).Votes, 1)
		assert.Equal(t, v1.VotingData, m.(types.MsgBallot).Votes[0].Data())
	}
}

func voteNoPollsReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(nil)

	assert.Error(t, k.Vote(ctx, voteForPoll1))
}

func votePollIdMismatchReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(nil)

	assert.NoError(t, k.InitPoll(ctx, poll1))

	assert.Error(t, k.Vote(ctx, voteForPoll2))
}

func voteOnNextBallot(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, out := newBroadcaster()
	k := newKeeper(b)

	assert.NoError(t, k.InitPoll(ctx, poll1))
	assert.NoError(t, k.Vote(ctx, voteForPoll1))

	k.SendBallot(ctx)

	// assert that exactly one ballot with one vote is sent out
	timeout, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	msgCount := 0
loop:
	for {
		select {
		case <-timeout.Done():
			break loop
		case msg := <-out:
			if msgCount == 0 {
				msgCount += 1
				assert.IsType(t, types.MsgBallot{}, msg)
				assert.Len(t, msg.(types.MsgBallot).Votes, 1)
				assert.Equal(t, voteForPoll1, msg.(types.MsgBallot).Votes[0])
			} else {
				assert.FailNow(t, "broadcasting multiple messages")
			}
		}
	}
	if msgCount < 1 {
		assert.FailNow(t, "broadcast timed out before receiving the ballot")
	}
}

func votesNotRepeatedInConsecutiveBallots(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, out := newBroadcaster()
	k := newKeeper(b)

	assert.NoError(t, k.InitPoll(ctx, poll1))
	assert.NoError(t, k.InitPoll(ctx, poll2))
	assert.NoError(t, k.InitPoll(ctx, poll3))

	assert.NoError(t, k.Vote(ctx, voteForPoll1))
	assert.NoError(t, k.Vote(ctx, voteForPoll2))

	k.SendBallot(ctx)

	// make sure that the first message is sent out before the second
	time.Sleep(10 * time.Millisecond)

	assert.NoError(t, k.Vote(ctx, voteForPoll3))

	k.SendBallot(ctx)

	// assert the votes are batched according to the timing of the SendBallot call
	timeout, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	msgCount := 0
loop:
	for {
		select {
		case <-timeout.Done():
			break loop
		case msg := <-out:
			assert.IsType(t, types.MsgBallot{}, msg)
			switch msgCount {
			case 0:
				assert.Len(t, msg.(types.MsgBallot).Votes, 2)
				assert.Equal(t, voteForPoll1, msg.(types.MsgBallot).Votes[0])
				assert.Equal(t, voteForPoll2, msg.(types.MsgBallot).Votes[1])
			case 1:
				assert.Len(t, msg.(types.MsgBallot).Votes, 1)
				assert.Equal(t, voteForPoll3, msg.(types.MsgBallot).Votes[0])
			default:
				assert.FailNow(t, "broadcasting too many messages")
			}
			msgCount += 1
		}
	}
	if msgCount < 2 {
		assert.FailNow(t, "broadcast timed out before receiving both ballots")
	}
}

func voteMultipleTimesReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	k := newKeeper(b)

	assert.NoError(t, k.InitPoll(ctx, poll1))
	assert.NoError(t, k.Vote(ctx, voteForPoll1))
	v2 := *voteForPoll1
	v2.VotingData = "different data"
	v3 := v2
	v3.VotingData = "even more different data"

	assert.Error(t, k.Vote(ctx, &v2))
	assert.Error(t, k.Vote(ctx, &v3))
}

func noVotesNoBallot(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, out := newBroadcaster()
	k := newKeeper(b)

	assert.NoError(t, k.InitPoll(ctx, poll1))
	assert.NoError(t, k.InitPoll(ctx, poll2))
	assert.NoError(t, k.InitPoll(ctx, poll3))

	k.SendBallot(ctx)

	timeout, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	select {
	case <-timeout.Done():
		break
	case <-out:
		assert.FailNow(t, "should not receive any messages")
	}
}

func tallyNonExistingPollReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	k := newKeeper(b, stExported.Validator{Address: b.GetLocalPrincipal(ctx), Power: 10})
	k.SetVotingThreshold(ctx, utils.Threshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v2 := *voteForPoll2
	v2.SetSender(b.Proxy)
	err := k.TallyVote(ctx, &v2)
	assert.Error(t, err)
}

func tallyUnknownVoterReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	k := newKeeper(b, stExported.Validator{Address: b.GetLocalPrincipal(ctx), Power: 10})
	k.SetVotingThreshold(ctx, utils.Threshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v1.SetSender(sdk.AccAddress("some other proxy"))
	err := k.TallyVote(ctx, &v1)
	assert.Error(t, err)
}

func tallyNoWinner(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	k := newKeeper(
		b,
		stExported.Validator{Address: b.GetLocalPrincipal(ctx), Power: 10},
		stExported.Validator{Address: sdk.ValAddress("big spender"), Power: 90},
	)
	k.SetVotingThreshold(ctx, utils.Threshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v1.SetSender(b.Proxy)
	err := k.TallyVote(ctx, &v1)
	res := k.Result(ctx, v1.PollMeta)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func tallyWithWinner(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	k := newKeeper(
		b,
		stExported.Validator{Address: b.GetLocalPrincipal(ctx), Power: 90},
		stExported.Validator{Address: sdk.ValAddress("small fish"), Power: 10},
	)
	k.SetVotingThreshold(ctx, utils.Threshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v1.SetSender(b.Proxy)
	err := k.TallyVote(ctx, &v1)
	res := k.Result(ctx, v1.PollMeta)
	assert.NoError(t, err)
	assert.Equal(t, voteForPoll1.VotingData, res.Data())
}

func tallyTwoVotesFromSameValidatorReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	validator := stExported.Validator{Address: b.GetLocalPrincipal(ctx), Power: 10}
	k := newKeeper(b, validator)
	k.SetVotingThreshold(ctx, utils.Threshold{Numerator: 2, Denominator: 3})

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v2 := v1
	v3 := v1

	v1.SetSender(b.Proxy)

	// different decision
	v2.SetSender(b.Proxy)
	v2.VotingData = "different data"

	// different data
	v3.SetSender(b.Proxy)
	v3.VotingData = "even more different data"

	assert.NoError(t, k.InitPoll(ctx, poll1))

	err := k.TallyVote(ctx, &v1)
	assert.NoError(t, err)

	err = k.TallyVote(ctx, &v2)
	assert.Error(t, err)

	err = k.TallyVote(ctx, &v3)
	assert.Error(t, err)
}

func tallyMultipleVotesUntilDecision(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	proxy2 := sdk.AccAddress("proxy2")
	proxy3 := sdk.AccAddress("proxy3")
	val1 := stExported.Validator{Address: b.GetLocalPrincipal(ctx), Power: 10}
	val2 := stExported.Validator{Address: sdk.ValAddress("val2"), Power: 10}
	val3 := stExported.Validator{Address: sdk.ValAddress("val3"), Power: 7}
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	assert.NoError(t, b.RegisterProxy(ctx, val2.Address, proxy2))
	assert.NoError(t, b.RegisterProxy(ctx, val3.Address, proxy3))

	k := newKeeper(b, val1, val2, val3)
	k.SetVotingThreshold(ctx, utils.Threshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v2 := v1
	v3 := v1

	v1.SetSender(b.Proxy)
	v2.SetSender(proxy2)
	v3.SetSender(proxy3)

	v3.VotingData = "different data"

	err := k.TallyVote(ctx, &v1)
	res := k.Result(ctx, v1.PollMeta)
	assert.NoError(t, err)
	assert.Nil(t, res)

	err = k.TallyVote(ctx, &v3)
	res = k.Result(ctx, v3.PollMeta)
	assert.NoError(t, err)
	assert.Nil(t, res)

	err = k.TallyVote(ctx, &v2)
	res = k.Result(ctx, v2.PollMeta)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, v1.VotingData, res.Data())
}

func tallyForDecidedPoll(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	proxy2 := sdk.AccAddress("proxy2")
	proxy3 := sdk.AccAddress("proxy3")
	val1 := stExported.Validator{Address: b.GetLocalPrincipal(ctx), Power: 10}
	val2 := stExported.Validator{Address: sdk.ValAddress("val2"), Power: 10}
	val3 := stExported.Validator{Address: sdk.ValAddress("val3"), Power: 7}
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	assert.NoError(t, b.RegisterProxy(ctx, val2.Address, proxy2))
	assert.NoError(t, b.RegisterProxy(ctx, val3.Address, proxy3))

	k := newKeeper(b, val1, val2, val3)
	k.SetVotingThreshold(ctx, utils.Threshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v2 := v1
	v3 := v1

	v1.SetSender(b.Proxy)
	v2.SetSender(proxy2)
	v3.SetSender(proxy3)

	v3.VotingData = "different data"

	err := k.TallyVote(ctx, &v1)
	res := k.Result(ctx, v1.PollMeta)
	assert.NoError(t, err)
	assert.Nil(t, res)

	err = k.TallyVote(ctx, &v2)
	res = k.Result(ctx, v2.PollMeta)
	assert.NoError(t, err)
	assert.Equal(t, v1.VotingData, res.Data())

	err = k.TallyVote(ctx, &v3)
	res = k.Result(ctx, v3.PollMeta)
	assert.NoError(t, err)
	assert.Equal(t, v1.VotingData, res.Data())
}

type mockVote struct {
	Path       string
	MsgType    string
	PollMeta   exported.PollMeta
	VotingData exported.VotingData
	sender     sdk.AccAddress
}

func (m *mockVote) SetSender(address sdk.AccAddress) {
	m.sender = address
}

func (m mockVote) Route() string {
	return m.Path
}

func (m mockVote) Type() string {
	return m.MsgType
}

func (m mockVote) ValidateBasic() error {
	return nil
}

func (m mockVote) GetSignBytes() []byte {
	return nil
}

func (m mockVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.sender}
}

func (m mockVote) Poll() exported.PollMeta {
	return m.PollMeta
}

func (m mockVote) Data() exported.VotingData {
	return m.VotingData
}
