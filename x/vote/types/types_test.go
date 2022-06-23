package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types/mock"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestNewTalliedVote(t *testing.T) {
	t.Run("panic on nil data", func(t *testing.T) {
		assert.Panics(t, func() {
			types.NewTalliedVote(rand.ValAddr(), rand.PosI64(), nil)
		})
	})

	t.Run("panic on nil voter", func(t *testing.T) {
		assert.Panics(t, func() {
			types.NewTalliedVote(nil, rand.PosI64(), &gogoprototypes.BoolValue{Value: true})
		})
	})
}

func TestTalliedVote_Marshaling(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	cdc := encCfg.Codec

	output := tofnd.KeygenOutput{PubKey: []byte("a public key"), GroupRecoverInfo: []byte{0}, PrivateRecoverInfo: []byte{0, 1, 2, 3}}
	data := tofnd.MessageOut_KeygenResult{KeygenResultData: &tofnd.MessageOut_KeygenResult_Data{Data: &output}}
	vote := types.NewTalliedVote(rand.ValAddr(), 23, &data)

	bz := cdc.MustMarshalLengthPrefixed(&vote)
	var actual types.TalliedVote
	cdc.MustUnmarshalLengthPrefixed(bz, &actual)

	assert.Equal(t, vote, actual)

	bz = cdc.MustMarshalJSON(&vote)
	var actual2 types.TalliedVote
	cdc.MustUnmarshalJSON(bz, &actual2)

	assert.Equal(t, vote.Tally, actual2.Tally)
	assert.Equal(t, vote.Data.GetCachedValue(), actual2.Data.GetCachedValue())
}

func TestPoll_Is(t *testing.T) {
	for _, state := range exported.PollState_value {
		poll := types.Poll{PollMetadata: exported.PollMetadata{State: exported.PollState(state)}}

		assert.True(t, poll.Is(exported.PollState(state)))

		for _, otherState := range exported.PollState_value {
			if otherState == state {
				continue
			}
			assert.False(t, poll.Is(exported.PollState(otherState)), "poll: %s, other: %s", poll.State, exported.PollState(otherState))
		}
	}
}

func TestPoll(t *testing.T) {
	var (
		poll      *types.Poll
		pollStore *mock.StoreMock
	)

	repeats := 20

	givenPoll := Given("poll", func() {
		voterCount := rand.I64Between(10, 20)
		voters := make([]exported.Voter, voterCount)
		for i := 0; i < int(voterCount); i++ {
			voters[i] = exported.Voter{
				Validator:   rand.ValAddr(),
				VotingPower: rand.I64Between(10, 100),
			}
		}

		pollID := exported.PollID(rand.I64Between(10, 100))
		pollMetadata := types.NewPollMetaData(pollID, types.DefaultParams().DefaultVotingThreshold, voters)

		pollStore = &mock.StoreMock{}
		poll = types.NewPoll(pollMetadata, pollStore)
	})

	withState := func(state exported.PollState) func() {
		return func() {
			poll.State = state
		}
	}

	t.Run("Vote", testutils.Func(func(t *testing.T) {
		givenPoll.
			When("poll does not exist", withState(exported.NonExistent)).
			Then("should return error", func(t *testing.T) {
				result, voted, err := poll.Vote(
					rand.Of(poll.Voters...).Validator,
					rand.PosI64(),
					&gogoprototypes.BoolValue{Value: true},
				)

				assert.Nil(t, result)
				assert.False(t, voted)
				assert.ErrorContains(t, err, "poll does not exist")
			}).
			Run(t)

		givenPoll.
			When("poll is in any state", withState(rand.Of(exported.Completed, exported.Failed, exported.Pending))).
			When("the voter has voted already", func() {
				pollStore.HasVotedFunc = func(v sdk.ValAddress) bool { return true }
			}).
			Then("should return error", func(t *testing.T) {
				result, voted, err := poll.Vote(
					rand.Of(poll.Voters...).Validator,
					rand.PosI64(),
					&gogoprototypes.BoolValue{Value: true},
				)

				assert.Nil(t, result)
				assert.False(t, voted)
				assert.ErrorContains(t, err, "has voted already")
			}).
			Run(t)

		givenPoll.
			When("poll is failed", withState(exported.Failed)).
			Then("should do nothing", func(t *testing.T) {
				pollStore.HasVotedFunc = func(v sdk.ValAddress) bool { return false }

				result, voted, err := poll.Vote(
					rand.Of(poll.Voters...).Validator,
					rand.PosI64(),
					&gogoprototypes.BoolValue{Value: true},
				)

				assert.Nil(t, result)
				assert.False(t, voted)
				assert.NoError(t, err)
			}).
			Run(t)

		blockHeight := rand.I64Between(100, 1000)
		givenPoll.
			When("poll is completed", withState(exported.Completed)).
			When("poll is within its grace period", func() {
				poll.GracePeriod = rand.I64Between(1, blockHeight)
				poll.CompletedAt = rand.I64Between(blockHeight-int64(poll.GracePeriod), blockHeight+1)
			}).
			Then("should allow late vote", func(t *testing.T) {
				voter := rand.Of(poll.Voters...)
				pollStore.HasVotedFunc = func(v sdk.ValAddress) bool { return !v.Equals(voter.Validator) }
				pollStore.SetVoteFunc = func(voter sdk.ValAddress, data codec.ProtoMarshaler, votingPower int64, isLate bool) {}

				result, voted, err := poll.Vote(
					voter.Validator,
					blockHeight,
					&gogoprototypes.BoolValue{Value: true},
				)

				assert.Nil(t, result)
				assert.True(t, voted)
				assert.NoError(t, err)

				assert.Len(t, pollStore.SetVoteCalls(), 1)
				assert.Equal(t, voter.Validator, pollStore.SetVoteCalls()[0].Voter)
				assert.NotNil(t, pollStore.SetVoteCalls()[0].Data)
				assert.Equal(t, voter.VotingPower, pollStore.SetVoteCalls()[0].VotingPower)
				assert.True(t, pollStore.SetVoteCalls()[0].IsLate)
			}).
			Run(t)

		blockHeight = rand.I64Between(100, 1000)
		givenPoll.
			When("poll is completed", withState(exported.Completed)).
			When("poll is not within its grace period", func() {
				poll.GracePeriod = rand.I64Between(1, blockHeight)
				poll.CompletedAt = rand.I64Between(0, blockHeight-int64(poll.GracePeriod))
			}).
			Then("should not allow late vote", func(t *testing.T) {
				voter := rand.Of(poll.Voters...)
				pollStore.HasVotedFunc = func(v sdk.ValAddress) bool { return !v.Equals(voter.Validator) }

				result, voted, err := poll.Vote(
					voter.Validator,
					blockHeight,
					&gogoprototypes.BoolValue{Value: true},
				)

				assert.Nil(t, result)
				assert.False(t, voted)
				assert.NoError(t, err)
			}).
			Run(t)

		var voter exported.Voter
		givenPoll.
			When("poll is pending", withState(exported.Pending)).
			When("voter is not eligible", func() {
				pollStore.HasVotedFunc = func(v sdk.ValAddress) bool { return false }
				voter = exported.Voter{Validator: rand.ValAddr()}
			}).
			Then("should return error", func(t *testing.T) {
				result, voted, err := poll.Vote(
					voter.Validator,
					rand.PosI64(),
					&gogoprototypes.BoolValue{Value: true},
				)

				assert.Nil(t, result)
				assert.False(t, voted)
				assert.ErrorContains(t, err, "is not eligible")
			}).
			Run(t)

		givenPoll.
			When("poll is completed", withState(exported.Completed)).
			When("voter is not eligible", func() {
				poll.GracePeriod = rand.I64Between(1, blockHeight)
				poll.CompletedAt = rand.I64Between(0, blockHeight-poll.GracePeriod)
				pollStore.HasVotedFunc = func(v sdk.ValAddress) bool { return false }
				voter = exported.Voter{Validator: rand.ValAddr()}
			}).
			Then("should return error", func(t *testing.T) {
				result, voted, err := poll.Vote(
					voter.Validator,
					rand.PosI64(),
					&gogoprototypes.BoolValue{Value: true},
				)

				assert.Nil(t, result)
				assert.False(t, voted)
				assert.ErrorContains(t, err, "is not eligible")
			}).
			Run(t)

		data := gogoprototypes.StringValue{Value: rand.Str(5)}
		result, _ := codectypes.NewAnyWithValue(&data)
		givenPoll.
			When("poll is pending", withState(exported.Pending)).
			When("enough votes have been received", func() {
				voter = rand.Of(poll.Voters...)
				pollStore.HasVotedFunc = func(v sdk.ValAddress) bool { return !v.Equals(voter.Validator) }
				pollStore.GetVotesFunc = func() []types.TalliedVote {
					return []types.TalliedVote{
						{
							Tally:  poll.TotalVotingPower.MulRaw(poll.VotingThreshold.Numerator).QuoRaw(poll.VotingThreshold.Denominator).AddRaw(1),
							Voters: []sdk.ValAddress{voter.Validator},
							Data:   result,
						},
					}
				}
				pollStore.SetVoteFunc = func(voter sdk.ValAddress, data codec.ProtoMarshaler, votingPower int64, isLate bool) {}
				pollStore.SetMetadataFunc = func(metadata exported.PollMetadata) {}
			}).
			Then("should succeed voting", func(t *testing.T) {
				blockHeight := rand.PosI64()
				result, voted, err := poll.Vote(
					voter.Validator,
					blockHeight,
					&data,
				)

				assert.NotNil(t, result)
				assert.True(t, voted)
				assert.NoError(t, err)

				assert.Len(t, pollStore.SetVoteCalls(), 1)
				assert.Equal(t, voter.Validator, pollStore.SetVoteCalls()[0].Voter)
				assert.NotNil(t, pollStore.SetVoteCalls()[0].Data)
				assert.Equal(t, voter.VotingPower, pollStore.SetVoteCalls()[0].VotingPower)
				assert.False(t, pollStore.SetVoteCalls()[0].IsLate)
				assert.Len(t, pollStore.SetMetadataCalls(), 1)
				assert.NotNil(t, pollStore.SetMetadataCalls()[0].Metadata.Result)
				assert.Equal(t, exported.Completed, pollStore.SetMetadataCalls()[0].Metadata.State)
				assert.Equal(t, blockHeight, pollStore.SetMetadataCalls()[0].Metadata.CompletedAt)
			}).
			Run(t)

		givenPoll.
			When("poll is pending", withState(exported.Pending)).
			When("poll cannot possibly complete", func() {
				voter = rand.Of(poll.Voters...)
				pollStore.HasVotedFunc = func(v sdk.ValAddress) bool { return !v.Equals(voter.Validator) }
				pollStore.GetVotesFunc = func() []types.TalliedVote {
					return slices.Map(
						slices.Filter(poll.Voters, func(v exported.Voter) bool { return !v.Validator.Equals(voter.Validator) }),
						func(v exported.Voter) types.TalliedVote {
							data, _ := codectypes.NewAnyWithValue(&gogoprototypes.StringValue{Value: rand.Str(5)})

							return types.TalliedVote{
								Tally:  sdk.NewInt(v.VotingPower),
								Voters: []sdk.ValAddress{v.Validator},
								Data:   data,
							}
						},
					)
				}
				pollStore.SetVoteFunc = func(voter sdk.ValAddress, data codec.ProtoMarshaler, votingPower int64, isLate bool) {}
				pollStore.SetMetadataFunc = func(metadata exported.PollMetadata) {}
			}).
			Then("should succeed voting but poll is failed", func(t *testing.T) {
				result, voted, err := poll.Vote(
					voter.Validator,
					rand.PosI64(),
					&gogoprototypes.StringValue{Value: rand.Str(5)},
				)

				assert.Nil(t, result)
				assert.True(t, voted)
				assert.NoError(t, err)

				assert.Len(t, pollStore.SetVoteCalls(), 1)
				assert.Equal(t, voter.Validator, pollStore.SetVoteCalls()[0].Voter)
				assert.NotNil(t, pollStore.SetVoteCalls()[0].Data)
				assert.Equal(t, voter.VotingPower, pollStore.SetVoteCalls()[0].VotingPower)
				assert.False(t, pollStore.SetVoteCalls()[0].IsLate)
				assert.Len(t, pollStore.SetMetadataCalls(), 1)
				assert.Nil(t, pollStore.SetMetadataCalls()[0].Metadata.Result)
				assert.Equal(t, exported.Failed, pollStore.SetMetadataCalls()[0].Metadata.State)
				assert.EqualValues(t, 0, pollStore.SetMetadataCalls()[0].Metadata.CompletedAt)
			}).
			Run(t)

		givenPoll.
			When("poll is pending", withState(exported.Pending)).
			When("no voter has voted yet", func() {
				voter = rand.Of(poll.Voters...)
				pollStore.HasVotedFunc = func(v sdk.ValAddress) bool { return !v.Equals(voter.Validator) }
				pollStore.GetVotesFunc = func() []types.TalliedVote {
					data, _ := codectypes.NewAnyWithValue(&gogoprototypes.StringValue{Value: rand.Str(5)})

					return []types.TalliedVote{
						{
							Tally:  sdk.NewInt(voter.VotingPower),
							Data:   data,
							Voters: []sdk.ValAddress{voter.Validator},
						},
					}
				}
				pollStore.SetVoteFunc = func(voter sdk.ValAddress, data codec.ProtoMarshaler, votingPower int64, isLate bool) {}
			}).
			Then("should succeed voting", func(t *testing.T) {
				result, voted, err := poll.Vote(
					voter.Validator,
					rand.PosI64(),
					&gogoprototypes.StringValue{Value: rand.Str(5)},
				)

				assert.Nil(t, result)
				assert.True(t, voted)
				assert.NoError(t, err)

				assert.Len(t, pollStore.SetVoteCalls(), 1)
				assert.Equal(t, voter.Validator, pollStore.SetVoteCalls()[0].Voter)
				assert.NotNil(t, pollStore.SetVoteCalls()[0].Data)
				assert.Equal(t, voter.VotingPower, pollStore.SetVoteCalls()[0].VotingPower)
				assert.False(t, pollStore.SetVoteCalls()[0].IsLate)
			}).
			Run(t)
	}).Repeat(repeats))
}
