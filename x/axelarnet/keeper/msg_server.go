package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
	nexus types.Nexus
	bank  types.BankKeeper
	ibcK  IBCKeeper
}

// NewMsgServerImpl returns an implementation of the axelarnet MsgServiceServer interface for the provided Keeper.
func NewMsgServerImpl(k Keeper, n types.Nexus, b types.BankKeeper, ibcK IBCKeeper) types.MsgServiceServer {
	return msgServer{
		Keeper: k,
		nexus:  n,
		bank:   b,
		ibcK:   ibcK,
	}
}

func (s msgServer) CallContract(c context.Context, req *types.CallContractRequest) (*types.CallContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if !s.nexus.IsChainActivated(ctx, exported.Axelarnet) {
		return nil, fmt.Errorf("chain %s is not activated yet", exported.Axelarnet.Name)
	}

	if !s.nexus.IsChainActivated(ctx, chain) {
		return nil, fmt.Errorf("chain %s is not activated yet", chain.Name)
	}

	recipient := nexus.CrossChainAddress{Chain: chain, Address: req.ContractAddress}
	if err := s.nexus.ValidateAddress(ctx, recipient); err != nil {
		return nil, err
	}

	sender := nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: req.Sender.String()}

	// axelar gateway expects keccak256 hashes for payloads
	payloadHash := crypto.Keccak256(req.Payload)

	msgID, txID, nonce := s.nexus.GenerateMessageID(ctx)
	msg := nexus.NewGeneralMessage(msgID, sender, recipient, payloadHash, txID, nonce, nil)

	events.Emit(ctx, &types.ContractCallSubmitted{
		MessageID:        msg.ID,
		Sender:           msg.GetSourceAddress(),
		SourceChain:      msg.GetSourceChain(),
		DestinationChain: msg.GetDestinationChain(),
		ContractAddress:  msg.GetDestinationAddress(),
		PayloadHash:      msg.PayloadHash,
		Payload:          req.Payload,
	})

	if req.Fee != nil {
		lockableAsset, err := s.nexus.NewLockableAsset(ctx, s.ibcK, s.bank, req.Fee.Amount)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "unrecognized fee denom")
		}

		err = s.bank.SendCoins(ctx, req.Sender, req.Fee.Recipient, sdk.NewCoins(req.Fee.Amount))
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to transfer fee")
		}

		feePaidEvent := types.FeePaid{
			MessageID:        msgID,
			Recipient:        req.Fee.Recipient,
			Fee:              req.Fee.Amount,
			Asset:            lockableAsset.GetAsset().Denom,
			SourceChain:      msg.GetSourceChain(),
			DestinationChain: msg.GetDestinationChain(),
		}
		if req.Fee.RefundRecipient != nil {
			feePaidEvent.RefundRecipient = req.Fee.RefundRecipient.String()
		}
		events.Emit(ctx, &feePaidEvent)
	}

	if err := s.nexus.SetNewMessage(ctx, msg); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to add general message")
	}

	s.Logger(ctx).Debug(fmt.Sprintf("successfully enqueued contract call for contract address %s on chain %s from sender %s with message id %s", req.ContractAddress, req.Chain.String(), req.Sender, msg.ID),
		types.AttributeKeyDestinationChain, req.Chain.String(),
		types.AttributeKeyDestinationAddress, req.ContractAddress,
		types.AttributeKeySourceAddress, req.Sender,
		types.AttributeKeyMessageID, msg.ID,
		types.AttributeKeyPayloadHash, hex.EncodeToString(payloadHash),
	)

	return &types.CallContractResponse{}, nil
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
	coin := s.bank.SpendableBalance(ctx, req.DepositAddress, req.Denom)
	if coin.IsZero() {
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

	lockableAsset, err := s.nexus.NewLockableAsset(ctx, s.ibcK, s.bank, coin)
	if err != nil {
		return nil, err
	}

	if err := lockableAsset.LockFrom(ctx, req.DepositAddress); err != nil {
		return nil, err
	}

	transferID, err := s.nexus.EnqueueForTransfer(ctx, depositAddr, lockableAsset.GetAsset())
	if err != nil {
		return nil, err
	}

	s.Logger(ctx).Info(fmt.Sprintf("deposit confirmed for %s on chain %s to recipient %s on chain %s for asset %s with transfer ID %d", req.DepositAddress.String(), exported.Axelarnet.Name, recipient.Address, recipient.Chain.Name, coin.String(), transferID),
		sdk.AttributeKeyAction, types.AttributeValueConfirm,
		types.AttributeKeySourceChain, exported.Axelarnet.Name,
		types.AttributeKeyDepositAddress, req.DepositAddress.String(),
		types.AttributeKeyDestinationChain, recipient.Chain.Name,
		types.AttributeKeyDestinationAddress, recipient.Address,
		sdk.AttributeKeyAmount, coin.String(),
		types.AttributeKeyAsset, coin.Denom,
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
			sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
			sdk.NewAttribute(types.AttributeKeyAsset, coin.Denom),
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

	transferLimit := s.Keeper.GetTransferLimit(ctx)
	pageRequest := &query.PageRequest{
		Key:        nil,
		Offset:     0,
		Limit:      transferLimit,
		CountTotal: false,
		Reverse:    false,
	}
	pendingTransfers, _, err := s.nexus.GetTransfersForChainPaginated(ctx, chain, nexus.Pending, pageRequest)
	if err != nil {
		return nil, err
	}

	if len(pendingTransfers) == 0 {
		s.Logger(ctx).Debug("no pending transfers found")
		return &types.ExecutePendingTransfersResponse{}, nil
	}

	for _, pendingTransfer := range pendingTransfers {
		success := utils.RunCached(ctx, s, func(ctx sdk.Context) (bool, error) {
			recipient, err := sdk.AccAddressFromBech32(pendingTransfer.Recipient.Address)
			if err != nil {
				// NOTICE: Addresses that previously failed validation were marked as Archived. Starting in v1.1, they are now marked as TransferFailed.
				s.Logger(ctx).Error(fmt.Sprintf("transfer failed due to invalid recipient %s", pendingTransfer.Recipient.Address))
				return false, err
			}

			if err = transfer(ctx, s.Keeper, s.nexus, s.bank, s.ibcK, recipient, pendingTransfer.Asset); err != nil {
				s.Logger(ctx).Error("failed to transfer asset to axelarnet", "err", err)
				return false, err
			}

			s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
			events.Emit(ctx,
				&types.AxelarTransferCompleted{
					ID:         pendingTransfer.ID,
					Receipient: pendingTransfer.Recipient.Address,
					Asset:      pendingTransfer.Asset,
					Recipient:  pendingTransfer.Recipient.Address,
				})

			return true, nil
		})

		if !success {
			s.nexus.MarkTransferAsFailed(ctx, pendingTransfer)
		}
	}

	// release transfer fees
	if collector, ok := s.GetFeeCollector(ctx); ok {
		for _, fee := range s.nexus.GetTransferFees(ctx) {
			if err := transfer(ctx, s.Keeper, s.nexus, s.bank, s.ibcK, collector, fee); err != nil {
				s.Logger(ctx).Error("failed to collect fees", "err", err)
				continue
			}

			events.Emit(ctx,
				&types.FeeCollected{
					Collector: collector,
					Fee:       fee,
				})

			s.nexus.SubTransferFee(ctx, fee)
		}
	}

	return &types.ExecutePendingTransfersResponse{}, nil
}

// AddCosmosBasedChain handles register a cosmos based chain to nexus
func (s msgServer) AddCosmosBasedChain(c context.Context, req *types.AddCosmosBasedChainRequest) (*types.AddCosmosBasedChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, found := s.nexus.GetChain(ctx, req.CosmosChain); found {
		return nil, fmt.Errorf("chain '%s' is already registered", req.CosmosChain)
	}

	if chain, found := s.GetChainNameByIBCPath(ctx, req.IBCPath); found {
		return nil, fmt.Errorf("ibc path %s is already registered for chain %s", req.IBCPath, chain)
	}

	chain := nexus.Chain{
		Name:                  req.CosmosChain,
		KeyType:               tss.None,
		SupportsForeignAssets: true,
		Module:                types.ModuleName,
	}
	s.nexus.SetChain(ctx, chain)

	// register asset in chain state
	for _, asset := range req.NativeAssets {
		if err := s.nexus.RegisterAsset(ctx, chain, asset, utils.MaxUint, types.DefaultRateLimitWindow); err != nil {
			return nil, err
		}
	}

	if err := s.SetCosmosChain(ctx, types.CosmosChain{
		Name:       chain.Name,
		IBCPath:    req.IBCPath,
		AddrPrefix: req.AddrPrefix,
	}); err != nil {
		return nil, err
	}

	if err := s.SetChainByIBCPath(ctx, req.IBCPath, chain.Name); err != nil {
		return nil, err
	}

	return &types.AddCosmosBasedChainResponse{}, nil
}

// RegisterAsset handles register an asset to a cosmos based chain
func (s msgServer) RegisterAsset(c context.Context, req *types.RegisterAssetRequest) (*types.RegisterAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, found := s.nexus.GetChain(ctx, req.Chain)
	if !found {
		return nil, fmt.Errorf("chain '%s' not found", req.Chain)
	}

	if _, found := s.GetCosmosChainByName(ctx, req.Chain); !found {
		return nil, fmt.Errorf("chain '%s' is not a cosmos chain", req.Chain)
	}

	// register asset in chain state
	err := s.nexus.RegisterAsset(ctx, chain, req.Asset, req.Limit, req.Window)
	if err != nil {
		return nil, err
	}

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

		transferLimit := s.Keeper.GetTransferLimit(ctx)
		pageRequest := &query.PageRequest{
			Key:        nil,
			Offset:     0,
			Limit:      transferLimit,
			CountTotal: false,
			Reverse:    false,
		}
		pendingTransfers, _, err := s.nexus.GetTransfersForChainPaginated(ctx, chain, nexus.Pending, pageRequest)
		if err != nil {
			return nil, err
		}
		for _, p := range pendingTransfers {
			lockableAsset, err := s.nexus.NewLockableAsset(ctx, s.ibcK, s.bank, p.Asset)
			if err != nil {
				s.Logger(ctx).Error(fmt.Sprintf("failed to route IBC transfer %s: %s", p.String(), err))
				continue
			}

			if err := lockableAsset.UnlockTo(ctx, types.AxelarIBCAccount); err != nil {
				s.Logger(ctx).Error(fmt.Sprintf("failed to route IBC transfer %s: %s", p.String(), err))
				continue
			}

			funcs.MustNoErr(s.EnqueueIBCTransfer(ctx, types.NewIBCTransfer(types.AxelarIBCAccount, p.Recipient.Address, lockableAsset.GetCoin(ctx), portID, channelID, p.ID)))
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

	t, ok := s.GetTransfer(ctx, req.ID)
	if !ok {
		return nil, fmt.Errorf("transfer %s not found", req.ID.String())
	}

	if t.Status != types.TransferFailed {
		return nil, fmt.Errorf("IBC transfer %s does not have failed status", req.ID.String())
	}

	path := types.NewIBCPath(t.PortID, t.ChannelID)
	chainName, ok := s.GetChainNameByIBCPath(ctx, path)
	if !ok {
		return nil, fmt.Errorf("no cosmos chain registered for ibc path %s", path)
	}

	chain, ok := s.nexus.GetChain(ctx, chainName)
	if !ok {
		return nil, fmt.Errorf("invalid chain %s", chainName)
	}

	if !s.nexus.IsChainActivated(ctx, chain) {
		return nil, fmt.Errorf("chain %s is not activated", chain.Name)
	}

	lockableAsset, err := s.nexus.NewLockableAsset(ctx, s.ibcK, s.bank, t.Token)
	if err != nil {
		return nil, err
	}

	err = lockableAsset.UnlockTo(ctx, types.AxelarIBCAccount)
	if err != nil {
		return nil, err
	}

	// Note: Starting from version 1.1, all IBC transfers are routed through AxelarIBCAccount,
	// and all previously failed transfers have been migrated to use AxelarIBCAccount as the sender.
	//
	// This is a temporary measure to prevent pending transfers that would fail during the upgrade process.
	// It can be removed if no such cases exists during upgrade, or the migration can be re-run to update senders in version 1.2.
	t.Sender = types.AxelarIBCAccount

	err = s.ibcK.SendIBCTransfer(ctx, t)
	if err != nil {
		return nil, err
	}

	funcs.MustNoErr(s.SetTransferPending(ctx, t.ID))

	events.Emit(ctx,
		&types.IBCTransferRetried{
			ID:         t.ID,
			Receipient: t.Receiver,
			Asset:      t.Token,
			Sequence:   t.Sequence,
			PortID:     t.PortID,
			ChannelID:  t.ChannelID,
			Recipient:  t.Receiver,
		})

	return &types.RetryIBCTransferResponse{}, nil
}

// RouteMessage calls IBC for cosmos messages or updates the state if the message in the nexus module for any other kind
func (s msgServer) RouteMessage(c context.Context, req *types.RouteMessageRequest) (*types.RouteMessageResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	routingCtx := nexus.RoutingContext{
		Sender:     req.Sender,
		FeeGranter: req.Feegranter,
		Payload:    req.Payload,
	}
	if err := s.nexus.RouteMessage(ctx, req.ID, routingCtx); err != nil {
		return nil, err
	}

	return &types.RouteMessageResponse{}, nil
}

func transfer(ctx sdk.Context, k Keeper, n types.Nexus, b types.BankKeeper, ibc types.IBCKeeper, recipient sdk.AccAddress, coin sdk.Coin) error {
	lockableAsset, err := n.NewLockableAsset(ctx, ibc, b, coin)
	if err != nil {
		return err
	}

	if err := lockableAsset.UnlockTo(ctx, recipient); err != nil {
		return err
	}

	k.Logger(ctx).Debug(fmt.Sprintf("successfully sent %s to %s", coin, recipient))

	return nil
}
