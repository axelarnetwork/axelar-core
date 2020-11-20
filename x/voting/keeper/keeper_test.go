package keeper

import (
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/store"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/mock"
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/types"
)

var (
	once = sync.Once{}

	// test data
	poll1 = exported.PollMeta{Module: "testModule", Type: "testType", ID: "poll1"}
	poll2 = exported.PollMeta{Module: "testModule", Type: "testType", ID: "poll2"}
	poll3 = exported.PollMeta{Module: "otherModule", Type: "otherType", ID: "poll3"}

	voteForPoll1 = &mockVote{R: poll1.Module, T: poll1.Type, P: poll1, D: "poll1 data", C: true}
	voteForPoll2 = &mockVote{R: poll2.Module, T: poll2.Type, P: poll2, D: "poll2 data", C: false}
	voteForPoll3 = &mockVote{R: poll3.Module, T: poll3.Type, P: poll3, D: "poll3 data", C: false}
)

func cdc() *codec.Codec {
	cdc := testutils.Codec()
	once.Do(func() {
		cdc.RegisterConcrete(&mockVote{}, "mockVote", nil)
		cdc.RegisterConcrete("", "string", nil)
	})
	return cdc
}

func newKeeper(b bcExported.Broadcaster, validators ...staking.ValidatorI) Keeper {
	return NewKeeper(cdc(), mock.NewKVStoreKey("voting"), store.NewSubjectiveStore(), mock.NewTestStaker(validators...), b)
}

func newBroadcaster() (mock.Broadcaster, <-chan sdk.Msg) {
	out := make(chan sdk.Msg, 10)
	b := mock.NewBroadcaster(cdc(), sdk.AccAddress("sender"), sdk.ValAddress("validator"), out)
	return b, out
}

func TestKeeper(t *testing.T) {
	t.Run("no error on initializing new poll", func(t *testing.T) { initPollNoError(t) })
	t.Run("error when initializing poll with same id as existing poll", func(t *testing.T) { initPollSameIdReturnError(t) })
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
	timeout := testutils.StartTimeout(100 * time.Millisecond)
	msgCount := 0
loop:
	for {
		select {
		case <-timeout:
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
	timeout := testutils.StartTimeout(100 * time.Millisecond)
	msgCount := 0
loop:
	for {
		select {
		case <-timeout:
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
	v2.C = false
	v3 := v2
	v3.D = "different data"

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

	timeout := testutils.StartTimeout(100 * time.Millisecond)
	select {
	case <-timeout:
		break
	case <-out:
		assert.FailNow(t, "should not receive any messages")
	}
}

func tallyNonExistingPollReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	k := newKeeper(b, mock.NewTestValidator(b.GetLocalPrincipal(ctx), 10))
	k.SetVotingThreshold(ctx, types.VotingThreshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v2 := *voteForPoll2
	v2.SetSender(b.Proxy)
	_, err := k.TallyVote(ctx, &v2)
	assert.Error(t, err)
}

func tallyUnknownVoterReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	k := newKeeper(b, mock.NewTestValidator(b.GetLocalPrincipal(ctx), 10))
	k.SetVotingThreshold(ctx, types.VotingThreshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v1.SetSender(sdk.AccAddress("some other proxy"))
	_, err := k.TallyVote(ctx, &v1)
	assert.Error(t, err)
}

func tallyNoWinner(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	k := newKeeper(
		b,
		mock.NewTestValidator(b.GetLocalPrincipal(ctx), 10),
		mock.NewTestValidator(sdk.ValAddress("big spender"), 90),
	)
	k.SetVotingThreshold(ctx, types.VotingThreshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v1.SetSender(b.Proxy)
	data, err := k.TallyVote(ctx, &v1)
	assert.NoError(t, err)
	assert.Nil(t, data)
}

func tallyWithWinner(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	k := newKeeper(
		b,
		mock.NewTestValidator(b.GetLocalPrincipal(ctx), 90),
		mock.NewTestValidator(sdk.ValAddress("small fish"), 10),
	)
	k.SetVotingThreshold(ctx, types.VotingThreshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v1.SetSender(b.Proxy)
	res, err := k.TallyVote(ctx, &v1)
	assert.NoError(t, err)
	assert.Equal(t, voteForPoll1.D, res.Data())
	assert.Equal(t, voteForPoll1.C, res.Confirms())
}

func tallyTwoVotesFromSameValidatorReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	validator := mock.NewTestValidator(b.GetLocalPrincipal(ctx), 10)
	k := newKeeper(b, validator)
	k.SetVotingThreshold(ctx, types.VotingThreshold{Numerator: 2, Denominator: 3})

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v2 := v1
	v3 := v1

	v1.SetSender(b.Proxy)

	// different decision
	v2.SetSender(b.Proxy)
	v2.C = !v2.C

	// different data
	v3.SetSender(b.Proxy)
	v3.D = "unique data"

	assert.NoError(t, k.InitPoll(ctx, poll1))

	_, err := k.TallyVote(ctx, &v1)
	assert.NoError(t, err)

	_, err = k.TallyVote(ctx, &v2)
	assert.Error(t, err)

	_, err = k.TallyVote(ctx, &v3)
	assert.Error(t, err)
}

func tallyMultipleVotesUntilDecision(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	proxy2 := sdk.AccAddress("proxy2")
	proxy3 := sdk.AccAddress("proxy3")
	val1 := mock.NewTestValidator(b.GetLocalPrincipal(ctx), 10)
	val2 := mock.NewTestValidator(sdk.ValAddress("val2"), 10)
	val3 := mock.NewTestValidator(sdk.ValAddress("val3"), 7)
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	assert.NoError(t, b.RegisterProxy(ctx, val2.GetOperator(), proxy2))
	assert.NoError(t, b.RegisterProxy(ctx, val3.GetOperator(), proxy3))

	k := newKeeper(b, val1, val2, val3)
	k.SetVotingThreshold(ctx, types.VotingThreshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v2 := v1
	v3 := v1

	v1.SetSender(b.Proxy)
	v2.SetSender(proxy2)
	v3.SetSender(proxy3)

	v3.C = false

	res, err := k.TallyVote(ctx, &v1)
	assert.NoError(t, err)
	assert.Nil(t, res)

	res, err = k.TallyVote(ctx, &v3)
	assert.NoError(t, err)
	assert.Nil(t, res)

	res, err = k.TallyVote(ctx, &v2)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, v1.D, res.Data())
	assert.Equal(t, v1.C, res.Confirms())
}

func tallyForDecidedPoll(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	b, _ := newBroadcaster()
	proxy2 := sdk.AccAddress("proxy2")
	proxy3 := sdk.AccAddress("proxy3")
	val1 := mock.NewTestValidator(b.GetLocalPrincipal(ctx), 10)
	val2 := mock.NewTestValidator(sdk.ValAddress("val2"), 10)
	val3 := mock.NewTestValidator(sdk.ValAddress("val3"), 7)
	assert.NoError(t, b.RegisterProxy(ctx, b.GetLocalPrincipal(ctx), b.Proxy))
	assert.NoError(t, b.RegisterProxy(ctx, val2.GetOperator(), proxy2))
	assert.NoError(t, b.RegisterProxy(ctx, val3.GetOperator(), proxy3))

	k := newKeeper(b, val1, val2, val3)
	k.SetVotingThreshold(ctx, types.VotingThreshold{Numerator: 2, Denominator: 3})

	assert.NoError(t, k.InitPoll(ctx, poll1))

	// copy to not overwrite defaults
	v1 := *voteForPoll1
	v2 := v1
	v3 := v1

	v1.SetSender(b.Proxy)
	v2.SetSender(proxy2)
	v3.SetSender(proxy3)

	v3.C = false

	res, err := k.TallyVote(ctx, &v1)
	assert.NoError(t, err)
	assert.Nil(t, res)

	res, err = k.TallyVote(ctx, &v2)
	assert.NoError(t, err)
	assert.Equal(t, v1.D, res.Data())
	assert.Equal(t, v1.C, res.Confirms())

	res, err = k.TallyVote(ctx, &v3)
	assert.NoError(t, err)
	assert.Equal(t, v1.D, res.Data())
	assert.Equal(t, v1.C, res.Confirms())
}

type mockVote struct {
	R      string
	T      string
	P      exported.PollMeta
	D      exported.VotingData
	C      bool
	sender sdk.AccAddress
}

func (m *mockVote) SetSender(address sdk.AccAddress) {
	m.sender = address
}

func (m mockVote) Route() string {
	return m.R
}

func (m mockVote) Type() string {
	return m.T
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
	return m.P
}

func (m mockVote) Data() exported.VotingData {
	return m.D
}

func (m mockVote) Confirms() bool {
	return m.C
}
