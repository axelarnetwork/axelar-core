package exported_test

import (
	"testing"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/stretchr/testify/assert"
)

func TestTransferStateFromString(t *testing.T) {
	assert.Equal(t, exported.Pending, exported.TransferStateFromString("pending"))
	assert.Equal(t, exported.Archived, exported.TransferStateFromString("archived"))
	assert.Equal(t, exported.TRANSFER_STATE_UNSPECIFIED, exported.TransferStateFromString(rand.StrBetween(1, 100)))
}
