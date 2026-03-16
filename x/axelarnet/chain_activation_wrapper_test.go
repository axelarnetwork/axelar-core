package axelarnet_test

import (
	"testing"

	"cosmossdk.io/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/axelarnet"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func TestChainActivationICS4Wrapper(t *testing.T) {
	var (
		ctx     sdk.Context
		channel *mock.ChannelKeeperMock
		keeper  *mock.BaseKeeperMock
		n       *mock.NexusMock
		wrapper axelarnet.ChainActivationICS4Wrapper
	)

	setup := func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
		channel = &mock.ChannelKeeperMock{
			SendPacketFunc: func(ctx sdk.Context, _ *capabilitytypes.Capability, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (uint64, error) {
				return 1, nil
			},
		}
		keeper = &mock.BaseKeeperMock{}
		n = &mock.NexusMock{}
		wrapper = axelarnet.NewChainActivationICS4Wrapper(channel, keeper, n)
	}

	t.Run("packet to active chain is sent", func(t *testing.T) {
		setup()
		keeper.GetChainNameByIBCPathFunc = func(_ sdk.Context, _ string) (nexus.ChainName, bool) {
			return "osmosis", true
		}
		n.GetChainFunc = func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{Name: chain}, true
		}
		n.IsChainActivatedFunc = func(_ sdk.Context, _ nexus.Chain) bool { return true }

		seq, err := wrapper.SendPacket(ctx, nil, "transfer", "channel-0", clienttypes.Height{}, 0, []byte("data"))
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), seq)
		assert.Len(t, channel.SendPacketCalls(), 1)
	})

	t.Run("packet to deactivated chain is rejected", func(t *testing.T) {
		setup()
		keeper.GetChainNameByIBCPathFunc = func(_ sdk.Context, _ string) (nexus.ChainName, bool) {
			return "osmosis", true
		}
		n.GetChainFunc = func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{Name: chain}, true
		}
		n.IsChainActivatedFunc = func(_ sdk.Context, _ nexus.Chain) bool { return false }

		_, err := wrapper.SendPacket(ctx, nil, "transfer", "channel-0", clienttypes.Height{}, 0, []byte("data"))
		assert.ErrorContains(t, err, "deactivated")
		assert.Len(t, channel.SendPacketCalls(), 0)
	})

	t.Run("packet on unregistered IBC path is sent", func(t *testing.T) {
		setup()
		keeper.GetChainNameByIBCPathFunc = func(_ sdk.Context, _ string) (nexus.ChainName, bool) {
			return "", false
		}

		seq, err := wrapper.SendPacket(ctx, nil, "transfer", "channel-99", clienttypes.Height{}, 0, []byte("data"))
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), seq)
		assert.Len(t, channel.SendPacketCalls(), 1)
	})
}
