package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// TestSetProcessingMessageIDRetryGoesToBack verifies that a message which fails and is
// retried re-enters the queue at the back, keeping delivery order FIFO (scheduled later
// => executed later). A second setProcessingMessageID call while still queued is a no-op.
func TestSetProcessingMessageIDRetryGoesToBack(t *testing.T) {
	ctx, k := setup(t)

	route := func() exported.GeneralMessage {
		id, _, _ := k.GenerateMessageID(ctx)
		m := getRandomMessage(id) // Status == Processing
		funcs.MustNoErr(k.setMessage(ctx, m))
		funcs.MustNoErr(k.setProcessingMessageID(ctx, m))
		return m
	}

	ids := func() []string {
		got := k.GetProcessingMessages(ctx, evm.Ethereum.Name, 100)
		out := make([]string, len(got))
		for i, m := range got {
			out[i] = m.ID
		}
		return out
	}

	first := route()
	second := route()

	// re-queuing an already-queued message keeps its slot
	funcs.MustNoErr(k.setProcessingMessageID(ctx, first))
	assert.Equal(t, []string{first.ID, second.ID}, ids())

	// first fails, then is retried (re-enters processing)
	funcs.MustNoErr(k.SetMessageFailed(ctx, first.ID))
	retried, _ := k.GetMessage(ctx, first.ID)
	retried.Status = exported.Processing
	funcs.MustNoErr(k.setMessage(ctx, retried))
	funcs.MustNoErr(k.setProcessingMessageID(ctx, retried))

	// the retry re-entered at the back, so it is now delivered after second
	assert.Equal(t, []string{second.ID, first.ID}, ids())
}
