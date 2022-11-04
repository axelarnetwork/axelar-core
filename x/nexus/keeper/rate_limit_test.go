package keeper_test

import (
	mathrand "math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	testutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestSetRateLimit(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	repeated := 10

	var (
		k      nexusKeeper.Keeper
		ctx    sdk.Context
		chain  nexus.ChainName
		asset  string
		limit  sdk.Coin
		window time.Duration
	)

	givenKeeper := Given("a keeper", func() {
		k, ctx = setup(cfg)
	})

	whenAssetIsRegistered := When("asset is registered", func() {
		nexusChain := funcs.MustOk(k.GetChain(ctx, chain))
		err := k.RegisterAsset(ctx, nexusChain, nexus.NewAsset(asset, false))
		assert.NoError(t, err)
	})

	setRateLimitFails := func(msg string) ThenStatement {
		return Then("set rate limit will fail", func(t *testing.T) {
			err := k.SetRateLimit(ctx, chain, limit, window)
			assert.ErrorContains(t, err, msg)
		})
	}

	givenKeeper.
		When("using a non existent chain", func() {
			chain = nexus.ChainName(rand.StrBetween(1, 20))
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
		When("using an invalid denom", func() {
			chain = randChain(k, ctx).Name
			asset = "1" + rand.StrBetween(2, 20)
			limit = sdk.Coin{Denom: asset, Amount: sdk.NewInt(mathrand.Int63())}
			window = rand.Duration()
		}).
		When2(whenAssetIsRegistered).
		Then2(setRateLimitFails("invalid denom")).
		Run(t, repeated)

	givenKeeper.
		When("using an invalid rate limit", func() {
			chain = randChain(k, ctx).Name
			asset = rand.Denom(3, 20)
			limit = sdk.Coin{Denom: asset, Amount: sdk.NewInt(-1)}
			window = rand.Duration()
		}).
		When2(whenAssetIsRegistered).
		Then2(setRateLimitFails("")).
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
			assert.Equal(t, 1, len(ctx.EventManager().Events()))
		}).
		Run(t, repeated)
}

func TestRateLimitTransfer(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	repeated := 10

	var (
		k         nexusKeeper.Keeper
		ctx       sdk.Context
		chain     nexus.ChainName
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
		err := k.RegisterAsset(ctx, nexusChain, nexus.NewAsset(denom, false))
		assert.NoError(t, err)
	})

	givenKeeper.
		When("no rate limit is set", func() {
			chain = testutils.RandomChainName()
			asset = rand.Coin()
			direction = testutils.RandomDirection()
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
			direction = testutils.RandomDirection()

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
			err := k.RateLimitTransfer(ctx, chain, asset, exported.Incoming)
			assert.NoError(t, err)

			err = k.RateLimitTransfer(ctx, chain, asset, exported.Outgoing)
			assert.NoError(t, err)
		}).
		Then("rate limit transfer fails on another transfer", func(t *testing.T) {
			asset = sdk.NewInt64Coin(asset.Denom, 1)
			err := k.RateLimitTransfer(ctx, chain, asset, exported.Incoming)
			assert.ErrorContains(t, err, "exceeded rate limit")

			err = k.RateLimitTransfer(ctx, chain, asset, exported.Outgoing)
			assert.ErrorContains(t, err, "exceeded rate limit")
		}).
		Run(t)
}
