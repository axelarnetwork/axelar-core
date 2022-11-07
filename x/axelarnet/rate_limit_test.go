package axelarnet_test

import (
	"fmt"
	"strconv"
	"testing"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	axelartestutils "github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	. "github.com/axelarnetwork/utils/test"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestRateLimitPacket(t *testing.T) {
	var (
		ctx         sdk.Context
		k           keeper.Keeper
		packet      ibcchanneltypes.Packet
		transfer    ibctransfertypes.FungibleTokenPacketData
		denom       string
		rateLimiter axelarnet.RateLimiter
		n           *mock.NexusMock
		channelK    *mock.ChannelKeeperMock
		err         error
		chain       nexus.ChainName
		direction   nexus.TransferDirection
	)
	repeats := 10

	givenKeeper := Given("a keeper", func() {
		ctx, k, channelK = setup()
		n = &mock.NexusMock{}
		rateLimiter = axelarnet.NewRateLimiter(k, channelK, n)
	})

	givenPacket := Given("a random ICS-20 packet", func() {
		denom = axelartestutils.RandomFullDenom()
		transfer = ibctransfertypes.NewFungibleTokenPacketData(
			denom, strconv.FormatInt(rand.PosI64(), 10), rand.AccAddr().String(), rand.AccAddr().String(),
		)
		packet = axelartestutils.RandomPacket(transfer)
		chain = nexustestutils.RandomChainName()
		direction = nexustestutils.RandomDirection()
	})

	whenIBCPathIsRegistered := When("ibc path is registered", func() {
		ibcPath := fmt.Sprintf("%s/%s", packet.GetSourcePort(), packet.GetSourceChannel())
		err = k.SetChainByIBCPath(ctx, ibcPath, chain)
		assert.NoError(t, err)
	})

	whenChainIsRegistered := When("chain is deactivated", func() {
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{Name: chain}, true
		}
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool {
			return true
		}
	})

	givenKeeper.
		Given2(givenPacket).
		When("ibc path is not registered", func() {}).
		Then("rate limit packet succeeds", func(t *testing.T) {
			err = rateLimiter.RateLimitPacket(ctx, packet, direction)
			assert.NoError(t, err)
		}).
		Run(t, repeats)

	givenKeeper.
		Given2(givenPacket).
		When2(whenIBCPathIsRegistered).
		When("chain is deactivated", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{Name: chain}, true
			}
			n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool {
				return false
			}
		}).
		Then("rate limit packet fails", func(t *testing.T) {
			err = rateLimiter.RateLimitPacket(ctx, packet, direction)
			assert.ErrorContains(t, err, "deactivated")
		}).
		Run(t, repeats)

	givenKeeper.
		Given2(givenPacket).
		When2(whenIBCPathIsRegistered).
		When2(whenChainIsRegistered).
		When("packet is not ICS-20 transfer", func() {
			packet.Data = nil
		}).
		Then("rate limit packet fails", func(t *testing.T) {
			err = rateLimiter.RateLimitPacket(ctx, packet, direction)
			assert.ErrorContains(t, err, "cannot unmarshal")
		}).
		Run(t, repeats)

	givenKeeper.
		Given2(givenPacket).
		When("invalid ICS-20 packet", func() {
			transfer.Amount = rand.StrBetween(1, 20) + "a"
			packet = axelartestutils.RandomPacket(transfer)
		}).
		When2(whenIBCPathIsRegistered).
		When2(whenChainIsRegistered).
		Then("rate limit packet fails", func(t *testing.T) {
			err = rateLimiter.RateLimitPacket(ctx, packet, direction)
			assert.ErrorContains(t, err, "unable to parse transfer amount")
		}).
		Run(t, repeats)

	givenKeeper.
		Given2(givenPacket).
		When("invalid ICS-20 packet", func() {
			transfer.Amount = "-" + strconv.FormatInt(rand.PosI64(), 10)
			packet = axelartestutils.RandomPacket(transfer)
		}).
		When2(whenIBCPathIsRegistered).
		When2(whenChainIsRegistered).
		Then("rate limit packet fails", func(t *testing.T) {
			err = rateLimiter.RateLimitPacket(ctx, packet, direction)
			assert.ErrorContains(t, err, "negative coin amount")
		}).
		Run(t, repeats)

	givenKeeper.
		Given2(givenPacket).
		When2(whenIBCPathIsRegistered).
		When2(whenChainIsRegistered).
		When("rate limit transfer exceeded", func() {
			n.RateLimitTransferFunc = func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
				return fmt.Errorf("rate limit exceeded")
			}
		}).
		Then("rate limit packet fails", func(t *testing.T) {
			err = rateLimiter.RateLimitPacket(ctx, packet, direction)
			assert.ErrorContains(t, err, "rate limit exceeded")
		}).
		Run(t, repeats)

	givenKeeper.
		Given2(givenPacket).
		When2(whenIBCPathIsRegistered).
		When2(whenChainIsRegistered).
		When("rate limit transfer succeeds", func() {
			n.RateLimitTransferFunc = func(ctx sdk.Context, chain nexus.ChainName, asset sdk.Coin, direction nexus.TransferDirection) error {
				return nil
			}
		}).
		Then("rate limit packet succeeds", func(t *testing.T) {
			err = rateLimiter.RateLimitPacket(ctx, packet, direction)
			assert.NoError(t, err)
		}).
		Run(t, repeats)
}

func setup() (sdk.Context, keeper.Keeper, *mock.ChannelKeeperMock) {
	encCfg := appParams.MakeEncodingConfig()
	axelarnetSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

	channelK := &mock.ChannelKeeperMock{}

	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), axelarnetSubspace, channelK)
	return ctx, k, channelK
}
