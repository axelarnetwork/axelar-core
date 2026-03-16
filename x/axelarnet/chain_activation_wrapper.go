package axelarnet

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/utils/funcs"
)

// ChainActivationICS4Wrapper wraps an ICS4Wrapper and rejects outgoing IBC
// packets to deactivated chains.
type ChainActivationICS4Wrapper struct {
	channel porttypes.ICS4Wrapper
	keeper  types.BaseKeeper
	nexus   types.Nexus
}

func NewChainActivationICS4Wrapper(channel porttypes.ICS4Wrapper, keeper types.BaseKeeper, nexus types.Nexus) ChainActivationICS4Wrapper {
	return ChainActivationICS4Wrapper{
		channel: channel,
		keeper:  keeper,
		nexus:   nexus,
	}
}

func (w ChainActivationICS4Wrapper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	ibcPath := types.NewIBCPath(sourcePort, sourceChannel)
	chainName, ok := w.keeper.GetChainNameByIBCPath(ctx, ibcPath)
	if ok {
		chain := funcs.MustOk(w.nexus.GetChain(ctx, chainName))
		if !w.nexus.IsChainActivated(ctx, chain) {
			return 0, fmt.Errorf("chain %s registered for IBC path %s is deactivated", chain.Name, ibcPath)
		}
	}

	return w.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

func (w ChainActivationICS4Wrapper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return w.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

func (w ChainActivationICS4Wrapper) GetAppVersion(ctx sdk.Context, portID string, channelID string) (string, bool) {
	return w.channel.GetAppVersion(ctx, portID, channelID)
}
