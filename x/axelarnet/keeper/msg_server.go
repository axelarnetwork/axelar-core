package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var _ types.MsgServiceServer = msgServer{}

const (
	evmCallContractGasCost    = storetypes.Gas(10000000)
	cosmosCallContractGasCost = storetypes.Gas(1000000)
)

type msgServer struct {
	Keeper
	nexus   types.Nexus
	bank    types.BankKeeper
	account types.AccountKeeper
	ibcK    IBCKeeper
}

// NewMsgServerImpl returns an implementation of the axelarnet MsgServiceServer interface for the provided Keeper.
func NewMsgServerImpl(k Keeper, n types.Nexus, b types.BankKeeper, a types.AccountKeeper, ibcK IBCKeeper) types.MsgServiceServer {
	return msgServer{
		Keeper:  k,
		nexus:   n,
		bank:    b,
		account: a,
		ibcK:    ibcK,
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
	msg := nexus.NewGeneralMessage(msgID, sender, recipient, payloadHash, nexus.Approved, txID, nonce, nil)

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

		if s.bank.BlockedAddr(req.Fee.Recipient) {
			return nil, fmt.Errorf("fee recipient is a blocked address")
		}

		err := s.bank.SendCoins(ctx, req.Sender, req.Fee.Recipient, sdk.NewCoins(req.Fee.Amount))
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to transfer fee")
		}
		events.Emit(ctx, &types.FeePaid{
			MessageID: msgID,
			Recipient: req.Fee.Recipient,
			Fee:       req.Fee.Amount,
		})
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

	coin, err := NewCoin(ctx, s.ibcK, s.nexus, amount)
	if err != nil {
		return nil, err
	}

	if err := coin.Lock(s.bank, req.DepositAddress); err != nil {
		return nil, err
	}

	transferID, err := s.nexus.EnqueueForTransfer(ctx, depositAddr, coin.Coin)
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

		events.Emit(ctx,
			&types.AxelarTransferCompleted{
				ID:         pendingTransfer.ID,
				Receipient: pendingTransfer.Recipient.Address,
				Asset:      pendingTransfer.Asset,
				Recipient:  pendingTransfer.Recipient.Address,
			})

		s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
	}

	// release transfer fees
	if collector, ok := s.GetFeeCollector(ctx); ok {
		for _, fee := range s.nexus.GetTransferFees(ctx) {
			if err := transfer(ctx, s.Keeper, s.nexus, s.bank, s.account, collector, fee); err != nil {
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

		// also register on axelarnet, it routes assets from cosmos chains to evm chains
		if err := s.nexus.RegisterAsset(ctx, exported.Axelarnet, nexus.NewAsset(asset.Denom, false), utils.MaxUint, types.DefaultRateLimitWindow); err != nil {
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

	// also register on axelarnet, it routes assets from cosmos chains to evm chains
	// ignore the error in case above chain is axelarnet, or if the asset is already registered
	_ = s.nexus.RegisterAsset(ctx, exported.Axelarnet, nexus.NewAsset(req.Asset.Denom, false), utils.MaxUint, types.DefaultRateLimitWindow)

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

	if path != types.NewIBCPath(t.PortID, t.ChannelID) {
		return nil, fmt.Errorf("chain %s IBC path doesn't match %s IBC transfer path", chain.Name, path)
	}
	err := s.ibcK.SendIBCTransfer(ctx, t)
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

	msg, ok := s.nexus.GetMessage(ctx, req.ID)
	if !ok {
		return nil, fmt.Errorf("message %s not found", req.ID)
	}

	if !s.nexus.IsChainActivated(ctx, msg.Sender.Chain) {
		return nil, fmt.Errorf("chain %s is not activated", msg.GetSourceChain())
	}

	if !s.nexus.IsChainActivated(ctx, msg.Recipient.Chain) {
		return nil, fmt.Errorf("chain %s is not activated", msg.GetDestinationChain())
	}

	if msg.Type() == nexus.TypeGeneralMessageWithToken {
		funcs.MustTrue(s.nexus.IsAssetRegistered(ctx, msg.Recipient.Chain, msg.Asset.GetDenom()))
	}

	if !msg.Match(req.Payload) {
		return nil, fmt.Errorf("payload hash does not match")
	}

	if !(msg.Is(nexus.Approved) || msg.Is(nexus.Failed)) {
		return nil, fmt.Errorf("general message %s already executed", req.ID)
	}

	// send ibc message if destination is cosmos
	if msg.Recipient.Chain.IsFrom(exported.ModuleName) {
		bz, err := types.TranslateMessage(msg, req.Payload)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "invalid payload")
		}

		asset, err := s.escrowAssetToMessageSender(ctx, req.Sender, msg)
		if err != nil {
			return nil, err
		}

		err = s.ibcK.SendMessage(c, msg.Recipient, asset, string(bz), msg.ID)
		if err != nil {
			return nil, err
		}
	}

	err := s.nexus.SetMessageProcessing(ctx, msg.ID)
	if err != nil {
		return nil, err
	}

	if msg.Recipient.Chain.IsFrom(evmtypes.ModuleName) {
		ctx.GasMeter().ConsumeGas(evmCallContractGasCost, "execute-message")
	} else {
		ctx.GasMeter().ConsumeGas(cosmosCallContractGasCost, "execute-message")
	}

	s.Logger(ctx).Debug("set general message status to processing", "messageID", msg.ID)

	return &types.RouteMessageResponse{}, nil
}

// toICS20 converts a cross chain transfer to ICS20 token
func toICS20(ctx sdk.Context, k Keeper, n types.Nexus, coin sdk.Coin) sdk.Coin {
	// if chain or path not found, it will create coin with base denom
	chain, _ := n.GetChainByNativeAsset(ctx, coin.GetDenom())
	path, _ := k.GetIBCPath(ctx, chain.Name)

	prefixedDenom := types.NewIBCPath(path, coin.Denom)
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

// all general messages are sent from the Axelar general message sender, so receiver can use the packet sender to authenticate the message
// escrowAssetToMessageSender sends the asset to general msg sender account
func (s msgServer) escrowAssetToMessageSender(ctx sdk.Context, reqSender sdk.AccAddress, msg nexus.GeneralMessage) (sdk.Coin, error) {
	var asset sdk.Coin
	var acc sdk.AccAddress
	var err error

	switch msg.Type() {
	case nexus.TypeGeneralMessage:
		// pure general message, take dust amount from request sender to satisfy ibc transfer requirements
		asset = sdk.NewCoin(exported.NativeAsset, sdk.NewInt(1))
		acc = reqSender
	case nexus.TypeGeneralMessageWithToken:
		// general message with token, get token from corresponding account
		asset, acc, err = prepareTransfer(ctx, s.Keeper, s.nexus, s.bank, s.account, *msg.Asset)
		if err != nil {
			return sdk.Coin{}, err
		}
	default:
		return sdk.Coin{}, fmt.Errorf("unrecognized message type")
	}

	// use GeneralMessageSender account as the canonical general message sender
	err = s.bank.SendCoins(ctx, acc, types.AxelarGMPAccount, sdk.NewCoins(asset))
	if err != nil {
		return sdk.Coin{}, err
	}

	return asset, nil
}
