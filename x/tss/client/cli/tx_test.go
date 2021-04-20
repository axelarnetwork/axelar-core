package cli

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func TestGetCmdSignStart_UnmarshalJSONTxHash(t *testing.T) {
	input := "{\"type\":\"tss/bytes\",\"value\":\"1EtL4N8J5dE37bL53qRRbtSW0dQbGYjseMr4ks3lbNM=\"}"
	bz := []byte(input)
	var realMsg []byte
	types.ModuleCdc.LegacyAmino.MustUnmarshalJSON(bz, &realMsg)
	assert.NotNil(t, realMsg)
}

func TestGetCmdSignStart_MarshalJSONTxHash(t *testing.T) {
	realMsg := []byte(rand.Str(chainhash.HashSize))

	bz := types.ModuleCdc.LegacyAmino.MustMarshalJSON(&realMsg)
	assert.NotNil(t, bz)
}

func TestGetCmdSignStart_UnmarshalJSONMarshalJSONTxHash(t *testing.T) {
	input := "{\"type\":\"tss/bytes\",\"value\":\"1EtL4N8J5dE37bL53qRRbtSW0dQbGYjseMr4ks3lbNM=\"}"
	bz := []byte(input)
	var realMsg []byte
	types.ModuleCdc.LegacyAmino.MustUnmarshalJSON(bz, &realMsg)
	assert.NotNil(t, realMsg)

	bz = types.ModuleCdc.LegacyAmino.MustMarshalJSON(&realMsg)
	assert.NotNil(t, bz)
}
