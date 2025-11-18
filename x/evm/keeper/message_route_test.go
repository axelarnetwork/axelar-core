package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func TestNewMessageRoute(t *testing.T) {
	route := keeper.NewMessageRoute()

	t.Run("should increment the gas meter", func(t *testing.T) {
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))

		assert.NoError(t, route(ctx, nexus.RoutingContext{}, nexus.GeneralMessage{}))
		assert.Positive(t, ctx.GasMeter().GasConsumed())
	})
}
