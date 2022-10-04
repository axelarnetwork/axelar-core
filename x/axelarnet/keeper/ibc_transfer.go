package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// IBCKeeper provides function to send IBC transfer
type IBCKeeper struct {
	Keeper
	ibcTransferK types.IBCTransferKeeper
	channelK     types.ChannelKeeper
}

// NewIBCKeeper returns a new  IBCKeeper
func NewIBCKeeper(k Keeper, ibcTransferK types.IBCTransferKeeper, channelK types.ChannelKeeper) IBCKeeper {
	return IBCKeeper{Keeper: k, ibcTransferK: ibcTransferK, channelK: channelK}
}

// SendIBCTransfer inits an IBC transfer
func (i IBCKeeper) SendIBCTransfer(ctx sdk.Context, transfer types.IBCTransfer) error {
	// map the packet sequence to transfer id
	err := i.SetSeqIDMapping(ctx, transfer)
	if err != nil {
		return err
	}

	_, state, err := i.channelK.GetChannelClientState(ctx, transfer.PortID, transfer.ChannelID)
	if err != nil {
		return err
	}

	height := clienttypes.NewHeight(state.GetLatestHeight().GetRevisionNumber(), state.GetLatestHeight().GetRevisionHeight()+i.GetRouteTimeoutWindow(ctx))
	return i.ibcTransferK.SendTransfer(ctx, transfer.PortID, transfer.ChannelID, transfer.Token, transfer.Sender, transfer.Receiver, height, 0)
}

// ParseIBCDenom retrieves the full identifiers trace and base denomination from the IBC transfer keeper store
func (i IBCKeeper) ParseIBCDenom(ctx sdk.Context, ibcDenom string) (ibctypes.DenomTrace, error) {
	denomSplit := strings.Split(ibcDenom, "/")

	hash, err := ibctypes.ParseHexHash(denomSplit[1])
	if err != nil {
		return ibctypes.DenomTrace{}, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid denom trace hash %s, %s", hash, err))
	}
	denomTrace, found := i.ibcTransferK.GetDenomTrace(ctx, hash)
	if !found {
		return ibctypes.DenomTrace{}, status.Error(
			codes.NotFound,
			sdkerrors.Wrap(ibctypes.ErrTraceNotFound, denomSplit[1]).Error(),
		)
	}
	return denomTrace, nil
}
