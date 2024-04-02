package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/auxiliary/types"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
)

func TestBatchRequest_ValidateBasic(t *testing.T) {
	t.Run("should fail with nested batch", func(t *testing.T) {
		sender := rand.AccAddr()

		linkRequest := evmtypes.NewLinkRequest(sender, rand.NormalizedStr(5), rand.NormalizedStr(5), rand.NormalizedStr(5), rand.NormalizedStr(5))
		batch := types.NewBatchRequest(sender, []sdk.Msg{linkRequest})
		message := types.NewBatchRequest(sender, []sdk.Msg{linkRequest, batch})

		assert.ErrorContains(t, message.ValidateBasic(), "nested batch")
	})

	t.Run("should fail with different signers", func(t *testing.T) {

		message := types.NewBatchRequest(rand.AccAddr(), []sdk.Msg{
			evmtypes.NewLinkRequest(rand.AccAddr(), rand.NormalizedStr(5), rand.NormalizedStr(5), rand.NormalizedStr(5), rand.NormalizedStr(5)),
		})

		assert.ErrorContains(t, message.ValidateBasic(), "message signer mismatch")
	})
}
