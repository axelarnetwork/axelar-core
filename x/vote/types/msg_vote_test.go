package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestVoteRequest_ValidateBasic(t *testing.T) {
	t.Run("no sender", func(t *testing.T) {
		vote := NewVoteRequest(nil, exported.PollID(rand.PosI64()), &gogoprototypes.StringValue{Value: "result"})
		assert.Error(t, vote.ValidateBasic())
	})

	t.Run("no vote", func(t *testing.T) {
		vote := &VoteRequest{
			Sender: rand.AccAddr(),
			PollID: exported.PollID(rand.PosI64()),
			Vote:   nil,
		}
		assert.Error(t, vote.ValidateBasic())
	})

	t.Run("faulty vote type", func(t *testing.T) {
		vote := &VoteRequest{
			Sender: rand.AccAddr(),
			PollID: exported.PollID(rand.PosI64()),
			Vote:   types.UnsafePackAny("result"),
		}
		assert.Error(t, vote.ValidateBasic())
	})

	t.Run("correct vote", func(t *testing.T) {
		vote := NewVoteRequest(rand.AccAddr(), exported.PollID(rand.PosI64()), &gogoprototypes.StringValue{Value: "result"})
		assert.NoError(t, vote.ValidateBasic())
	})
}
