package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestNewTalliedVote(t *testing.T) {
	t.Run("panic on nil data", func(t *testing.T) {
		assert.Panics(t, func() {
			types.NewTalliedVote(exported.PollID(0), nil)
		})
	})

	t.Run("should return a new tallied vote", func(t *testing.T) {
		pollID := exported.PollID(1)
		actual := types.NewTalliedVote(pollID, &gogoprototypes.StringValue{Value: rand.Str(5)})

		assert.Equal(t, pollID, actual.PollID)
		assert.Equal(t, sdk.ZeroUint(), actual.Tally)
		assert.NotNil(t, actual.Data)
		assert.Nil(t, actual.IsVoterLate)
	})
}

func TestTalliedVote(t *testing.T) {
	var (
		talliedVote types.TalliedVote
		vote        codec.ProtoMarshaler
	)

	encCfg := app.MakeEncodingConfig()
	encCfg.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &gogoprototypes.StringValue{})
	cdc := encCfg.Codec

	givenTalliedVote := Given("given a tallied vote", func() {
		vote = &gogoprototypes.StringValue{Value: rand.Str(5)}
		talliedVote = types.NewTalliedVote(exported.PollID(1), vote)
		talliedVote.IsVoterLate = map[string]bool{
			rand.ValAddr().String(): true,
			rand.ValAddr().String(): false,
			rand.ValAddr().String(): true,
			rand.ValAddr().String(): false,
			rand.ValAddr().String(): true,
		}
	})

	t.Run("marshaling", func(t *testing.T) {
		var (
			bz []byte
		)

		givenTalliedVote.
			When("marshalized", func() {
				bz = cdc.MustMarshalLengthPrefixed(&talliedVote)
			}).
			Then("should marshalize to the same bytes", func(t *testing.T) {
				for i := 0; i < 100; i++ {
					assert.Equal(t, bz, cdc.MustMarshalLengthPrefixed(&talliedVote))
				}
			}).
			Run(t, 20)

		givenTalliedVote.
			When("marshalized", func() {
				bz = cdc.MustMarshalLengthPrefixed(&talliedVote)
			}).
			Then("should be able to unmarshalize", func(t *testing.T) {
				var actual types.TalliedVote
				cdc.MustUnmarshalLengthPrefixed(bz, &actual)

				assert.Equal(t, talliedVote, actual)
				assert.NotNil(t, actual.Data.GetCachedValue())
				assert.Equal(t, vote, actual.Data.GetCachedValue())
			}).
			Run(t, 20)
	})

	t.Run("TallyVote", func(t *testing.T) {
		whenTalliedVoteIsNew := When("when tallied vote is new", func() {
			talliedVote.Tally = sdk.ZeroUint()
			talliedVote.IsVoterLate = nil
		})

		givenTalliedVote.
			When2(whenTalliedVoteIsNew).
			Then("should panic if given voter is nil", func(t *testing.T) {
				assert.Panics(t, func() {
					talliedVote.TallyVote(nil, sdk.OneUint(), true)
				})
			}).
			Run(t)

		givenTalliedVote.
			When2(whenTalliedVoteIsNew).
			Then("should tally the vote", func(t *testing.T) {
				voter := rand.ValAddr()
				talliedVote.TallyVote(voter, sdk.OneUint(), true)

				assert.True(t, talliedVote.IsVoterLate[voter.String()])
				assert.Equal(t, sdk.OneUint(), talliedVote.Tally)
				assert.NoError(t, talliedVote.ValidateBasic())
			}).
			Run(t)
	})
}
