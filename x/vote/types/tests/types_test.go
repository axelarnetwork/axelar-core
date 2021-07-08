package tests

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

func TestSdkInt_Marshaling(t *testing.T) {
	i := sdk.NewInt(75)
	cdc := app.MakeEncodingConfig().Amino

	bz := cdc.MustMarshalBinaryLengthPrefixed(i)
	var unmarshaled sdk.Int
	cdc.MustUnmarshalBinaryLengthPrefixed(bz, &unmarshaled)

	assert.Equal(t, unmarshaled, i)

	bz = cdc.MustMarshalJSON(i)
	var unmarshaled2 sdk.Int
	cdc.MustUnmarshalJSON(bz, &unmarshaled2)

	assert.Equal(t, unmarshaled2, i)

}

func TestTalliedVote_Marshaling(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	cdc := encCfg.Marshaler

	data := tofnd.MessageOut_KeygenResult{KeygenResultData: &tofnd.MessageOut_KeygenResult_Pubkey{Pubkey: []byte("a public key")}}
	vote := types.NewTalliedVote(snapshot.NewValidator(&mock.SDKValidatorMock{GetOperatorFunc: func() sdk.ValAddress {
		return rand.Bytes(sdk.AddrLen)
	}}, 23), &data)

	bz := cdc.MustMarshalBinaryLengthPrefixed(&vote)
	var actual types.TalliedVote
	cdc.MustUnmarshalBinaryLengthPrefixed(bz, &actual)

	assert.Equal(t, vote, actual)

	bz = cdc.MustMarshalJSON(&vote)
	var actual2 types.TalliedVote
	cdc.MustUnmarshalJSON(bz, &actual2)

	assert.Equal(t, vote.Tally, actual2.Tally)
	assert.Equal(t, vote.Data.GetCachedValue(), actual2.Data.GetCachedValue())
}
