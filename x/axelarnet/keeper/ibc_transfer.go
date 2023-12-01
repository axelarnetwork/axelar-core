package keeper

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// IBCKeeper provides function to send IBC transfer
type IBCKeeper struct {
	Keeper
	ibcTransferK types.IBCTransferKeeper
}

// NewIBCKeeper returns a new  IBCKeeper
func NewIBCKeeper(k Keeper, ibcTransferK types.IBCTransferKeeper) IBCKeeper {
	return IBCKeeper{Keeper: k, ibcTransferK: ibcTransferK}
}

// SendIBCTransfer inits an IBC transfer
func (i IBCKeeper) SendIBCTransfer(ctx sdk.Context, transfer types.IBCTransfer) error {
	// map the packet sequence to transfer id
	err := i.SetSeqIDMapping(ctx, transfer)
	if err != nil {
		return err
	}

	height, err := i.getPacketTimeoutHeight(ctx, transfer.PortID, transfer.ChannelID)
	if err != nil {
		return err
	}

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

// SendMessage sends general message via ICS20 packet memo
func (i IBCKeeper) SendMessage(c context.Context, recipient nexus.CrossChainAddress, asset sdk.Coin, payload string, id string) error {
	ctx := sdk.UnwrapSDKContext(c)

	portID, channelID, err := i.getPortAndChannel(ctx, recipient.Chain.Name)
	if err != nil {
		return err
	}

	height, err := i.getPacketTimeoutHeight(ctx, portID, channelID)
	if err != nil {
		return err
	}

	msg := ibctypes.NewMsgTransfer(portID, channelID, asset, types.AxelarGMPAccount.String(), recipient.Address, height, 0)
	msg.Memo = payload

	res, err := i.ibcTransferK.Transfer(c, msg)
	if err != nil {
		return err
	}

	return i.SetSeqMessageIDMapping(ctx, portID, channelID, res.Sequence, id)
}

func (i IBCKeeper) getPacketTimeoutHeight(ctx sdk.Context, portID, channelID string) (clienttypes.Height, error) {
	_, state, err := i.channelK.GetChannelClientState(ctx, portID, channelID)
	if err != nil {
		return clienttypes.Height{}, err
	}

	return clienttypes.NewHeight(state.GetLatestHeight().GetRevisionNumber(), state.GetLatestHeight().GetRevisionHeight()+i.GetRouteTimeoutWindow(ctx)), nil
}

func (i IBCKeeper) getPortAndChannel(ctx sdk.Context, chain nexus.ChainName) (string, string, error) {
	path, ok := i.GetIBCPath(ctx, chain)
	if !ok {
		return "", "", fmt.Errorf("%s does not have a valid IBC path", chain.String())
	}

	pathSplit := strings.SplitN(path, "/", 2)
	if len(pathSplit) != 2 {
		return "", "", fmt.Errorf("invalid path %s for chain %s", path, chain.String())
	}

	return pathSplit[0], pathSplit[1], nil

}
