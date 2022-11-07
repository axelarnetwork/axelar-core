package axelarnet

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexustypes "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// RateLimiter implements an ICS4Wrapper middleware to rate limit IBC transfers
type RateLimiter struct {
	keeper  keeper.Keeper
	channel porttypes.ICS4Wrapper
	nexus   types.Nexus
}

// NewRateLimiter returns a new RateLimiter
func NewRateLimiter(keeper keeper.Keeper, channel porttypes.ICS4Wrapper, nexus types.Nexus) RateLimiter {
	return RateLimiter{
		keeper:  keeper,
		channel: channel,
		nexus:   nexus,
	}
}

// RateLimitPacket implements rate limiting of IBC packets
// - If the IBC channel that the packet is sent on is a registered chain, check the activation status.
// - If the packet is an ICS-20 coin transfer, apply rate limiting on (chain, base denom) pair.
// - If the rate limit is exceeded, an error is returned.
func (r RateLimiter) RateLimitPacket(ctx sdk.Context, packet ibcexported.PacketI, direction nexustypes.TransferDirection) error {
	ibcPath := fmt.Sprintf("%s/%s", packet.GetSourcePort(), packet.GetSourceChannel())
	chainName, ok := r.keeper.GetChainNameByIBCPath(ctx, ibcPath)
	if !ok {
		// TODO: if IBC channel is not registered, use axelarnet for activation/deactivation state
		return nil
	}
	chain := funcs.MustOk(r.nexus.GetChain(ctx, chainName))

	if !r.nexus.IsChainActivated(ctx, chain) {
		return fmt.Errorf("chain %s registered for IBC path %s is deactivated", chain.Name, ibcPath)
	}

	// Only ICS-20 packets are expected in the middleware
	var data ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	// parse the denomination from the full denom path
	trace := ibctransfertypes.ParseDenomTrace(data.Denom)

	// parse the transfer amount
	transferAmount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return sdkerrors.Wrapf(ibctransfertypes.ErrInvalidAmount, "unable to parse transfer amount (%s) into sdk.Int", data.Amount)
	}
	token := sdk.Coin{Denom: trace.GetBaseDenom(), Amount: transferAmount}
	if err := token.Validate(); err != nil {
		return err
	}

	if err := r.nexus.RateLimitTransfer(ctx, chain.Name, token, direction); err != nil {
		return err
	}

	return nil
}

// SendPacket implements the ICS4 Wrapper interface
func (r RateLimiter) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI) error {
	if err := r.channel.SendPacket(ctx, chanCap, packet); err != nil {
		return err
	}

	// Cross-chain transfers using IBC have already been tracked by EnqueueTransfer, so skip those
	if _, found := r.keeper.GetSeqIDMapping(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence()); found {
		return nil
	}

	return r.RateLimitPacket(ctx, packet, nexustypes.Outgoing)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (r RateLimiter) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return r.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
