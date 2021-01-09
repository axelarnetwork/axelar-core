package cli

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
)

func TestGetCmdSignStart_UnmarshalJSONTxHash(t *testing.T) {
	input := "{\"type\":\"tss/bytes\",\"value\":\"1EtL4N8J5dE37bL53qRRbtSW0dQbGYjseMr4ks3lbNM=\"}"
	bz := []byte(input)
	var realMsg []byte
	cdc := testutils.Codec()
	cdc.MustUnmarshalJSON(bz, &realMsg)
	assert.NotNil(t, realMsg)
}

func TestGetCmdSignStart_MarshalJSONTxHash(t *testing.T) {
	realMsg := []byte(testutils.RandString(chainhash.HashSize))
	cdc := testutils.Codec()
	bz := cdc.MustMarshalJSON(&realMsg)
	assert.NotNil(t, bz)
}

func TestGetCmdSignStart_UnmarshalJSONMarshalJSONTxHash(t *testing.T) {
	input := "{\"type\":\"tss/bytes\",\"value\":\"1EtL4N8J5dE37bL53qRRbtSW0dQbGYjseMr4ks3lbNM=\"}"
	bz := []byte(input)
	var realMsg []byte
	cdc := testutils.Codec()
	cdc.MustUnmarshalJSON(bz, &realMsg)
	assert.NotNil(t, realMsg)

	bz = cdc.MustMarshalJSON(&realMsg)
	assert.NotNil(t, bz)
}
