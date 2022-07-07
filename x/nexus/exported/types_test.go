package exported_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
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
