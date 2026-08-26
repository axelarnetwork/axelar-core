package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/utils/funcs"
)

// TestMigrate8to9RebuildsProcessingIndex plants a processing message in the old
// layout (a message record plus a membership entry whose value is the id, with no
// order entry) and verifies Migrate8to9 re-indexes it so GetProcessingMessages
// returns it.
func TestMigrate8to9RebuildsProcessingIndex(t *testing.T) {
	ctx, k := setup(t)

	id, _, _ := k.GenerateMessageID(ctx)
	msg := getRandomMessage(id) // Status == Processing
	funcs.MustNoErr(k.setMessage(ctx, msg))
	// old layout: membership value is the id, no order entry
	k.getStore(ctx).SetRawNew(getProcessingMessageKey(msg.GetDestinationChain(), msg.ID), []byte(msg.ID))
	// a stale old-layout entry with no live message must be purged, not carried over
	staleKey := getProcessingMessageKey(msg.GetDestinationChain(), "0xstale-0")
	k.getStore(ctx).SetRawNew(staleKey, []byte("0xstale-0"))

	assert.Empty(t, k.GetProcessingMessages(ctx, msg.GetDestinationChain(), 100))

	funcs.MustNoErr(Migrate8to9(k)(ctx))

	got := k.GetProcessingMessages(ctx, msg.GetDestinationChain(), 100)
	assert.Len(t, got, 1)
	if len(got) == 1 {
		assert.Equal(t, msg.ID, got[0].ID)
	}
	assert.Nil(t, k.getStore(ctx).GetRawNew(staleKey), "stale old-layout entry should be purged")
}
