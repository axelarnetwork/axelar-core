package ante_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/x/ante"
)

// mockFeeTx is a minimal sdk.FeeTx; the decorator only reads GetFee().
type mockFeeTx struct {
	sdk.Tx
	fee sdk.Coins
}

func (m mockFeeTx) GetGas() uint64     { return 0 }
func (m mockFeeTx) GetFee() sdk.Coins  { return m.fee }
func (m mockFeeTx) FeePayer() []byte   { return nil }
func (m mockFeeTx) FeeGranter() []byte { return nil }

func coin(denom string, amount int64) sdk.Coin {
	return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
}

// fakeFeePolicy is a static FeePolicy for testing.
type fakeFeePolicy struct{ denoms []string }

func (f fakeFeePolicy) GetAllowedFeeDenoms(sdk.Context) []string { return f.denoms }

func TestLimitFeeDenomDecorator(t *testing.T) {
	const allowed = "uaxl"
	decorator := ante.NewLimitFeeDenomDecorator(fakeFeePolicy{denoms: []string{allowed}})

	run := func(fee sdk.Coins) (bool, error) {
		nextCalled := false
		next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
			nextCalled = true
			return ctx, nil
		}
		_, err := decorator.AnteHandle(sdk.Context{}, mockFeeTx{fee: fee}, false, next)
		return nextCalled, err
	}

	t.Run("single allowed denom passes", func(t *testing.T) {
		passed, err := run(sdk.Coins{coin(allowed, 100)})
		require.NoError(t, err)
		assert.True(t, passed)
	})

	t.Run("empty fee passes", func(t *testing.T) {
		passed, err := run(sdk.Coins{})
		require.NoError(t, err)
		assert.True(t, passed)
	})

	t.Run("zero-amount allowed denom passes", func(t *testing.T) {
		passed, err := run(sdk.Coins{coin(allowed, 0)})
		require.NoError(t, err)
		assert.True(t, passed)
	})

	t.Run("zero-amount non-allowed denom is rejected", func(t *testing.T) {
		// The allowlist applies regardless of amount: a single [stake:0] is rejected.
		passed, err := run(sdk.Coins{coin("stake", 0)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed")
		assert.False(t, passed)
	})

	t.Run("single non-allowed denom is rejected", func(t *testing.T) {
		passed, err := run(sdk.Coins{coin("ibc/DEADBEEF", 100)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed")
		assert.False(t, passed)
	})

	t.Run("two positive denoms are rejected even if one is allowed", func(t *testing.T) {
		// gas paid in the allowed denom + a junk voucher riding along.
		passed, err := run(sdk.Coins{coin(allowed, 100), coin("ibc/DEADBEEF", 100)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at most one denomination")
		assert.False(t, passed)
	})

	t.Run("two zero-amount denoms are rejected", func(t *testing.T) {
		// [0uaxl, 0btc]: two denominations, rejected even though both are zero.
		passed, err := run(sdk.Coins{coin(allowed, 0), coin("btc", 0)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at most one denomination")
		assert.False(t, passed)
	})

	t.Run("allowed positive denom plus a zero-amount junk coin is rejected", func(t *testing.T) {
		// Still two denominations, so rejected by the count rule.
		passed, err := run(sdk.Coins{coin(allowed, 100), coin("ibc/DEADBEEF", 0)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at most one denomination")
		assert.False(t, passed)
	})
}
