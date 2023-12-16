package exported_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

func TestTransferStateFromString(t *testing.T) {
	assert.Equal(t, exported.Pending, exported.TransferStateFromString("pending"))
	assert.Equal(t, exported.Archived, exported.TransferStateFromString("archived"))
	assert.Equal(t, exported.TRANSFER_STATE_UNSPECIFIED, exported.TransferStateFromString(rand.StrBetween(1, 100)))
}

func TestChainName(t *testing.T) {
	invalidName := exported.ChainName(rand.NormalizedStr(exported.ChainNameLengthMax + 1))
	assert.Error(t, invalidName.Validate())

	validName := exported.ChainName(rand.NormalizedStr(exported.ChainNameLengthMax))
	assert.NoError(t, validName.Validate())
}

func TestWasmBytes_MarshalJSON(t *testing.T) {
	bz, err := json.Marshal(exported.WasmBytes(funcs.Must(hex.DecodeString("cb9b5566c2f4876853333e481f4698350154259ffe6226e283b16ce18a64bcf1"))))

	assert.NoError(t, err)
	assert.Equal(t, []byte("[203,155,85,102,194,244,135,104,83,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241]"), bz)
}

func TestWasmBytes_UnmarshalJSON(t *testing.T) {
	var bz exported.WasmBytes
	err := json.Unmarshal([]byte("[203,155,85,102,194,244,135,104,83,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241]"), &bz)

	assert.NoError(t, err)
	assert.Equal(t, exported.WasmBytes(funcs.Must(hex.DecodeString("cb9b5566c2f4876853333e481f4698350154259ffe6226e283b16ce18a64bcf1"))), bz)
}
