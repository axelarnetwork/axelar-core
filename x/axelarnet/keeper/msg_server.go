package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.BaseKeeper
	nexus    types.Nexus
	bank     types.BankKeeper
	transfer types.IBCTransferKeeper
}

// NewMsgServerImpl returns an implementation of the axelarnet MsgServiceServer interface for the provided Keeper.
func NewMsgServerImpl(k types.BaseKeeper, n types.Nexus, b types.BankKeeper, t types.IBCTransferKeeper) types.MsgServiceServer {
	return msgServer{
		BaseKeeper: k,
		nexus:      n,
		bank:       b,
		transfer:   t,
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
		path := s.BaseKeeper.GetIBCPath(ctx, denomTrace.GetBaseDenom())
		if path == "" {
			return nil, fmt.Errorf("path not found for asset %s", denomTrace.GetBaseDenom())
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
		// transfer the coins from linked address to module account and burn them
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

		// pending transfer can be either of cosmos based assets from evm, Axelarnet native asset from evm, assets from supported chain
		switch {
		// if the asset has an IBC path, it will be convert to ICS20 token
		case s.BaseKeeper.GetIBCPath(ctx, pendingTransfer.Asset.Denom) != "":
			path := s.BaseKeeper.GetIBCPath(ctx, pendingTransfer.Asset.Denom)

			prefixedDenom := path + "/" + pendingTransfer.Asset.Denom
			// construct the denomination trace from the full raw denomination
			denomTrace := ibctypes.ParseDenomTrace(prefixedDenom)
			escrowAddress := types.GetEscrowAddress(denomTrace.IBCDenom())

			token := sdk.NewCoin(denomTrace.IBCDenom(), pendingTransfer.Asset.Amount)
			// unescrow source tokens. It fails if balance insufficient.
			if err := s.bank.SendCoins(
				ctx, escrowAddress, recipient, sdk.NewCoins(token),
			); err != nil {
				return nil, err
			}

		case pendingTransfer.Asset.Denom == exported.Axelarnet.NativeAsset:
			escrowAddress := types.GetEscrowAddress(pendingTransfer.Asset.Denom)
			// unescrow source tokens. It fails if balance insufficient.
			if err := s.bank.SendCoins(
				ctx, escrowAddress, recipient, sdk.NewCoins(pendingTransfer.Asset),
			); err != nil {
				return nil, err
			}

		case s.nexus.IsAssetRegistered(ctx, exported.Axelarnet.Name, pendingTransfer.Asset.Denom):
			if err := s.bank.MintCoins(
				ctx, types.ModuleName, sdk.NewCoins(pendingTransfer.Asset),
			); err != nil {
				return nil, err
			}

			if err := s.bank.SendCoinsFromModuleToAccount(
				ctx, types.ModuleName, recipient, sdk.NewCoins(pendingTransfer.Asset),
			); err != nil {
				panic(fmt.Sprintf("unable to send coins from module to account: %v", err))
			}

		default:
			// should not reach here
			panic(fmt.Sprintf("unrecognized %s token", pendingTransfer.Asset.Denom))
		}

		s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
	}

	return &types.ExecutePendingTransfersResponse{}, nil
}

// RegisterIBCPath handles register an IBC path for a asset
func (s msgServer) RegisterIBCPath(c context.Context, req *types.RegisterIBCPathRequest) (*types.RegisterIBCPathResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if err := s.BaseKeeper.RegisterIBCPath(ctx, req.Asset, req.Path); err != nil {
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

	return &types.AddCosmosBasedChainResponse{}, nil
}

// isIBCDenom validates that the given denomination is a valid ICS token representation (ibc/{hash})
func isIBCDenom(denom string) bool {
	if err := sdk.ValidateDenom(denom); err != nil {
		return false
	}

	denomSplit := strings.SplitN(denom, "/", 2)
	if len(denomSplit) != 2 || denomSplit[0] != ibctypes.DenomPrefix {
		return false
	}
	if _, err := ibctypes.ParseHexHash(denomSplit[1]); err != nil {
		return false
	}

	return true
}

// parseIBCDenom retrieves the full identifiers trace and base denomination from the IBC transfer keeper store
func (s msgServer) parseIBCDenom(ctx sdk.Context, ibcDenom string) (ibctypes.DenomTrace, error) {
	denomSplit := strings.Split(ibcDenom, "/")

	hash, err := ibctypes.ParseHexHash(denomSplit[1])
	if err != nil {
		return ibctypes.DenomTrace{}, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid denom trace hash %s, %s", hash, err))
	}
	denomTrace, found := s.transfer.GetDenomTrace(ctx, hash)
	if !found {
		return ibctypes.DenomTrace{}, status.Error(
			codes.NotFound,
			sdkerrors.Wrap(ibctypes.ErrTraceNotFound, denomSplit[1]).Error(),
		)
	}
	return denomTrace, nil
}
