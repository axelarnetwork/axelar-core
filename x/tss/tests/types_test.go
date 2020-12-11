package tests

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

func TestMsgVotePubKey_Marshaling(t *testing.T) {
	// proper addresses need to be a specific length, otherwise the json unmarshaling fails
	sender := make([]byte, sdk.AddrLen)
	for i := range sender {
		sender[i] = 0
	}
	vote := tss.MsgVotePubKey{
		Sender: sender,
		PollMeta: exported.PollMeta{
			Module: "test",
			Type:   "test",
			ID:     "test",
		},
		PubKeyBytes: []byte("some bytes"),
	}
	ballot := types.MsgBallot{
		Votes:  []exported.MsgVote{&vote},
		Sender: sender,
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
