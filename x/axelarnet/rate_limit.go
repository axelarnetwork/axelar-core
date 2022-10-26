package axelarnet

import (
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
)

type RateLimiter struct {
	keeper  keeper.Keeper
	channel porttypes.ICS4Wrapper
	// accountKeeper  *authkeeper.AccountKeeper
	// bankKeeper     *bankkeeper.BaseKeeper
	// paramSpace     paramtypes.Subspace
}

func (i RateLimiter) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI) error {
	// contract := i.GetParams(ctx)
	// if contract == "" {
	// 	// The contract has not been configured. Continue as usual
	// 	return i.channel.SendPacket(ctx, chanCap, packet)
	// }

	// amount, denom, err := GetFundsFromPacket(packet)
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "Rate limited SendPacket")
	// }
	// channelValue := i.CalculateChannelValue(ctx, denom)
	// err = CheckAndUpdateRateLimits(
	// 	ctx,
	// 	i.ContractKeeper,
	// 	"send_packet",
	// 	contract,
	// 	channelValue,
	// 	packet.GetSourceChannel(),
	// 	denom,
	// 	amount,
	// )
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "Rate limited SendPacket")
	// }

	return i.channel.SendPacket(ctx, chanCap, packet)
}

func (i RateLimiter) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
