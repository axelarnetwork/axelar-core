package types_test

import (
	"testing"

	address "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestMsgVotePubKey_Marshaling(t *testing.T) {
	// proper addresses need to be a specific length, otherwise the json unmarshaling fails
	sender := make([]byte, address.Len)
	for i := range sender {
		sender[i] = 0
	}
	result := &tofnd.MessageOut_KeygenResult{
		KeygenResultData: &tofnd.MessageOut_KeygenResult_Data{
			Data: &tofnd.KeygenOutput{
				PubKey: []byte("some bytes"), GroupRecoverInfo: []byte{0}, PrivateRecoverInfo: []byte{0, 1, 2, 3},
			},
		},
	}
	vote := tss.VotePubKeyRequest{
		Sender:  sender,
		PollKey: exported.NewPollKey("test", "test"),
		Result:  result,
	}
	encCfg := app.MakeEncodingConfig()

	bz := encCfg.Marshaler.MustMarshalLengthPrefixed(&vote)
	var msg tss.VotePubKeyRequest
	encCfg.Marshaler.MustUnmarshalLengthPrefixed(bz, &msg)

	assert.Equal(t, vote, msg)

	bz = encCfg.Marshaler.MustMarshalJSON(&vote)
	var msg2 tss.VotePubKeyRequest
	encCfg.Marshaler.MustUnmarshalJSON(bz, &msg2)

	assert.Equal(t, vote, msg2)
}
