package vote

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilsMock "github.com/axelarnetwork/axelar-core/utils/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	exportedMock "github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func TestHandlePollsAtExpiry(t *testing.T) {
	var (
		keeper      *mock.VoterMock
		pollQueue   *utilsMock.KVQueueMock
		poll        *exportedMock.PollMock
		voteHandler *exportedMock.VoteHandlerMock
	)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger()).WithBlockHeight(rand.I64Between(10, 100))
	repeats := 20

	givenPollQueue := Given("poll queue", func() {
		pollQueue = &utilsMock.KVQueueMock{}
		voteHandler = &exportedMock.VoteHandlerMock{}
		poll = &exportedMock.PollMock{
			GetKeyFunc: func() exported.PollKey { return exported.NewPollKey(rand.Str(5), rand.HexStr(32)) },
		}
		keeper = &mock.VoterMock{
			LoggerFunc:       func(ctx sdk.Context) log.Logger { return log.NewNopLogger() },
			GetPollQueueFunc: func(ctx sdk.Context) utils.KVQueue { return pollQueue },
			GetPollFunc:      func(ctx sdk.Context, key exported.PollKey) exported.Poll { return poll },
			GetVoteRouterFunc: func() types.VoteRouter {
				return &mock.VoteRouterMock{
					GetHandlerFunc: func(module string) exported.VoteHandler { return voteHandler },
				}
			},
		}
	})

	withPoll := func(expired bool, state exported.PollState) WhenStatement {
		pollMetadata := exported.PollMetadata{State: state}
		if expired {
			pollMetadata.ExpiresAt = rand.I64Between(1, ctx.BlockHeight()+1)
		} else {
			pollMetadata.ExpiresAt = ctx.BlockHeight() + rand.I64Between(1, 10)
		}

		return When(fmt.Sprintf("having poll (state=%s,expired=%t)", state.String(), expired), func() {
			poll.IsFunc = func(s exported.PollState) bool { return state == s }

			dequeued := false
			pollQueue.DequeueIfFunc = func(value codec.ProtoMarshaler, filter func(value codec.ProtoMarshaler) bool) bool {
				if dequeued {
					return false
				}

				if !filter(&pollMetadata) {
					return false
				}

				bz, _ := pollMetadata.Marshal()
				if err := value.Unmarshal(bz); err != nil {
					panic(err)
				}
				dequeued = true

				return true
			}
		})
	}

	givenPollQueue.
		When2(withPoll(false, rand.Of(exported.Pending, exported.Completed, exported.Failed))).
		Then("should do nothing", func(t *testing.T) {
			err := handlePollsAtExpiry(ctx, keeper)
			assert.NoError(t, err)
		}).
		Run(t, repeats)

	givenPollQueue.
		When2(withPoll(true, exported.Pending)).
		Then("set poll as expired", func(t *testing.T) {
			poll.SetExpiredFunc = func() {}
			poll.AllowOverrideFunc = func() {}
			voteHandler.HandleExpiredPollFunc = func(ctx sdk.Context, poll exported.Poll) error { return nil }

			err := handlePollsAtExpiry(ctx, keeper)
			assert.NoError(t, err)

			assert.Len(t, poll.SetExpiredCalls(), 1)
			assert.Len(t, poll.AllowOverrideCalls(), 1)
			assert.Len(t, voteHandler.HandleExpiredPollCalls(), 1)
		}).
		Run(t, repeats)

	givenPollQueue.
		When2(withPoll(true, exported.Failed)).
		Then("set poll as allow override", func(t *testing.T) {
			poll.AllowOverrideFunc = func() {}

			err := handlePollsAtExpiry(ctx, keeper)
			assert.NoError(t, err)

			assert.Len(t, poll.AllowOverrideCalls(), 1)
		}).
		Run(t, repeats)

	givenPollQueue.
		When2(withPoll(true, exported.Completed)).
		Then("should handle it as completed", func(t *testing.T) {
			poll.GetResultFunc = func() codec.ProtoMarshaler { return &gogoprototypes.StringValue{} }
			voteHandler.HandleCompletedPollFunc = func(ctx sdk.Context, poll exported.Poll) error { return nil }

			isResultFalsy := rand.Bools(0.5).Next()
			if isResultFalsy {
				poll.AllowOverrideFunc = func() {}
			}
			voteHandler.IsFalsyResultFunc = func(result codec.ProtoMarshaler) bool { return isResultFalsy }

			err := handlePollsAtExpiry(ctx, keeper)
			assert.NoError(t, err)

			if isResultFalsy {
				assert.Len(t, poll.AllowOverrideCalls(), 1)
			} else {
				assert.Len(t, poll.AllowOverrideCalls(), 0)
			}
			assert.Len(t, voteHandler.HandleCompletedPollCalls(), 1)
		}).
		Run(t, repeats)
}
