package axelarnet_test

import (
	"fmt"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	captypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	axelartestutils "github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	. "github.com/axelarnetwork/utils/test"
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
	port := rand.StrBetween(1, 20)
	channel := rand.StrBetween(1, 20)

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
		chain = nexustestutils.RandomChainName()
		direction = nexustestutils.RandomDirection()

		switch direction {
		case nexus.Incoming:
			packet = axelartestutils.RandomPacket(transfer, rand.StrBetween(1, 20), rand.StrBetween(1, 20), port, channel)
		case nexus.Outgoing:
			packet = axelartestutils.RandomPacket(transfer, port, channel, rand.StrBetween(1, 20), rand.StrBetween(1, 20))
		default:
			panic(fmt.Errorf("unknown direction %s", direction))
		}
	})

	whenIBCPathIsRegistered := When("ibc path is registered", func() {
		err = k.SetChainByIBCPath(ctx, types.NewIBCPath(port, channel), chain)
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
			packet.Data = ibctransfertypes.ModuleCdc.MustMarshalJSON(&transfer)
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
			packet.Data = ibctransfertypes.ModuleCdc.MustMarshalJSON(&transfer)
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

func TestSendPacket(t *testing.T) {
	var (
		ctx         sdk.Context
		k           keeper.Keeper
		packet      ibcchanneltypes.Packet
		transfer    ibctransfertypes.FungibleTokenPacketData
		denom       string
		rateLimiter axelarnet.RateLimiter
		n           *mock.NexusMock
		channelK    *mock.ChannelKeeperMock
		chain       nexus.ChainName
	)
	repeats := 10
	port := rand.StrBetween(1, 20)
	channel := rand.StrBetween(1, 20)

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
		packet = axelartestutils.RandomPacket(transfer, port, channel, rand.StrBetween(1, 20), rand.StrBetween(1, 20))
		chain = nexustestutils.RandomChainName()
	})

	givenKeeper.
		Given2(givenPacket).
		When("channel send packet fails", func() {
			channelK.SendPacketFunc = func(ctx sdk.Context, channelCap *captypes.Capability, packet ibcexported.PacketI) error {
				return fmt.Errorf("send packet failed")
			}
		}).
		Then("send packet fails", func(t *testing.T) {
			err := rateLimiter.SendPacket(ctx, &captypes.Capability{}, packet)
			assert.ErrorContains(t, err, "send packet failed")
		}).
		Run(t, repeats)

	givenKeeper.
		Given2(givenPacket).
		When("channel send packet succeeds", func() {
			channelK.SendPacketFunc = func(ctx sdk.Context, channelCap *captypes.Capability, packet ibcexported.PacketI) error {
				return nil
			}
		}).
		When("cross-chain transfer", func() {
			sequence := uint64(rand.PosI64())
			channelK.GetNextSequenceSendFunc = func(ctx sdk.Context, portID, channelID string) (uint64, bool) {
				return sequence, true
			}
			err := k.SetSeqIDMapping(ctx, types.IBCTransfer{
				Sequence:  packet.GetSequence(),
				PortID:    packet.GetSourcePort(),
				ChannelID: packet.GetSourceChannel(),
			})
			assert.NoError(t, err)
		}).
		Then("send packet succeeds", func(t *testing.T) {
			err := rateLimiter.SendPacket(ctx, &captypes.Capability{}, packet)
			assert.NoError(t, err)
		}).
		Run(t, repeats)

	givenKeeper.
		Given2(givenPacket).
		When("channel send packet succeeds", func() {
			channelK.SendPacketFunc = func(ctx sdk.Context, channelCap *captypes.Capability, packet ibcexported.PacketI) error {
				return nil
			}
		}).
		When("cross-chain transfer", func() {
			sequence := uint64(rand.PosI64())
			channelK.GetNextSequenceSendFunc = func(ctx sdk.Context, portID, channelID string) (uint64, bool) {
				return sequence, true
			}
			err := k.SetSeqIDMapping(ctx, types.IBCTransfer{
				Sequence:  sequence,
				PortID:    rand.StrBetween(1, 20),
				ChannelID: rand.StrBetween(1, 20),
			})
			assert.NoError(t, err)
			chain = nexustestutils.RandomChainName()
		}).
		When("rate limit packet fails", func() {
			err := k.SetChainByIBCPath(ctx, types.NewIBCPath(port, channel), chain)
			assert.NoError(t, err)

			n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{Name: chain}, true
			}
			n.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool {
				return false
			}
		}).
		Then("send packet fails", func(t *testing.T) {
			err := rateLimiter.SendPacket(ctx, &captypes.Capability{}, packet)
			assert.ErrorContains(t, err, "deactivated")
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
