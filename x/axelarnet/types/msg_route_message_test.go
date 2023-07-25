package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

func TestRouteMessageRequest_ValidateBasic(t *testing.T) {
	t.Run("invalid sender", func(t *testing.T) {
		message := NewRouteMessage(nil, nil, rand.NormalizedStr(5), []byte(rand.NormalizedStr(10)))
		assert.Error(t, message.ValidateBasic())
	})

	t.Run("invalid feegranter", func(t *testing.T) {
		acc := make([]byte, 256)
		message := NewRouteMessage(rand.AccAddr(), sdk.AccAddress(acc), rand.NormalizedStr(5), []byte(rand.NormalizedStr(10)))
		assert.Error(t, message.ValidateBasic())
	})

	t.Run("invalid id", func(t *testing.T) {
		sender := rand.AccAddr()
		message := NewRouteMessage(sender, nil, "", []byte(rand.NormalizedStr(10)))
		assert.Error(t, message.ValidateBasic())
	})

	t.Run("correct message", func(t *testing.T) {
		sender := rand.AccAddr()
		feegranter := rand.AccAddr()
		id := rand.NormalizedStr(5)
		payload := rand.BytesBetween(0, 10)

		message := NewRouteMessage(sender, nil, id, payload)
		assert.NoError(t, message.ValidateBasic())

		message = NewRouteMessage(sender, feegranter, id, payload)
		assert.NoError(t, message.ValidateBasic())
	})

	t.Run("allow empty payload", func(t *testing.T) {
		sender := rand.AccAddr()
		id := rand.NormalizedStr(5)
		message := NewRouteMessage(sender, nil, id, []byte{})
		assert.NoError(t, message.ValidateBasic())
	})
}
