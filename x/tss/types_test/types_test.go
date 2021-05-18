package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestMsgVotePubKey_Marshaling(t *testing.T) {
	// proper addresses need to be a specific length, otherwise the json unmarshaling fails
	sender := make([]byte, sdk.AddrLen)
	for i := range sender {
		sender[i] = 0
	}
	vote := &tss.VotePubKeyRequest{
		Sender:      sender,
		PollMeta:    exported.NewPollMeta("test", "test"),
		PubKeyBytes: []byte("some bytes"),
	}
	encCfg := testutils.MakeEncodingConfig()

	bz := encCfg.Amino.MustMarshalBinaryLengthPrefixed(vote)
	var msg sdk.Msg
	encCfg.Amino.MustUnmarshalBinaryLengthPrefixed(bz, &msg)

	assert.Equal(t, vote, msg)

	bz = encCfg.Amino.MustMarshalJSON(vote)
	var msg2 sdk.Msg
	encCfg.Amino.MustUnmarshalJSON(bz, &msg2)

	assert.Equal(t, vote, msg2)
}
