package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	nexus types.Nexus
	bank  types.BankKeeper
}

// NewMsgServerImpl returns an implementation of the axelarnet MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(n types.Nexus, b types.BankKeeper) types.MsgServiceServer {
	return msgServer{
		nexus: n,
		bank:  b,
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

	linkedAddress := types.NewLinkedAddress(recipientChain.Name, req.Asset, req.RecipientAddr)
	s.nexus.LinkAddresses(ctx,
		nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: linkedAddress.String()},
		nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr})

	return &types.LinkResponse{DepositAddr: linkedAddress.String()}, nil
}

// ConfirmDeposit handles deposit confirmations
func (s msgServer) ConfirmDeposit(c context.Context, req *types.ConfirmDepositRequest) (*types.ConfirmDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// transfer the coins from linked address to module account and burn them
	if err := s.bank.SendCoinsFromAccountToModule(
		ctx, req.BurnerAddress, types.ModuleName, sdk.NewCoins(req.Token),
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

	depositAddr := nexus.CrossChainAddress{Address: req.BurnerAddress.String(), Chain: exported.Axelarnet}
	if err := s.nexus.EnqueueForTransfer(ctx, depositAddr, req.Token); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyTxID, hex.EncodeToString(req.TxID)),
			sdk.NewAttribute(types.AttributeKeyBurnAddress, req.BurnerAddress.String()),
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
			return nil, err
		}

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

		s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
	}

	return &types.ExecutePendingTransfersResponse{}, nil
}
