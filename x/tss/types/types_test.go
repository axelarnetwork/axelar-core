package types_test

import (
	"crypto/ecdsa"
	rand3 "crypto/rand"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
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
		Result:  *result,
	}
	encCfg := app.MakeEncodingConfig()

	bz := encCfg.Codec.MustMarshalLengthPrefixed(&vote)
	var msg tss.VotePubKeyRequest
	encCfg.Codec.MustUnmarshalLengthPrefixed(bz, &msg)

	assert.Equal(t, vote, msg)

	bz = encCfg.Codec.MustMarshalJSON(&vote)
	var msg2 tss.VotePubKeyRequest
	encCfg.Codec.MustUnmarshalJSON(bz, &msg2)

	assert.Equal(t, vote, msg2)
}

func TestMultisigKeyInfo(t *testing.T) {
	t.Run("should complete multisig keygen", testutils.Func(func(t *testing.T) {
		var validatorLst []sdk.ValAddress
		var validatorShares []int64
		totalShareCount := int64(0)
		for i := int64(0); i < rand.I64Between(1, 100); i++ {
			shares := rand.I64Between(1, 20)
			totalShareCount += shares
			validatorShares = append(validatorShares, shares)
			validatorLst = append(validatorLst, rand.ValAddr())
		}

		multisigKeygenInfo := tss.MultisigInfo{
			ID:        rand.StrBetween(5, 20),
			Timeout:   rand.I64Between(1, 20000),
			TargetNum: totalShareCount,
		}
		assert.False(t, multisigKeygenInfo.IsCompleted())
		assert.Equal(t, int64(0), multisigKeygenInfo.Count())

		var expectedPubKeys []ecdsa.PublicKey
		var expectedParticipant []sdk.ValAddress
		currKeys := int64(0)
		for i, val := range validatorLst {
			expectedParticipant = append(expectedParticipant, val)
			var pks [][]byte
			for j := int64(0); j < validatorShares[i]; j++ {
				pk := btcec.PublicKey(generatePubKey())
				pks = append(pks, pk.SerializeCompressed())
				expectedPubKeys = append(expectedPubKeys, *pk.ToECDSA())
			}
			multisigKeygenInfo.AddData(val, pks)
			currKeys += validatorShares[i]
			assert.Equal(t, currKeys, multisigKeygenInfo.Count())
		}
		assert.True(t, multisigKeygenInfo.IsCompleted())
		assert.Equal(t, expectedPubKeys, multisigKeygenInfo.GetKeys())
		for _, p := range expectedParticipant {
			assert.True(t, multisigKeygenInfo.DoesParticipate(p))
		}
	}).Repeat(20))
}

func generatePubKey() ecdsa.PublicKey {
	sk, err := ecdsa.GenerateKey(btcec.S256(), rand3.Reader)
	if err != nil {
		panic(err)
	}
	return sk.PublicKey
}
