package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctransfertypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.BaseKeeper
	nexus        types.Nexus
	bank         types.BankKeeper
	ibcTransfer  types.IBCTransferKeeper
	ibcChannel   types.ChannelKeeper
	account      types.AccountKeeper
	msgSvcRouter *baseapp.MsgServiceRouter
	router       sdk.Router
}

// NewMsgServerImpl returns an implementation of the axelarnet MsgServiceServer interface for the provided Keeper.
func NewMsgServerImpl(k types.BaseKeeper, n types.Nexus, b types.BankKeeper, t types.IBCTransferKeeper, c types.ChannelKeeper, a types.AccountKeeper, m *baseapp.MsgServiceRouter, r sdk.Router) types.MsgServiceServer {
	return msgServer{
		BaseKeeper:   k,
		nexus:        n,
		bank:         b,
		ibcTransfer:  t,
		ibcChannel:   c,
		account:      a,
		msgSvcRouter: m,
		router:       r,
	}
}

// Link handles address linking
func (s msgServer) Link(c context.Context, req *types.LinkRequest) (*types.LinkResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	recipientChain, ok := s.nexus.GetChain(ctx, req.RecipientChain)
	if !ok {
		return nil, fmt.Errorf("unknown recipient chain")
	}

	found := s.nexus.IsAssetRegistered(ctx, recipientChain.Name, req.Asset)
	if !found {
		return nil, fmt.Errorf("asset '%s' not registered for chain '%s'", req.Asset, recipientChain.Name)
	}

	depositAddress := types.NewLinkedAddress(recipientChain.Name, req.Asset, req.RecipientAddr)
	s.nexus.LinkAddresses(ctx,
		nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: depositAddress.String()},
		nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr})

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLink,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, depositAddress.String()),
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipientChain.Name),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, req.RecipientAddr),
		),
	)

	return &types.LinkResponse{DepositAddr: depositAddress.String()}, nil
}

// ConfirmDeposit handles deposit confirmations
func (s msgServer) ConfirmDeposit(c context.Context, req *types.ConfirmDepositRequest) (*types.ConfirmDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	depositAddr := nexus.CrossChainAddress{Address: req.DepositAddress.String(), Chain: exported.Axelarnet}

	// deposit can be either of ICS 20 token from cosmos based chains, Axelarnet native asset, and wrapped asset from supported chain
	switch {
	// check if the format of token denomination is 'ibc/{hash}'
	case isIBCDenom(req.Token.Denom):
		// get base denomination and tracing path
		denomTrace, err := s.parseIBCDenom(ctx, req.Token.Denom)
		if err != nil {
			return nil, err
		}

		// check if the asset registered with a path
		chain, ok := s.BaseKeeper.GetCosmosChain(ctx, denomTrace.GetBaseDenom())
		if !ok {
			return nil, fmt.Errorf("asset %s is not linked to a cosmos chain", denomTrace.GetBaseDenom())
		}
		path, ok := s.BaseKeeper.GetIBCPath(ctx, chain)
		if !ok {
			return nil, fmt.Errorf("path not found for chain %s", chain)
		}
		if path != denomTrace.Path {
			return nil, fmt.Errorf("path %s does not match registered path %s for asset %s", denomTrace.GetPath(), path, denomTrace.GetBaseDenom())
		}

		// lock tokens in escrow address
		escrowAddress := types.GetEscrowAddress(req.Token.Denom)
		if err := s.bank.SendCoins(
			ctx, req.DepositAddress, escrowAddress, sdk.NewCoins(req.Token),
		); err != nil {
			return nil, err
		}

		// convert denomination from 'ibc/{hash}' to native asset that recognized by nexus module
		req.Token = sdk.NewCoin(denomTrace.GetBaseDenom(), req.Token.Amount)
		// TODO: make this public for now, we will refactor nexus module
		s.nexus.AddToChainTotal(ctx, exported.Axelarnet, req.Token)

	case req.Token.Denom == exported.Axelarnet.NativeAsset:
		// lock tokens in escrow address
		escrowAddress := types.GetEscrowAddress(req.Token.Denom)
		if err := s.bank.SendCoins(
			ctx, req.DepositAddress, escrowAddress, sdk.NewCoins(req.Token),
		); err != nil {
			return nil, err
		}

	case s.nexus.IsAssetRegistered(ctx, exported.Axelarnet.Name, req.Token.Denom):
		// transfer coins from linked address to module account and burn them
		if err := s.bank.SendCoinsFromAccountToModule(
			ctx, req.DepositAddress, types.ModuleName, sdk.NewCoins(req.Token),
		); err != nil {
			return nil, err
		}

		if err := s.bank.BurnCoins(
			ctx, types.ModuleName, sdk.NewCoins(req.Token),
		); err != nil {
			// NOTE: should not happen as the module account was
			// retrieved on the step above and it has enough balance
			// to burn.
			panic(fmt.Sprintf("cannot burn coins after a successful send to a module account: %v", err))
		}

	default:
		return nil, sdkerrors.Wrap(types.ErrAxelarnet,
			fmt.Sprintf("unrecognized %s token", req.Token.Denom))

	}

	if err := s.nexus.EnqueueForTransfer(ctx, depositAddr, req.Token); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyTxID, hex.EncodeToString(req.TxID)),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, req.DepositAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, req.Token.String()),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	return &types.ConfirmDepositResponse{}, nil
}

// ExecutePendingTransfers handles execute pending transfers
func (s msgServer) ExecutePendingTransfers(c context.Context, req *types.ExecutePendingTransfersRequest) (*types.ExecutePendingTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, types.ModuleName)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", types.ModuleName)
	}

	pendingTransfers := s.nexus.GetTransfersForChain(ctx, chain, nexus.Pending)

	if len(pendingTransfers) == 0 {
		return &types.ExecutePendingTransfersResponse{}, nil
	}

	for _, pendingTransfer := range pendingTransfers {
		recipient, err := sdk.AccAddressFromBech32(pendingTransfer.Recipient.Address)
		if err != nil {
			ctx.Logger().Debug(fmt.Sprintf("Discard invalid recipient %s and continue", pendingTransfer.Recipient.Address))
			s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
			continue
		}

		token, escrowAddress, err := prepareTransfer(ctx, s.BaseKeeper, s.nexus, s.bank, s.account, pendingTransfer)
		if err != nil {
			return nil, err
		}

		if err = s.bank.SendCoins(
			ctx, escrowAddress, recipient, sdk.NewCoins(token),
		); err != nil {
			return nil, err
		}
		s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
	}

	return &types.ExecutePendingTransfersResponse{}, nil
}

// RegisterIBCPath handles register an IBC path for a chain
func (s msgServer) RegisterIBCPath(c context.Context, req *types.RegisterIBCPathRequest) (*types.RegisterIBCPathResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if err := s.BaseKeeper.RegisterIBCPath(ctx, req.Chain, req.Path); err != nil {
		return nil, err
	}

	return &types.RegisterIBCPathResponse{}, nil
}

// AddCosmosBasedChain handles register a cosmos based chain to nexus
func (s msgServer) AddCosmosBasedChain(c context.Context, req *types.AddCosmosBasedChainRequest) (*types.AddCosmosBasedChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, found := s.nexus.GetChain(ctx, req.Chain.Name); found {
		return &types.AddCosmosBasedChainResponse{}, fmt.Errorf("chain '%s' is already registered", req.Chain.Name)
	}
	s.nexus.SetChain(ctx, req.Chain)
	s.nexus.RegisterAsset(ctx, exported.Axelarnet.Name, req.Chain.NativeAsset)
	s.nexus.RegisterAsset(ctx, req.Chain.Name, req.Chain.NativeAsset)

	s.BaseKeeper.RegisterAssetToCosmosChain(ctx, req.Chain.NativeAsset, req.Chain.Name)
	return &types.AddCosmosBasedChainResponse{}, nil
}

// RegisterAsset handles register an asset to a cosmos based chain
func (s msgServer) RegisterAsset(c context.Context, req *types.RegisterAssetRequest) (*types.RegisterAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, found := s.nexus.GetChain(ctx, req.Chain); !found {
		return &types.RegisterAssetResponse{}, fmt.Errorf("chain '%s' not found", req.Chain)
	}

	s.nexus.RegisterAsset(ctx, req.Chain, req.Denom)
	s.BaseKeeper.RegisterAssetToCosmosChain(ctx, req.Denom, req.Chain)

	return &types.RegisterAssetResponse{}, nil
}

func (s msgServer) RefundMsg(c context.Context, req *types.RefundMsgRequest) (*types.RefundMsgResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	msg := req.GetInnerMessage()
	if msg == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid inner message")
	}

	result, err := s.routeInnerMsg(ctx, msg)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to execute message")
	}

	fee, found := s.BaseKeeper.GetPendingRefund(ctx, *req)
	if found {
		// refund tx fee to the given account.
		err = s.bank.SendCoinsFromModuleToAccount(ctx, authtypes.FeeCollectorName, msg.GetSigners()[0], sdk.NewCoins(fee))
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to refund tx fee")
		}

		s.BaseKeeper.DeletePendingRefund(ctx, *req)
	}

	ctx.EventManager().EmitEvents(result.GetEvents())

	return &types.RefundMsgResponse{Data: result.Data, Log: result.Log}, nil
}

// RouteIBCTransfers routes Transfer to cosmos chains
func (s msgServer) RouteIBCTransfers(c context.Context, req *types.RouteIBCTransfersRequest) (*types.RouteIBCTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	// loop through all cosmos chains
	for _, c := range s.BaseKeeper.GetCosmosChains(ctx) {
		chain, ok := s.nexus.GetChain(ctx, c)
		if !ok {
			return &types.RouteIBCTransfersResponse{}, fmt.Errorf(fmt.Sprintf("%s is not a registered chain", chain.Name))
		}
		// Get the channel id for the chain
		path, ok := s.BaseKeeper.GetIBCPath(ctx, chain.Name)
		if !ok {
			return &types.RouteIBCTransfersResponse{}, fmt.Errorf(fmt.Sprintf("destination chain %s does not have a channel", chain.Name))
		}
		pendingTransfers := s.nexus.GetTransfersForChain(ctx, chain, nexus.Pending)
		for _, p := range pendingTransfers {
			token, sender, err := prepareTransfer(ctx, s.BaseKeeper, s.nexus, s.bank, s.account, p)
			if err != nil {
				return nil, err
			}

			err = IBCTransfer(ctx, s.BaseKeeper, s.ibcTransfer, s.ibcChannel, token, sender, p.Recipient.Address, path)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("failed to send IBC transfer %s for %s:  %e", token, p.Recipient.Address, err))
				continue
			}
			s.nexus.ArchivePendingTransfer(ctx, p)
		}
	}

	return &types.RouteIBCTransfersResponse{}, nil
}

// IBCTransfer inits an IBC transfer
func IBCTransfer(ctx sdk.Context, k types.BaseKeeper, t types.IBCTransferKeeper, c types.ChannelKeeper, token sdk.Coin, sender sdk.AccAddress, receiver string, path string) error {
	// split path to portID and channelID
	pathSplit := strings.SplitN(path, "/", 2)
	if len(pathSplit) != 2 {
		return fmt.Errorf("invalid path %s", path)
	}
	portID, channelID := pathSplit[0], pathSplit[1]
	_, state, err := c.GetChannelClientState(ctx, portID, channelID)
	if err != nil {
		return err
	}

	height := clienttypes.NewHeight(state.GetLatestHeight().GetRevisionNumber(), state.GetLatestHeight().GetRevisionHeight()+k.GetRouteTimeoutWindow(ctx))
	err = t.SendTransfer(ctx, portID, channelID, token, sender, receiver, height, 0)
	if err == nil {
		// SendTransfer would return error if the next sequence not found
		seq, _ := c.GetNextSequenceSend(ctx, portID, channelID)
		k.SetPendingIBCTransfer(ctx, portID, channelID, seq-1, types.IBCTransfer{Sender: sender, Receiver: receiver, Token: token})
	}
	return err
}

// RegisterFeeCollector handles register axelar fee collector account
func (s msgServer) RegisterFeeCollector(c context.Context, req *types.RegisterFeeCollectorRequest) (*types.RegisterFeeCollectorResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	s.BaseKeeper.SetFeeCollector(ctx, req.FeeCollector)
	return &types.RegisterFeeCollectorResponse{}, nil
}

// isIBCDenom validates that the given denomination is a valid ICS token representation (ibc/{hash})
func isIBCDenom(denom string) bool {
	if err := sdk.ValidateDenom(denom); err != nil {
		return false
	}

	denomSplit := strings.SplitN(denom, "/", 2)
	if len(denomSplit) != 2 || denomSplit[0] != ibctransfertypes.DenomPrefix {
		return false
	}
	if _, err := ibctransfertypes.ParseHexHash(denomSplit[1]); err != nil {
		return false
	}

	return true
}

// parseIBCDenom retrieves the full identifiers trace and base denomination from the IBC transfer keeper store
func (s msgServer) parseIBCDenom(ctx sdk.Context, ibcDenom string) (ibctransfertypes.DenomTrace, error) {
	denomSplit := strings.Split(ibcDenom, "/")

	hash, err := ibctransfertypes.ParseHexHash(denomSplit[1])
	if err != nil {
		return ibctransfertypes.DenomTrace{}, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid denom trace hash %s, %s", hash, err))
	}
	denomTrace, found := s.ibcTransfer.GetDenomTrace(ctx, hash)
	if !found {
		return ibctransfertypes.DenomTrace{}, status.Error(
			codes.NotFound,
			sdkerrors.Wrap(ibctransfertypes.ErrTraceNotFound, denomSplit[1]).Error(),
		)
	}
	return denomTrace, nil
}

func (s msgServer) routeInnerMsg(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {

	var msgResult *sdk.Result
	var err error

	if handler := s.msgSvcRouter.Handler(msg); handler != nil {
		// ADR 031 request type routing
		msgResult, err = handler(ctx, msg)
	} else if legacyMsg, ok := msg.(legacytx.LegacyMsg); ok {
		// legacy sdk.Msg routing
		// Assuming that the app developer has migrated all their Msgs to
		// proto messages and has registered all `Msg services`, then this
		// path should never be called, because all those Msgs should be
		// registered within the `msgServiceRouter` already.
		msgRoute := legacyMsg.Route()
		s.Logger(ctx).Debug(fmt.Sprintf("received legacy message type %s", legacyMsg.Type()))
		handler := s.router.Route(ctx, msgRoute)
		if handler == nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message route: %s", msgRoute)
		}

		msgResult, err = handler(ctx, msg)
	} else {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "can't route message %+v", msg)
	}

	return msgResult, err
}

// toICS20 converts a cross chain transfer to ICS20 token
func toICS20(ctx sdk.Context, k types.BaseKeeper, transfer nexus.CrossChainTransfer) sdk.Coin {
	path, _ := k.GetIBCPath(ctx, transfer.Recipient.Chain.Name)
	prefixedDenom := fmt.Sprintf("%s/%s", path, transfer.Asset.Denom)
	// construct the denomination trace from the full raw denomination
	denomTrace := ibctransfertypes.ParseDenomTrace(prefixedDenom)
	return sdk.NewCoin(denomTrace.IBCDenom(), transfer.Asset.Amount)
}

// isFromCosmosChain returns true if the asset origins from cosmos chains
func isFromCosmosChain(ctx sdk.Context, k types.BaseKeeper, transfer nexus.CrossChainTransfer) bool {
	chain, ok := k.GetCosmosChain(ctx, transfer.Asset.GetDenom())
	if !ok {
		return false
	}
	_, ok = k.GetIBCPath(ctx, chain)
	return ok
}

func prepareTransfer(ctx sdk.Context, k types.BaseKeeper, n types.Nexus, b types.BankKeeper, a types.AccountKeeper, transfer nexus.CrossChainTransfer) (sdk.Coin, sdk.AccAddress, error) {
	var token sdk.Coin
	var sender sdk.AccAddress
	// pending transfer can be either of cosmos chains assets, Axelarnet native asset, assets from supported chain
	switch {
	// asset origins from cosmos chains, it will be converted to ICS20 token
	case isFromCosmosChain(ctx, k, transfer):
		token = toICS20(ctx, k, transfer)
		sender = types.GetEscrowAddress(token.GetDenom())
	case transfer.Asset.Denom == exported.Axelarnet.NativeAsset:
		token = transfer.Asset
		sender = types.GetEscrowAddress(transfer.Asset.Denom)
	case n.IsAssetRegistered(ctx, exported.Axelarnet.Name, transfer.Asset.Denom):
		if err := b.MintCoins(
			ctx, types.ModuleName, sdk.NewCoins(transfer.Asset),
		); err != nil {
			return sdk.Coin{}, nil, err
		}

		token = transfer.Asset
		sender = a.GetModuleAddress(types.ModuleName)
	default:
		// should not reach here
		panic(fmt.Sprintf("unrecognized %s token", transfer.Asset.Denom))
	}

	return token, sender, nil
}
