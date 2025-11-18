package testutils

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// RandRateLimit returns a random rate limit for a given chain and asset
func RandRateLimit(chain exported.ChainName, asset string) types.RateLimit {
	return types.RateLimit{Chain: chain, Limit: sdk.NewCoin(asset, RandInt(100000000, 200000000)), Window: time.Hour}
}

// RandFee returns a random fee info for a given chain and asset
func RandFee(chain exported.ChainName, asset string) exported.FeeInfo {
	rate := math.LegacyNewDecWithPrec(math.Int(RandInt(0, 100)).Int64(), 3)
	min := RandInt(0, 10)
	max := RandInt(min.Int64()+1, 100)
	return exported.NewFeeInfo(chain, asset, rate, min, max)
}

// RandInt returns a random sdk.Int for a given int64 range
func RandInt(min int64, max int64) math.Int {
	return math.NewInt(rand.I64Between(min, max))
}
