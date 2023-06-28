package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

func TestRouteMessageRequest_ValidateBasic(t *testing.T) {
	t.Run("invalid sender", func(t *testing.T) {
		message := NewRouteMessage(nil, rand.NormalizedStr(5), []byte(rand.NormalizedStr(10)))
		assert.Error(t, message.ValidateBasic())
	})

	t.Run("invalid id", func(t *testing.T) {
		sender := rand.AccAddr()
		message := NewRouteMessage(sender, "", []byte(rand.NormalizedStr(10)))
		assert.Error(t, message.ValidateBasic())
	})

	t.Run("correct message", func(t *testing.T) {
		sender := rand.AccAddr()
		id := rand.NormalizedStr(5)
		payload := rand.BytesBetween(0, 10)
		message := NewRouteMessage(sender, id, payload)
		assert.NoError(t, message.ValidateBasic())
	})

	t.Run("allow empty payload", func(t *testing.T) {
		sender := rand.AccAddr()
		id := rand.NormalizedStr(5)
		message := NewRouteMessage(sender, id, []byte{})
		assert.NoError(t, message.ValidateBasic())
	})
}
