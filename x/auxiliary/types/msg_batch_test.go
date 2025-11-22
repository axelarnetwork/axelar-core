package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/auxiliary/types"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/utils/funcs"
)

func TestBatchRequest_ValidateBasic(t *testing.T) {
	t.Run("should fail with nested batch", func(t *testing.T) {
		sender := rand.AccAddr()

		linkRequest := evmtypes.NewLinkRequest(sender, rand.NormalizedStr(5), rand.NormalizedStr(5), rand.NormalizedStr(5), rand.NormalizedStr(5))
		batch := types.NewBatchRequest(sender, []sdk.Msg{linkRequest})
		message := types.NewBatchRequest(sender, []sdk.Msg{linkRequest, batch})

		assert.ErrorContains(t, message.ValidateBasic(), "nested batch")
	})

	t.Run("should unwrap messages", func(t *testing.T) {
		cdc := app.MakeEncodingConfig().Codec

		sender := rand.AccAddr()
		messages := []sdk.Msg{
			evmtypes.NewLinkRequest(sender, rand.NormalizedStr(5), rand.NormalizedStr(5), rand.NormalizedStr(5), rand.Denom(5, 10)),
			evmtypes.NewLinkRequest(sender, rand.NormalizedStr(5), rand.NormalizedStr(5), rand.NormalizedStr(5), rand.Denom(5, 10)),
		}
		batch := types.NewBatchRequest(sender, messages)

		bz := funcs.Must(batch.Marshal())
		var unmarshaledBatch types.BatchRequest
		funcs.MustNoErr(cdc.Unmarshal(bz, &unmarshaledBatch))

		assert.Equal(t, messages, unmarshaledBatch.UnwrapMessages())
	})
}
