package utils_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	. "github.com/axelarnetwork/utils/test"
)

func TestCounter(t *testing.T) {
	var (
		counter utils.Counter[uint]
		ctx     sdk.Context
	)

	givenCounter := Given("the counter", func() {
		encCfg := params.MakeEncodingConfig()

		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		counter = utils.NewCounter[uint](key.FromUInt(uint64(rand.I64Between(10, 100))), utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("counter")), encCfg.Codec))
	})

	t.Run("Incr", func(t *testing.T) {
		givenCounter.
			When("", func() {}).
			Then("should increment the value one by one", func(t *testing.T) {
				count := int(rand.I64Between(10, 100))
				for i := 0; i < count; i++ {
					assert.Equal(t, uint(i), counter.Incr(ctx))
				}
			}).
			Run(t)
	})

	t.Run("Curr", func(t *testing.T) {
		givenCounter.
			When("counter has been incremented", func() {
				for i := 0; i < 10; i++ {
					assert.Equal(t, uint(i), counter.Incr(ctx))
				}
			}).
			Then("should get the current value", func(t *testing.T) {
				assert.Equal(t, uint(10), counter.Curr(ctx))
			}).
			Run(t)
	})
}
