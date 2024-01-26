package keeper_test

import (
	"fmt"
	"math"
	mathrand "math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestSetRateLimit(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	repeated := 1

	var (
		k      nexusKeeper.Keeper
		ctx    sdk.Context
		chain  exported.ChainName
		asset  string
		limit  sdk.Coin
		window time.Duration
	)

	givenKeeper := Given("a keeper", func() {
		k, ctx = setup(cfg)
	})

	whenAssetIsRegistered := When("asset is registered", func() {
		nexusChain := funcs.MustOk(k.GetChain(ctx, chain))
		err := k.RegisterAsset(ctx, nexusChain, exported.NewAsset(asset, false), utils.MaxUint, time.Hour)
		assert.NoError(t, err)
	})

	setRateLimitFails := func(msg string) ThenStatement {
		return Then(fmt.Sprintf("set rate limit will fail due to: %s", msg), func(t *testing.T) {
			err := k.SetRateLimit(ctx, chain, limit, window)
			assert.ErrorContains(t, err, msg)
		})
	}

	givenKeeper.
		When("using a non existent chain", func() {
			chain = exported.ChainName(rand.StrBetween(1, 20))
			asset = rand.Denom(3, 20)
			limit = sdk.NewInt64Coin(asset, mathrand.Int63())
			window = rand.Duration()
		}).
		Then2(setRateLimitFails("not a registered chain")).
		Run(t, repeated)

	givenKeeper.
		When("using a non registered asset", func() {
			chain = randChain(k, ctx).Name
			asset = rand.Denom(3, 20)
			limit = sdk.NewInt64Coin(asset, mathrand.Int63())
			window = rand.Duration()
		}).
		Then2(setRateLimitFails("not a registered asset")).
		Run(t, repeated)

	givenKeeper.
		When("using an invalid rate limit", func() {
			chain = randChain(k, ctx).Name
			asset = rand.Denom(3, 20)
			limit = sdk.Coin{Denom: asset, Amount: sdk.NewInt(-1)}
			window = rand.Duration()
		}).
		When2(whenAssetIsRegistered).
		Then2(setRateLimitFails("negative coin amount")).
		Run(t, repeated)

	givenKeeper.
		When("using an invalid window", func() {
			chain = randChain(k, ctx).Name
			asset = rand.Denom(3, 20)
			limit = sdk.NewInt64Coin(asset, mathrand.Int63())
			window = time.Duration(0)
		}).
		When2(whenAssetIsRegistered).
		Then2(setRateLimitFails("must be positive")).
		Run(t, repeated)

	givenKeeper.
		When("a rate limit is provided", func() {
			chain = randChain(k, ctx).Name
			asset = rand.Denom(3, 20)
			limit = sdk.NewInt64Coin(asset, mathrand.Int63())
			window = rand.Duration()
		}).
		When2(whenAssetIsRegistered).
		Then("set rate limit succeeds", func(t *testing.T) {
			err := k.SetRateLimit(ctx, chain, limit, window)
			assert.NoError(t, err)
		}).
		Then("set rate limit overwrite succeeds", func(t *testing.T) {
			limit = sdk.NewInt64Coin(asset, mathrand.Int63())
			window = rand.Duration()

			err := k.SetRateLimit(ctx, chain, limit, window)
			assert.NoError(t, err)
		}).
		Then("remove rate limit", func(t *testing.T) {
			limit.Amount = sdk.Int(utils.MaxUint)
			err := k.SetRateLimit(ctx, chain, limit, window)
			assert.NoError(t, err)
		}).
		Run(t, repeated)
}

func TestRateLimitTransfer(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	repeated := 10

	var (
		k         nexusKeeper.Keeper
		ctx       sdk.Context
		chain     exported.ChainName
		denom     string
		asset     sdk.Coin
		direction exported.TransferDirection
		limit     sdk.Coin
		window    time.Duration
	)

	givenKeeper := Given("a keeper", func() {
		k, ctx = setup(cfg)
	})

	whenAssetIsRegistered := When("asset is registered", func() {
		chain = randChain(k, ctx).Name
		denom = rand.Denom(3, 20)
		nexusChain := funcs.MustOk(k.GetChain(ctx, chain))
		err := k.RegisterAsset(ctx, nexusChain, exported.NewAsset(denom, false), utils.MaxUint, time.Hour)
		assert.NoError(t, err)
	})

	givenKeeper.
		When("no rate limit is set", func() {
			chain = nexustestutils.RandomChainName()
			asset = rand.Coin()
			direction = nexustestutils.RandomDirection()
		}).
		Then("rate limit transfer succeeds", func(t *testing.T) {
			err := k.RateLimitTransfer(ctx, chain, asset, direction)
			assert.NoError(t, err)
		}).
		Run(t, repeated)

	givenKeeper.
		When2(whenAssetIsRegistered).
		When("a rate limit is set", func() {
			limit = sdk.NewInt64Coin(denom, mathrand.Int63())
			window = rand.Duration()
			direction = nexustestutils.RandomDirection()

			err := k.SetRateLimit(ctx, chain, limit, window)
			assert.NoError(t, err)
		}).
		When("transfer amount is within rate limit", func() {
			asset = limit
			asset.Amount = rand.IntBetween(sdk.ZeroInt(), limit.Amount)
		}).
		Then("rate limit transfer succeeds", func(t *testing.T) {
			err := k.RateLimitTransfer(ctx, chain, asset, direction)
			assert.NoError(t, err)
		}).
		Run(t, repeated)

	givenKeeper.
		When2(whenAssetIsRegistered).
		When("a rate limit is set", func() {
			limit = sdk.NewInt64Coin(denom, mathrand.Int63())
			window = rand.Duration()

			err := k.SetRateLimit(ctx, chain, limit, window)
			assert.NoError(t, err)
		}).
		When("transfer amount is exactly the rate limit", func() {
			asset = limit
		}).
		Then("rate limit transfer succeeds", func(t *testing.T) {
			err := k.RateLimitTransfer(ctx, chain, asset, exported.TransferDirectionFrom)
			assert.NoError(t, err)

			err = k.RateLimitTransfer(ctx, chain, asset, exported.TransferDirectionTo)
			assert.NoError(t, err)
		}).
		Then("rate limit transfer fails on another transfer", func(t *testing.T) {
			asset = sdk.NewInt64Coin(asset.Denom, 1)
			err := k.RateLimitTransfer(ctx, chain, asset, exported.TransferDirectionFrom)
			assert.ErrorContains(t, err, "exceeded rate limit")

			err = k.RateLimitTransfer(ctx, chain, asset, exported.TransferDirectionTo)
			assert.ErrorContains(t, err, "exceeded rate limit")
		}).
		Then("reset rate limit and rate limit transfer succeeds", func(t *testing.T) {
			limit.Amount = sdk.Int(utils.MaxUint)
			err := k.SetRateLimit(ctx, chain, limit, window)
			assert.NoError(t, err)

			err = k.RateLimitTransfer(ctx, chain, asset, direction)
			assert.NoError(t, err)
		}).
		Run(t)

	givenKeeper.
		When2(whenAssetIsRegistered).
		When("a rate limit is set", func() {
			limit = sdk.NewInt64Coin(denom, mathrand.Int63())
			window = rand.Duration()
			direction = nexustestutils.RandomDirection()

			err := k.SetRateLimit(ctx, chain, limit, window)
			assert.NoError(t, err)
		}).
		When("transfer amount is above the rate limit", func() {
			asset = limit
			asset.Amount = rand.IntBetween(limit.Amount.AddRaw(1), limit.Amount.AddRaw(math.MaxInt64))
		}).
		Then("rate limit transfer fails", func(t *testing.T) {
			err := k.RateLimitTransfer(ctx, chain, asset, direction)
			assert.ErrorContains(t, err, "exceeded rate limit")
		}).
		Then("rate limit transfer succeeds on a small transfer", func(t *testing.T) {
			asset.Amount = rand.IntBetween(sdk.ZeroInt(), limit.Amount)
			err := k.RateLimitTransfer(ctx, chain, asset, direction)
			assert.NoError(t, err)
		}).
		Run(t)
}
