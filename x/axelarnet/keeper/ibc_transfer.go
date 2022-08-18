package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// IBCKeeper provides function to send IBC transfer
type IBCKeeper struct {
	k            Keeper
	ibcTransferK types.IBCTransferKeeper
	channelK     types.ChannelKeeper
}

// NewIBCKeeper returns a new  IBCKeeper
func NewIBCKeeper(k Keeper, ibcTransferK types.IBCTransferKeeper, channelK types.ChannelKeeper) IBCKeeper {
	return IBCKeeper{k: k, ibcTransferK: ibcTransferK, channelK: channelK}
}

// SendIBCTransfer inits an IBC transfer
func (i IBCKeeper) SendIBCTransfer(ctx sdk.Context, transfer types.IBCTransfer) error {
	// map the packet sequence to transfer id
	err := i.k.SetSeqIDMapping(ctx, transfer)
	if err != nil {
		return err
	}

	_, state, err := i.channelK.GetChannelClientState(ctx, transfer.PortID, transfer.ChannelID)
	if err != nil {
		return err
	}

	height := clienttypes.NewHeight(state.GetLatestHeight().GetRevisionNumber(), state.GetLatestHeight().GetRevisionHeight()+i.k.GetRouteTimeoutWindow(ctx))
	return i.ibcTransferK.SendTransfer(ctx, transfer.PortID, transfer.ChannelID, transfer.Token, transfer.Sender, transfer.Receiver, height, 0)
}
