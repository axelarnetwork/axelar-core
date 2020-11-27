package tests

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/types"
)

func TestMsgVotePubKey_Marshaling(t *testing.T) {
	vote := tss.MsgVotePubKey{
		PollMeta: exported.PollMeta{
			Module: "test",
			Type:   "test",
			ID:     "test",
		},
		PubKeyBytes: []byte("some bytes"),
	}
	ballot := types.MsgBallot{
		Votes:  []exported.MsgVote{&vote},
		Sender: sdk.AccAddress("sender"),
	}
	cdc := testutils.Codec()

	bz := cdc.MustMarshalBinaryLengthPrefixed(ballot)
	var msg sdk.Msg
	cdc.MustUnmarshalBinaryLengthPrefixed(bz, &msg)

	assert.Equal(t, ballot, msg)

	bz = cdc.MustMarshalJSON(ballot)
	var msg2 sdk.Msg
	cdc.MustUnmarshalJSON(bz, &msg2)

	assert.Equal(t, ballot, msg2)
}
