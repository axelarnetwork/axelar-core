package keeper

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
	nexus       types.Nexus
	bank        types.BankKeeper
	ibcTransfer types.IBCTransferKeeper
	account     types.AccountKeeper
	ibcK        IBCKeeper
}

// NewMsgServerImpl returns an implementation of the axelarnet MsgServiceServer interface for the provided Keeper.
func NewMsgServerImpl(k Keeper, n types.Nexus, b types.BankKeeper, t types.IBCTransferKeeper, a types.AccountKeeper, ibcK IBCKeeper) types.MsgServiceServer {
	return msgServer{
		Keeper:      k,
		nexus:       n,
		bank:        b,
		ibcTransfer: t,
		account:     a,
		ibcK:        ibcK,
	}
}

// Link handles address linking
func (s msgServer) Link(c context.Context, req *types.LinkRequest) (*types.LinkResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	recipientChain, ok := s.nexus.GetChain(ctx, req.RecipientChain)
	if !ok {
		return nil, fmt.Errorf("unknown recipient chain")
	}

	found := s.nexus.IsAssetRegistered(ctx, recipientChain, req.Asset)
	if !found {
		return nil, fmt.Errorf("asset '%s' not registered for chain '%s'", req.Asset, recipientChain.Name)
	}

	recipient := nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr}
	depositAddress := types.NewLinkedAddress(ctx, recipientChain.Name, req.Asset, req.RecipientAddr)
	if err := s.nexus.LinkAddresses(ctx,
		nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: depositAddress.String()},
		recipient,
	); err != nil {
		return nil, fmt.Errorf("could not link addresses: %s", err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLink,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeySourceChain, exported.Axelarnet.Name.String()),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, depositAddress.String()),
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipientChain.Name.String()),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, req.RecipientAddr),
			sdk.NewAttribute(types.AttributeKeyAsset, req.Asset),
		),
	)

	s.Logger(ctx).Debug(fmt.Sprintf("successfully linked deposit %s on chain %s to recipient %s on chain %s for asset %s", depositAddress.String(), exported.Axelarnet.Name, req.RecipientAddr, req.RecipientChain, req.Asset),
		types.AttributeKeySourceChain, exported.Axelarnet.Name,
		types.AttributeKeyDepositAddress, depositAddress.String(),
		types.AttributeKeyDestinationChain, recipientChain.Name,
		types.AttributeKeyDestinationAddress, req.RecipientAddr,
		types.AttributeKeyAsset, req.Asset,
	)

	return &types.LinkResponse{DepositAddr: depositAddress.String()}, nil
}

// ConfirmDeposit handles deposit confirmations
func (s msgServer) ConfirmDeposit(c context.Context, req *types.ConfirmDepositRequest) (*types.ConfirmDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	depositAddr := nexus.CrossChainAddress{Address: req.DepositAddress.String(), Chain: exported.Axelarnet}
	amount := s.bank.GetBalance(ctx, req.DepositAddress, req.Denom)
	if amount.IsZero() {
		return nil, fmt.Errorf("deposit address '%s' holds no funds for token %s", req.DepositAddress.String(), req.Denom)
	}

	recipient, ok := s.nexus.GetRecipient(ctx, depositAddr)
	if !ok {
		return nil, fmt.Errorf("no recipient linked to deposit address %s", req.DepositAddress.String())
	}

	if !s.nexus.IsChainActivated(ctx, exported.Axelarnet) {
		return nil, fmt.Errorf("source chain '%s'  is not activated", exported.Axelarnet.Name)
	}

	if !s.nexus.IsChainActivated(ctx, recipient.Chain) {
		return nil, fmt.Errorf("recipient chain '%s' is not activated", recipient.Chain.Name)
	}

	// deposit can be either of ICS 20 token from cosmos based chains, Axelarnet native asset, and wrapped asset from supported chain
	switch {
	// check if the format of token denomination is 'ibc/{hash}'
	case isIBCDenom(req.Denom):
		// get base denomination and tracing path
		denomTrace, err := s.parseIBCDenom(ctx, req.Denom)
		if err != nil {
			return nil, err
		}

		// check if the asset registered with a path
		chain, ok := s.nexus.GetChainByNativeAsset(ctx, denomTrace.GetBaseDenom())
		if !ok {
			return nil, fmt.Errorf("asset %s is not linked to a cosmos chain", denomTrace.GetBaseDenom())
		}
		path, ok := s.GetIBCPath(ctx, chain.Name)
		if !ok {
			return nil, fmt.Errorf("path not found for chain %s", chain.Name)
		}
		if path != denomTrace.Path {
			return nil, fmt.Errorf("path %s does not match registered path %s for asset %s", denomTrace.GetPath(), path, denomTrace.GetBaseDenom())
		}

		// lock tokens in escrow address
		escrowAddress := types.GetEscrowAddress(req.Denom)
		if err := s.bank.SendCoins(
			ctx, req.DepositAddress, escrowAddress, sdk.NewCoins(amount),
		); err != nil {
			return nil, err
		}

		// convert denomination from 'ibc/{hash}' to native asset that recognized by nexus module
		amount = sdk.NewCoin(denomTrace.GetBaseDenom(), amount.Amount)

	case isNativeAssetOnAxelarnet(ctx, s.nexus, req.Denom):
		// lock tokens in escrow address
		escrowAddress := types.GetEscrowAddress(req.Denom)
		if err := s.bank.SendCoins(
			ctx, req.DepositAddress, escrowAddress, sdk.NewCoins(amount),
		); err != nil {
			return nil, err
		}

	case s.nexus.IsAssetRegistered(ctx, exported.Axelarnet, req.Denom):
		// transfer coins from linked address to module account and burn them
		if err := s.bank.SendCoinsFromAccountToModule(
			ctx, req.DepositAddress, types.ModuleName, sdk.NewCoins(amount),
		); err != nil {
			return nil, err
		}

		if err := s.bank.BurnCoins(
			ctx, types.ModuleName, sdk.NewCoins(amount),
		); err != nil {
			// NOTE: should not happen as the module account was
			// retrieved on the step above and it has enough balance
			// to burn.
			panic(fmt.Sprintf("cannot burn coins after a successful send to a module account: %v", err))
		}

	default:
		return nil, sdkerrors.Wrap(types.ErrAxelarnet,
			fmt.Sprintf("unrecognized %s token", req.Denom))

	}

	transferID, err := s.nexus.EnqueueForTransfer(ctx, depositAddr, amount)
	if err != nil {
		return nil, err
	}

	s.Logger(ctx).Info(fmt.Sprintf("deposit confirmed for %s on chain %s to recipient %s on chain %s for asset %s with transfer ID %d", req.DepositAddress.String(), exported.Axelarnet.Name, recipient.Address, recipient.Chain.Name, amount.String(), transferID),
		sdk.AttributeKeyAction, types.AttributeValueConfirm,
		types.AttributeKeySourceChain, exported.Axelarnet.Name,
		types.AttributeKeyDepositAddress, req.DepositAddress.String(),
		types.AttributeKeyDestinationChain, recipient.Chain.Name,
		types.AttributeKeyDestinationAddress, recipient.Address,
		sdk.AttributeKeyAmount, amount.String(),
		types.AttributeKeyAsset, amount.Denom,
		types.AttributeKeyTransferID, transferID.String(),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
			sdk.NewAttribute(types.AttributeKeySourceChain, exported.Axelarnet.Name.String()),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, req.DepositAddress.String()),
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipient.Chain.Name.String()),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, recipient.Address),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyAsset, amount.Denom),
			sdk.NewAttribute(types.AttributeKeyTransferID, transferID.String()),
		))

	return &types.ConfirmDepositResponse{}, nil
}

// ExecutePendingTransfers handles execute pending transfers
func (s msgServer) ExecutePendingTransfers(c context.Context, _ *types.ExecutePendingTransfersRequest) (*types.ExecutePendingTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, types.ModuleName)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", types.ModuleName)
	}

	pendingTransfers := s.nexus.GetTransfersForChain(ctx, chain, nexus.Pending)

	if len(pendingTransfers) == 0 {
		s.Logger(ctx).Debug("no pending transfers found")
		return &types.ExecutePendingTransfersResponse{}, nil
	}

	for _, pendingTransfer := range pendingTransfers {
		recipient, err := sdk.AccAddressFromBech32(pendingTransfer.Recipient.Address)
		if err != nil {
			s.Logger(ctx).Error(fmt.Sprintf("discard invalid recipient %s and continue", pendingTransfer.Recipient.Address))
			s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
			continue
		}

		if err = transfer(ctx, s.Keeper, s.nexus, s.bank, s.account, recipient, pendingTransfer.Asset); err != nil {
			s.Logger(ctx).Error("failed to transfer asset to axelarnet", "err", err)
			continue
		}

		funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(
			&types.AxelarTransferCompleted{
				ID:         pendingTransfer.ID,
				Receipient: pendingTransfer.Recipient.Address,
				Asset:      pendingTransfer.Asset,
			}))

		s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
	}

	// release transfer fees
	if collector, ok := s.GetFeeCollector(ctx); ok {
		for _, fee := range s.nexus.GetTransferFees(ctx) {
			if err := transfer(ctx, s.Keeper, s.nexus, s.bank, s.account, collector, fee); err != nil {
				s.Logger(ctx).Error("failed to collect fees", "err", err)
				continue
			}

			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(
				&types.FeeCollected{
					Collector: collector,
					Fee:       fee,
				}))

			s.nexus.SubTransferFee(ctx, fee)
		}
	}

	return &types.ExecutePendingTransfersResponse{}, nil
}

// RegisterIBCPath handles register an IBC path for a chain
func (s msgServer) RegisterIBCPath(c context.Context, req *types.RegisterIBCPathRequest) (*types.RegisterIBCPathResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if err := s.SetIBCPath(ctx, req.Chain, req.Path); err != nil {
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

	// register asset in chain state
	for _, asset := range req.NativeAssets {
		if err := s.nexus.RegisterAsset(ctx, req.Chain, asset); err != nil {
			return nil, err
		}

		// also register on axelarnet, it routes assets from cosmos chains to evm chains
		if err := s.nexus.RegisterAsset(ctx, exported.Axelarnet, nexus.NewAsset(asset.Denom, false)); err != nil {
			return nil, err
		}
	}

	s.SetCosmosChain(ctx, types.CosmosChain{
		Name:       req.Chain.Name,
		AddrPrefix: req.AddrPrefix,
	})

	return &types.AddCosmosBasedChainResponse{}, nil
}

// RegisterAsset handles register an asset to a cosmos based chain
func (s msgServer) RegisterAsset(c context.Context, req *types.RegisterAssetRequest) (*types.RegisterAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, found := s.nexus.GetChain(ctx, req.Chain)
	if !found {
		return &types.RegisterAssetResponse{}, fmt.Errorf("chain '%s' not found", req.Chain)
	}

	if _, found := s.GetCosmosChainByName(ctx, req.Chain); !found {
		return &types.RegisterAssetResponse{}, fmt.Errorf("chain '%s' is not a cosmos chain", req.Chain)
	}

	// register asset in chain state
	err := s.nexus.RegisterAsset(ctx, chain, req.Asset)
	if err != nil {
		return nil, err
	}

	// also register on axelarnet, it routes assets from cosmos chains to evm chains
	// ignore the error in case above chain is axelarnet, or if the asset is already registered
	_ = s.nexus.RegisterAsset(ctx, exported.Axelarnet, nexus.NewAsset(req.Asset.Denom, false))

	return &types.RegisterAssetResponse{}, nil
}

// RouteIBCTransfers routes transfer to cosmos chains
func (s msgServer) RouteIBCTransfers(c context.Context, _ *types.RouteIBCTransfersRequest) (*types.RouteIBCTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	// loop through all cosmos chains
	for _, c := range s.GetCosmosChains(ctx) {
		chain, ok := s.nexus.GetChain(ctx, c)
		if !ok {
			s.Logger(ctx).Error(fmt.Sprintf("%s is not a registered chain", chain.Name))
			continue
		}

		if chain.Name.Equals(exported.Axelarnet.Name) {
			continue
		}

		// Get the channel id for the chain
		path, ok := s.GetIBCPath(ctx, chain.Name)
		if !ok {
			s.Logger(ctx).Error(fmt.Sprintf("%s is not a registered chain", chain.Name))
			continue
		}

		pathSplit := strings.SplitN(path, "/", 2)
		if len(pathSplit) != 2 {
			s.Logger(ctx).Error(fmt.Sprintf("invalid path %s for chain %s", path, chain.Name))
			continue
		}
		portID, channelID := pathSplit[0], pathSplit[1]

		pendingTransfers := s.nexus.GetTransfersForChain(ctx, chain, nexus.Pending)
		for _, p := range pendingTransfers {
			token, sender, err := prepareTransfer(ctx, s.Keeper, s.nexus, s.bank, s.account, p.Asset)
			if err != nil {
				s.Logger(ctx).Error(fmt.Sprintf("failed to prepare transfer %s: %s", p.String(), err))
				continue
			}

			funcs.MustNoErr(s.EnqueueIBCTransfer(ctx, types.NewIBCTransfer(sender, p.Recipient.Address, token, portID, channelID, p.ID)))
			s.nexus.ArchivePendingTransfer(ctx, p)
		}
	}

	return &types.RouteIBCTransfersResponse{}, nil
}

// RegisterFeeCollector handles register axelar fee collector account
func (s msgServer) RegisterFeeCollector(c context.Context, req *types.RegisterFeeCollectorRequest) (*types.RegisterFeeCollectorResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := s.SetFeeCollector(ctx, req.FeeCollector); err != nil {
		return nil, err
	}

	return &types.RegisterFeeCollectorResponse{}, nil
}

// RetryIBCTransfer handles retry a failed IBC transfer
func (s msgServer) RetryIBCTransfer(c context.Context, req *types.RetryIBCTransferRequest) (*types.RetryIBCTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("invalid chain %s", req.Chain)
	}

	if !s.nexus.IsChainActivated(ctx, chain) {
		return nil, fmt.Errorf("chain %s is not activated", chain.Name)
	}

	path, ok := s.GetIBCPath(ctx, chain.Name)
	if !ok {
		return nil, fmt.Errorf("%s does not have a valid IBC path", chain.Name)
	}

	t, ok := s.GetTransfer(ctx, req.ID)
	if !ok {
		return nil, fmt.Errorf("transfer %s not found", req.ID.String())
	}

	if t.Status != types.TransferFailed {
		return nil, fmt.Errorf("IBC transfer %s does not have failed status", req.ID.String())
	}

	if path != fmt.Sprintf("%s/%s", t.PortID, t.ChannelID) {
		return nil, fmt.Errorf("chain %s IBC path doesn't match %s IBC transfer path", chain.Name, path)
	}
	err := s.ibcK.SendIBCTransfer(ctx, t)
	if err != nil {
		return nil, err
	}

	funcs.MustNoErr(s.SetTransferPending(ctx, t.ID))

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(
		&types.IBCTransferSent{
			ID:         t.ID,
			Receipient: t.Receiver,
			Asset:      t.Token,
			Sequence:   t.Sequence,
			PortID:     t.PortID,
			ChannelID:  t.ChannelID,
		}))

	return &types.RetryIBCTransferResponse{}, nil
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

func isNativeAssetOnAxelarnet(ctx sdk.Context, n types.Nexus, denom string) bool {
	chain, ok := n.GetChainByNativeAsset(ctx, denom)
	return ok && chain.Name == exported.Axelarnet.Name
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

// toICS20 converts a cross chain transfer to ICS20 token
func toICS20(ctx sdk.Context, k Keeper, n types.Nexus, coin sdk.Coin) sdk.Coin {
	// if chain or path not found, it will create coin with base denom
	chain, _ := n.GetChainByNativeAsset(ctx, coin.GetDenom())
	path, _ := k.GetIBCPath(ctx, chain.Name)

	prefixedDenom := fmt.Sprintf("%s/%s", path, coin.Denom)
	// construct the denomination trace from the full raw denomination
	denomTrace := ibctransfertypes.ParseDenomTrace(prefixedDenom)
	return sdk.NewCoin(denomTrace.IBCDenom(), coin.Amount)
}

// isFromExternalCosmosChain returns true if the asset origins from cosmos chains
func isFromExternalCosmosChain(ctx sdk.Context, k Keeper, n types.Nexus, coin sdk.Coin) bool {
	chain, ok := n.GetChainByNativeAsset(ctx, coin.GetDenom())
	if !ok {
		return false
	}

	if _, ok = k.GetCosmosChainByName(ctx, chain.Name); !ok {
		return false
	}

	_, ok = k.GetIBCPath(ctx, chain.Name)
	return ok
}

func transfer(ctx sdk.Context, k Keeper, n types.Nexus, b types.BankKeeper, a types.AccountKeeper, recipient sdk.AccAddress, coin sdk.Coin) error {
	coin, escrowAddress, err := prepareTransfer(ctx, k, n, b, a, coin)
	if err != nil {
		return fmt.Errorf("failed to prepare transfer %s: %s", coin, err)
	}

	if err = b.SendCoins(
		ctx, escrowAddress, recipient, sdk.NewCoins(coin),
	); err != nil {
		return fmt.Errorf("failed to send %s from %s to %s: %s", coin, escrowAddress, recipient, err)
	}
	k.Logger(ctx).Debug(fmt.Sprintf("successfully sent %s from %s to %s", coin, escrowAddress, recipient))

	return nil
}

func prepareTransfer(ctx sdk.Context, k Keeper, n types.Nexus, b types.BankKeeper, a types.AccountKeeper, coin sdk.Coin) (sdk.Coin, sdk.AccAddress, error) {
	var sender sdk.AccAddress
	// pending transfer can be either of cosmos chains assets, Axelarnet native asset, assets from supported chain
	switch {
	// asset origins from cosmos chains, it will be converted to ICS20 token
	case isFromExternalCosmosChain(ctx, k, n, coin):
		coin = toICS20(ctx, k, n, coin)
		sender = types.GetEscrowAddress(coin.GetDenom())
	case isNativeAssetOnAxelarnet(ctx, n, coin.Denom):
		sender = types.GetEscrowAddress(coin.Denom)
	case n.IsAssetRegistered(ctx, exported.Axelarnet, coin.Denom):
		if err := b.MintCoins(
			ctx, types.ModuleName, sdk.NewCoins(coin),
		); err != nil {
			return sdk.Coin{}, nil, err
		}

		sender = a.GetModuleAddress(types.ModuleName)
	default:
		// should not reach here
		panic(fmt.Sprintf("unrecognized %s token", coin.Denom))
	}

	return coin, sender, nil
}
