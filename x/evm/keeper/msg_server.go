package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.BaseKeeper
	nexus          types.Nexus
	voter          types.Voter
	snapshotter    types.Snapshotter
	staking        types.StakingKeeper
	slashing       types.SlashingKeeper
	multisigKeeper types.MultisigKeeper
}

// NewMsgServerImpl returns an implementation of the evm MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper types.BaseKeeper, n types.Nexus, v types.Voter, snap types.Snapshotter, staking types.StakingKeeper, slashing types.SlashingKeeper, multisigKeeper types.MultisigKeeper) types.MsgServiceServer {
	return msgServer{
		BaseKeeper:     keeper,
		nexus:          n,
		voter:          v,
		snapshotter:    snap,
		staking:        staking,
		slashing:       slashing,
		multisigKeeper: multisigKeeper,
	}
}

func validateChainActivated(ctx sdk.Context, n types.Nexus, chain nexus.Chain) error {
	if !n.IsChainActivated(ctx, chain) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest,
			fmt.Sprintf("chain %s is not activated yet", chain.Name))
	}

	return nil
}

func excludeJailedOrTombstoned(ctx sdk.Context, slashing types.SlashingKeeper, snapshotter types.Snapshotter) func(v snapshot.ValidatorI) bool {
	isTombstoned := func(v snapshot.ValidatorI) bool {
		consAdd, err := v.GetConsAddr()
		if err != nil {
			return true
		}

		return slashing.IsTombstoned(ctx, consAdd)
	}

	isProxyActive := func(v snapshot.ValidatorI) bool {
		_, isActive := snapshotter.GetProxy(ctx, v.GetOperator())

		return isActive
	}

	return funcs.And(
		snapshot.ValidatorI.IsBonded,
		funcs.Not(snapshot.ValidatorI.IsJailed),
		funcs.Not(isTombstoned),
		isProxyActive,
	)
}

func (s msgServer) ConfirmGatewayTx(c context.Context, req *types.ConfirmGatewayTxRequest) (*types.ConfirmGatewayTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	keeper := s.ForChain(chain.Name)
	gatewayAddress, ok := keeper.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	pollParticipants, err := s.initializePoll(ctx, chain, req.TxID)
	if err != nil {
		return nil, err
	}

	height, ok := keeper.GetRequiredConfirmationHeight(ctx)
	if !ok {
		return nil, fmt.Errorf("required confirmation height not found")
	}

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.ConfirmGatewayTxStarted{
		TxID:               req.TxID,
		Chain:              chain.Name,
		GatewayAddress:     gatewayAddress,
		ConfirmationHeight: height,
		PollParticipants:   pollParticipants,
	}))

	return &types.ConfirmGatewayTxResponse{}, nil
}

func (s msgServer) SetGateway(c context.Context, req *types.SetGatewayRequest) (*types.SetGatewayResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	if _, ok := s.multisigKeeper.GetCurrentKeyID(ctx, chain.Name); !ok {
		return nil, fmt.Errorf("current key not set for chain %s", chain.Name)
	}

	keeper := s.ForChain(chain.Name)
	if _, ok := keeper.GetGatewayAddress(ctx); ok {
		return nil, fmt.Errorf("%s gateway already set", req.Chain)
	}

	keeper.SetGateway(ctx, req.Address)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeGateway,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address.Hex()),
		),
	)

	return &types.SetGatewayResponse{}, nil
}

func (s msgServer) Link(c context.Context, req *types.LinkRequest) (*types.LinkResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	senderChain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, senderChain); err != nil {
		return nil, err
	}

	keeper := s.ForChain(senderChain.Name)
	gatewayAddr, ok := keeper.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	recipientChain, ok := s.nexus.GetChain(ctx, req.RecipientChain)
	if !ok {
		return nil, fmt.Errorf("unknown recipient chain")
	}

	token := keeper.GetERC20TokenByAsset(ctx, req.Asset)
	found := s.nexus.IsAssetRegistered(ctx, recipientChain, req.Asset)
	if !found || !token.Is(types.Confirmed) {
		return nil, fmt.Errorf("asset '%s' not registered for chain '%s'", req.Asset, recipientChain.Name)
	}

	salt := keeper.GenerateSalt(ctx, req.RecipientAddr)
	burnerAddress, err := keeper.GetBurnerAddress(ctx, token, salt, gatewayAddr)
	if err != nil {
		return nil, err
	}

	symbol := token.GetDetails().Symbol
	recipient := nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr}

	err = s.nexus.LinkAddresses(ctx,
		nexus.CrossChainAddress{Chain: senderChain, Address: burnerAddress.Hex()},
		recipient)
	if err != nil {
		return nil, fmt.Errorf("could not link addresses: %s", err.Error())
	}

	burnerInfo := types.BurnerInfo{
		BurnerAddress:    burnerAddress,
		TokenAddress:     token.GetAddress(),
		DestinationChain: req.RecipientChain,
		Symbol:           symbol,
		Asset:            req.Asset,
		Salt:             salt,
	}
	keeper.SetBurnerInfo(ctx, burnerInfo)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLink,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeySourceChain, senderChain.Name.String()),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, burnerAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, req.RecipientAddr),
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipientChain.Name.String()),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, token.GetAddress().Hex()),
			sdk.NewAttribute(types.AttributeKeyAsset, req.Asset),
		),
	)

	s.Logger(ctx).Debug(fmt.Sprintf("successfully linked deposit %s on chain %s to recipient %s on chain %s for asset %s", burnerAddress.Hex(), req.Chain, req.RecipientAddr, req.RecipientChain, req.Asset),
		types.AttributeKeySourceChain, senderChain.Name,
		types.AttributeKeyDepositAddress, burnerAddress.Hex(),
		types.AttributeKeyDestinationChain, recipientChain.Name,
		types.AttributeKeyDestinationAddress, req.RecipientAddr,
		types.AttributeKeyAsset, req.Asset,
	)

	return &types.LinkResponse{DepositAddr: burnerAddress.Hex()}, nil
}

// ConfirmToken handles token deployment confirmation
func (s msgServer) ConfirmToken(c context.Context, req *types.ConfirmTokenRequest) (*types.ConfirmTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	_, ok = s.nexus.GetChain(ctx, req.Asset.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Asset.Chain)
	}

	keeper := s.ForChain(chain.Name)
	token := keeper.GetERC20TokenByAsset(ctx, req.Asset.Name)

	err := token.RecordDeployment(req.TxID)
	if err != nil {
		return nil, err
	}

	pollParticipants, err := s.initializePoll(ctx, chain, req.TxID)
	if err != nil {
		return nil, err
	}

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.ConfirmTokenStarted{
		TxID:               req.TxID,
		Chain:              chain.Name,
		GatewayAddress:     funcs.MustOk(keeper.GetGatewayAddress(ctx)),
		TokenAddress:       token.GetAddress(),
		TokenDetails:       token.GetDetails(),
		ConfirmationHeight: funcs.MustOk(keeper.GetRequiredConfirmationHeight(ctx)),
		PollParticipants:   pollParticipants,
	}))

	return &types.ConfirmTokenResponse{}, nil
}

// ConfirmDeposit handles deposit confirmations
func (s msgServer) ConfirmDeposit(c context.Context, req *types.ConfirmDepositRequest) (*types.ConfirmDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	keeper := s.ForChain(chain.Name)
	gatewayAddr, ok := keeper.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("gateway address not set for chain %s", chain.Name)
	}

	burnerInfo := keeper.GetBurnerInfo(ctx, req.BurnerAddress)
	if burnerInfo == nil {
		return nil, fmt.Errorf("no burner info found for address %s", req.BurnerAddress.Hex())
	}

	token := keeper.GetERC20TokenByAsset(ctx, burnerInfo.Asset)
	if !token.Is(types.Confirmed) {
		return nil, fmt.Errorf("token %s is not confirmed on %s", token.GetAsset(), chain.Name)
	}

	burnerAddress, err := keeper.GetBurnerAddress(ctx, token, burnerInfo.Salt, gatewayAddr)
	if err != nil {
		return nil, err
	}

	if burnerAddress != req.BurnerAddress {
		return nil, fmt.Errorf("provided burner address %s doesn't match expected address %s", req.BurnerAddress, burnerAddress)
	}

	pollParticipants, err := s.initializePoll(ctx, chain, req.TxID)
	if err != nil {
		return nil, err
	}

	height, _ := keeper.GetRequiredConfirmationHeight(ctx)
	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.ConfirmDepositStarted{
		TxID:               req.TxID,
		Chain:              chain.Name,
		DepositAddress:     req.BurnerAddress,
		TokenAddress:       burnerInfo.TokenAddress,
		ConfirmationHeight: height,
		PollParticipants:   pollParticipants,
	}))

	return &types.ConfirmDepositResponse{}, nil
}

// ConfirmTransferKey handles transfer operatorship confirmation
func (s msgServer) ConfirmTransferKey(c context.Context, req *types.ConfirmTransferKeyRequest) (*types.ConfirmTransferKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	if _, ok := s.multisigKeeper.GetNextKeyID(ctx, chain.Name); !ok {
		return nil, fmt.Errorf("next key for chain %s not set yet", chain.Name)
	}

	keeper := s.ForChain(chain.Name)

	gatewayAddr, ok := keeper.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	pollParticipants, err := s.initializePoll(ctx, chain, req.TxID)
	if err != nil {
		return nil, err
	}

	params := keeper.GetParams(ctx)
	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewConfirmKeyTransferStarted(chain.Name, req.TxID, gatewayAddr, params.ConfirmationHeight, pollParticipants)))

	return &types.ConfirmTransferKeyResponse{}, nil
}

func (s msgServer) CreateDeployToken(c context.Context, req *types.CreateDeployTokenRequest) (*types.CreateDeployTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	dailyMintLimit, err := sdk.ParseUint(req.DailyMintLimit)
	if err != nil {
		return nil, err
	}

	keeper := s.ForChain(chain.Name)

	switch req.Address.IsZeroAddress() {
	case true:
		originChain, found := s.nexus.GetChain(ctx, req.Asset.Chain)
		if !found {
			return nil, fmt.Errorf("%s is not a registered chain", req.Asset.Chain)
		}

		if !s.nexus.IsAssetRegistered(ctx, originChain, req.Asset.Name) {
			return nil, fmt.Errorf("asset %s is not registered on the origin chain %s", req.Asset.Name, originChain.Name)
		}
	case false:
		for _, c := range s.nexus.GetChains(ctx) {
			if s.nexus.IsAssetRegistered(ctx, c, req.Asset.Name) {
				return nil, fmt.Errorf("asset %s already registered on chain %s", req.Asset.Name, c.Name)
			}
		}

		for _, token := range keeper.GetTokens(ctx) {
			if bytes.Equal(token.GetAddress().Bytes(), req.Address.Bytes()) {
				return nil, fmt.Errorf("token %s already created for chain %s", token.GetAddress().Hex(), chain.Name)
			}
		}
	}

	keyID, ok := s.multisigKeeper.GetCurrentKeyID(ctx, chain.Name)
	if !ok {
		return nil, fmt.Errorf("current key not set for chain %s", chain.Name)
	}

	token, err := keeper.CreateERC20Token(ctx, req.Asset.Name, req.TokenDetails, req.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to initialize token %s(%s) for chain %s", req.TokenDetails.TokenName, req.TokenDetails.Symbol, chain.Name)
	}

	cmd, err := token.CreateDeployCommand(keyID, dailyMintLimit)
	if err != nil {
		return nil, err
	}

	if err := keeper.EnqueueCommand(ctx, cmd); err != nil {
		return nil, err
	}

	if err = s.nexus.RegisterAsset(ctx, chain, nexus.NewAsset(req.Asset.Name, false)); err != nil {
		return nil, err
	}

	return &types.CreateDeployTokenResponse{}, nil
}

func (s msgServer) CreateBurnTokens(c context.Context, req *types.CreateBurnTokensRequest) (*types.CreateBurnTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	keeper := s.ForChain(chain.Name)

	deposits := keeper.GetConfirmedDeposits(ctx)
	if len(deposits) == 0 {
		return &types.CreateBurnTokensResponse{}, nil
	}

	chainID, ok := keeper.GetChainID(ctx)
	if !ok {
		return nil, fmt.Errorf("could not find chain ID for '%s'", chain.Name)
	}

	keyID, ok := s.multisigKeeper.GetCurrentKeyID(ctx, chain.Name)
	if !ok {
		return nil, fmt.Errorf("current key not set for chain %s", chain.Name)
	}

	seen := map[string]bool{}
	for _, deposit := range deposits {
		keeper.DeleteDeposit(ctx, deposit)
		keeper.SetDeposit(ctx, deposit, types.DepositStatus_Burned)

		burnerAddressHex := deposit.BurnerAddress.Hex()

		if seen[burnerAddressHex] {
			continue
		}

		burnerInfo := keeper.GetBurnerInfo(ctx, deposit.BurnerAddress)
		if burnerInfo == nil {
			return nil, fmt.Errorf("no burner info found for address %s", burnerAddressHex)
		}

		token := keeper.GetERC20TokenByAsset(ctx, burnerInfo.Asset)
		if !token.Is(types.Confirmed) {
			return nil, fmt.Errorf("token %s is not confirmed on %s", token.GetAsset(), chain.Name)
		}

		cmd, err := types.CreateBurnTokenCommand(chainID, multisig.KeyID(keyID), ctx.BlockHeight(), *burnerInfo, token.IsExternal())
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to create burn-token command to burn token at address %s for chain %s", burnerAddressHex, chain.Name)
		}

		if err := keeper.EnqueueCommand(ctx, cmd); err != nil {
			return nil, err
		}

		seen[burnerAddressHex] = true
	}

	return &types.CreateBurnTokensResponse{}, nil
}

func (s msgServer) CreatePendingTransfers(c context.Context, req *types.CreatePendingTransfersRequest) (*types.CreatePendingTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	keeper := s.ForChain(chain.Name)

	pendingTransfers := s.nexus.GetTransfersForChain(ctx, chain, nexus.Pending)
	if len(pendingTransfers) == 0 {
		s.Logger(ctx).Debug("no pending transfers found")
		return &types.CreatePendingTransfersResponse{}, nil
	}

	keyID, ok := s.multisigKeeper.GetCurrentKeyID(ctx, chain.Name)
	if !ok {
		return nil, fmt.Errorf("current key not set for chain %s", chain.Name)
	}

	for _, transfer := range pendingTransfers {
		token := keeper.GetERC20TokenByAsset(ctx, transfer.Asset.Denom)
		if !token.Is(types.Confirmed) {
			s.Logger(ctx).Debug(fmt.Sprintf("token %s is not confirmed on %s", token.GetAsset(), chain.Name))
			continue
		}

		cmd, err := token.CreateMintCommand(multisig.KeyID(keyID), transfer)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed create mint-token command for transfer %d", transfer.ID)
		}

		s.Logger(ctx).Info(fmt.Sprintf("minting %s to recipient %s on %s with transfer ID %s and command ID %s", transfer.Asset.String(), transfer.Recipient.Address, transfer.Recipient.Chain.Name, transfer.ID.String(), cmd.ID.Hex()),
			types.AttributeKeyDestinationChain, transfer.Recipient.Chain.Name,
			types.AttributeKeyDestinationAddress, transfer.Recipient.Address,
			sdk.AttributeKeyAmount, transfer.Asset.String(),
			types.AttributeKeyAsset, transfer.Asset.Denom,
			types.AttributeKeyTransferID, transfer.ID.String(),
			types.AttributeKeyCommandsID, cmd.ID.Hex(),
		)

		if err := keeper.EnqueueCommand(ctx, cmd); err != nil {
			return nil, err
		}

		s.nexus.ArchivePendingTransfer(ctx, transfer)
	}

	return &types.CreatePendingTransfersResponse{}, nil
}

func (s msgServer) CreateTransferOperatorship(c context.Context, req *types.CreateTransferOperatorshipRequest) (*types.CreateTransferOperatorshipResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keeper := s.ForChain(req.Chain)

	if _, ok := keeper.GetGatewayAddress(ctx); !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	cmd, err := s.createTransferKeyCommand(ctx, keeper, req.Chain, multisig.KeyID(req.KeyID))
	if err != nil {
		return nil, err
	}

	if err := keeper.EnqueueCommand(ctx, cmd); err != nil {
		return nil, err
	}

	return &types.CreateTransferOperatorshipResponse{}, nil
}

func (s msgServer) createTransferKeyCommand(ctx sdk.Context, keeper types.ChainKeeper, chainStr nexus.ChainName, nextKeyID multisig.KeyID) (types.Command, error) {
	chain, ok := s.nexus.GetChain(ctx, chainStr)
	if !ok {
		return types.Command{}, fmt.Errorf("%s is not a registered chain", chainStr)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return types.Command{}, err
	}

	chainID, ok := keeper.GetChainID(ctx)
	if !ok {
		return types.Command{}, fmt.Errorf("could not find chain ID for '%s'", chainStr)
	}

	if _, ok := s.multisigKeeper.GetNextKeyID(ctx, chain.Name); ok {
		return types.Command{}, sdkerrors.Wrapf(types.ErrRotationInProgress, "finish rotating to next key for chain %s first", chain.Name)
	}

	if err := s.multisigKeeper.AssignKey(ctx, chain.Name, nextKeyID); err != nil {
		return types.Command{}, err
	}

	keyID, ok := s.multisigKeeper.GetCurrentKeyID(ctx, chain.Name)
	if !ok {
		return types.Command{}, fmt.Errorf("current key not set for chain %s", chain.Name)
	}

	nextKey, ok := s.multisigKeeper.GetKey(ctx, nextKeyID)
	if !ok {
		return types.Command{}, fmt.Errorf("could not find threshold key '%s'", nextKeyID)
	}

	return types.CreateMultisigTransferCommand(chainID, keyID, nextKey), nil
}

func getCommandBatchToSign(ctx sdk.Context, keeper types.ChainKeeper) (types.CommandBatch, error) {
	latest := keeper.GetLatestCommandBatch(ctx)

	switch latest.GetStatus() {
	case types.BatchSigning:
		return types.CommandBatch{}, sdkerrors.Wrapf(types.ErrSignCommandsInProgress, "command batch '%s'", hex.EncodeToString(latest.GetID()))
	case types.BatchAborted:
		return latest, nil
	default:
		return keeper.CreateNewBatchToSign(ctx)
	}
}

func (s msgServer) SignCommands(c context.Context, req *types.SignCommandsRequest) (*types.SignCommandsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	keeper := s.ForChain(chain.Name)

	if _, ok := keeper.GetChainID(ctx); !ok {
		return nil, fmt.Errorf("could not find chain ID for '%s'", chain.Name)
	}

	commandBatch, err := getCommandBatchToSign(ctx, keeper)
	if err != nil {
		return nil, err
	}
	if len(commandBatch.GetCommandIDs()) == 0 {
		return &types.SignCommandsResponse{CommandCount: 0, BatchedCommandsID: nil}, nil
	}

	if err := s.multisigKeeper.Sign(
		ctx,
		commandBatch.GetKeyID(),
		commandBatch.GetSigHash().Bytes(),
		types.ModuleName,
		types.NewSigMetadata(types.SigCommand, chain.Name, commandBatch.GetID()),
	); err != nil {
		return nil, err
	}

	if !commandBatch.SetStatus(types.BatchSigning) {
		return nil, fmt.Errorf("failed setting status of command batch %s to be signing", hex.EncodeToString(commandBatch.GetID()))
	}

	batchedCommandsIDHex := hex.EncodeToString(commandBatch.GetID())
	commandList := types.CommandIDsToStrings(commandBatch.GetCommandIDs())
	for _, commandID := range commandList {
		s.Logger(ctx).Info(
			fmt.Sprintf("signing command %s in batch %s for chain %s using key %s", commandID, batchedCommandsIDHex, chain.Name, string(commandBatch.GetKeyID())),
			types.AttributeKeyChain, chain.Name,
			types.AttributeKeyKeyID, string(commandBatch.GetKeyID()),
			"commandBatchID", batchedCommandsIDHex,
			"commandID", commandID,
		)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSign,
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyBatchedCommandsID, batchedCommandsIDHex),
			sdk.NewAttribute(types.AttributeKeyCommandsIDs, strings.Join(commandList, ",")),
		),
	)

	return &types.SignCommandsResponse{CommandCount: uint32(len(commandBatch.GetCommandIDs())), BatchedCommandsID: commandBatch.GetID()}, nil
}

func (s msgServer) AddChain(c context.Context, req *types.AddChainRequest) (*types.AddChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, found := s.nexus.GetChain(ctx, req.Name); found {
		return nil, fmt.Errorf("chain '%s' is already registered", req.Name)
	}

	if err := req.Params.Validate(); err != nil {
		return nil, err
	}

	chain := nexus.Chain{Name: req.Name, SupportsForeignAssets: true, KeyType: tss.Multisig, Module: types.ModuleName}
	s.nexus.SetChain(ctx, chain)
	s.ForChain(chain.Name).SetParams(ctx, req.Params)

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.ChainAdded{Chain: req.Name}))

	return &types.AddChainResponse{}, nil
}

func (s msgServer) RetryFailedEvent(c context.Context, req *types.RetryFailedEventRequest) (*types.RetryFailedEventResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	keeper := s.ForChain(chain.Name)

	event, ok := keeper.GetEvent(ctx, req.EventID)
	if !ok {
		return nil, fmt.Errorf("event %s not found for chain %s", req.EventID, req.Chain)
	}

	if event.Status != types.EventFailed {
		return nil, fmt.Errorf("event %s is not a failed event", req.EventID)
	}

	event.Status = types.EventConfirmed
	keeper.GetConfirmedEventQueue(ctx).Enqueue(getEventKey(req.EventID), &event)

	s.Logger(ctx).Info(
		"re-queued failed event",
		types.AttributeKeyChain, chain.Name,
		"eventID", req.EventID,
	)

	return &types.RetryFailedEventResponse{}, nil
}

func (s msgServer) initializePoll(ctx sdk.Context, chain nexus.Chain, txID types.Hash) (vote.PollParticipants, error) {
	keeper := s.ForChain(chain.Name)
	params := keeper.GetParams(ctx)
	snap, err := s.snapshotter.CreateSnapshot(
		ctx,
		s.nexus.GetChainMaintainers(ctx, chain),
		excludeJailedOrTombstoned(ctx, s.slashing, s.snapshotter),
		snapshot.QuadraticWeightFunc,
		params.VotingThreshold,
	)
	if err != nil {
		return vote.PollParticipants{}, err
	}

	pollID, err := s.voter.InitializePoll(
		ctx,
		vote.NewPollBuilder(types.ModuleName, params.VotingThreshold, snap, ctx.BlockHeight()+params.RevoteLockingPeriod).
			MinVoterCount(params.MinVoterCount).
			RewardPoolName(chain.Name.String()).
			GracePeriod(keeper.GetParams(ctx).VotingGracePeriod).
			ModuleMetadata(&types.PollMetadata{
				Chain: chain.Name,
				TxID:  txID,
			}),
	)
	return vote.PollParticipants{
		PollID:       pollID,
		Participants: snap.GetParticipantAddresses(),
	}, err
}
