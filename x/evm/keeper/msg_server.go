package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.BaseKeeper
	tss         types.TSS
	signer      types.Signer
	nexus       types.Nexus
	voter       types.Voter
	snapshotter types.Snapshotter
}

// NewMsgServerImpl returns an implementation of the bitcoin MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper types.BaseKeeper, t types.TSS, n types.Nexus, s types.Signer, v types.Voter, snap types.Snapshotter) types.MsgServiceServer {
	return msgServer{
		BaseKeeper:  keeper,
		tss:         t,
		signer:      s,
		nexus:       n,
		voter:       v,
		snapshotter: snap,
	}
}

func validateChainActivated(ctx sdk.Context, n types.Nexus, chain nexus.Chain) error {
	if !n.IsChainActivated(ctx, chain) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest,
			fmt.Sprintf("chain %s is not activated yet", chain.Name))
	}

	return nil
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

	period, ok := keeper.GetRevoteLockingPeriod(ctx)
	if !ok {
		return nil, fmt.Errorf("could not retrieve revote locking period")
	}

	votingThreshold, ok := keeper.GetVotingThreshold(ctx)
	if !ok {
		return nil, fmt.Errorf("voting threshold not found")
	}

	minVoterCount, ok := keeper.GetMinVoterCount(ctx)
	if !ok {
		return nil, fmt.Errorf("min voter count not found")
	}

	pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s", req.Chain, req.TxID.Hex()))
	if err := s.voter.InitializePoll(
		ctx,
		pollKey,
		s.nexus.GetChainMaintainers(ctx, chain),
		vote.ExpiryAt(ctx.BlockHeight()+period),
		vote.Threshold(votingThreshold),
		vote.MinVoterCount(minVoterCount),
		vote.RewardPool(chain.Name),
		vote.GracePeriod(keeper.GetParams(ctx).VotingGracePeriod),
	); err != nil {
		return nil, err
	}

	height, ok := keeper.GetRequiredConfirmationHeight(ctx)
	if !ok {
		return nil, fmt.Errorf("required confirmation height not found")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeGatewayTxConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyGatewayAddress, gatewayAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
		),
	)

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

	if _, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey); !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	if _, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.SecondaryKey); !ok {
		return nil, fmt.Errorf("no secondary key for chain %s found", chain.Name)
	}

	if _, ok := s.signer.GetExternalKeyIDs(ctx, chain); !ok {
		return nil, fmt.Errorf("no external keys for chain %s found", chain.Name)
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
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
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

	burnerAddress, salt, err := keeper.GetBurnerAddressAndSalt(ctx, token, req.RecipientAddr, gatewayAddr)
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
			sdk.NewAttribute(types.AttributeKeySourceChain, senderChain.Name),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, burnerAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, req.RecipientAddr),
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipientChain.Name),
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

	period, ok := keeper.GetRevoteLockingPeriod(ctx)
	if !ok {
		return nil, fmt.Errorf("could not retrieve revote locking period")
	}

	votingThreshold, ok := keeper.GetVotingThreshold(ctx)
	if !ok {
		return nil, fmt.Errorf("voting threshold not found")
	}

	minVoterCount, ok := keeper.GetMinVoterCount(ctx)
	if !ok {
		return nil, fmt.Errorf("min voter count not found")
	}

	pollKey := types.GetConfirmTokenKey(req.TxID, req.Asset.Name)
	if err := s.voter.InitializePoll(
		ctx,
		pollKey,
		s.nexus.GetChainMaintainers(ctx, chain),
		vote.ExpiryAt(ctx.BlockHeight()+period),
		vote.Threshold(votingThreshold),
		vote.MinVoterCount(minVoterCount),
		vote.RewardPool(chain.Name),
		vote.GracePeriod(keeper.GetParams(ctx).VotingGracePeriod),
	); err != nil {
		return nil, err
	}

	// if token was initialized, both token and gateway addresses are available
	tokenAddr := token.GetAddress()
	gatewayAddr, _ := keeper.GetGatewayAddress(ctx)
	height, _ := keeper.GetRequiredConfirmationHeight(ctx)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTokenConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyGatewayAddress, gatewayAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, tokenAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeySymbol, token.GetDetails().Symbol),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
		),
	)

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

	burnerInfo := keeper.GetBurnerInfo(ctx, req.BurnerAddress)
	if burnerInfo == nil {
		return nil, fmt.Errorf("no burner info found for address %s", req.BurnerAddress.Hex())
	}

	period, ok := keeper.GetRevoteLockingPeriod(ctx)
	if !ok {
		return nil, fmt.Errorf("could not retrieve revote locking period for chain %s", chain.Name)
	}

	votingThreshold, ok := keeper.GetVotingThreshold(ctx)
	if !ok {
		return nil, fmt.Errorf("voting threshold for chain %s not found", chain.Name)
	}

	minVoterCount, ok := keeper.GetMinVoterCount(ctx)
	if !ok {
		return nil, fmt.Errorf("min voter count for chain %s not found", chain.Name)
	}

	pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s", req.TxID.Hex(), req.BurnerAddress.Hex()))
	if err := s.voter.InitializePoll(
		ctx,
		pollKey,
		s.nexus.GetChainMaintainers(ctx, chain),
		vote.ExpiryAt(ctx.BlockHeight()+period),
		vote.Threshold(votingThreshold),
		vote.MinVoterCount(minVoterCount),
		vote.RewardPool(chain.Name),
		vote.GracePeriod(keeper.GetParams(ctx).VotingGracePeriod),
	); err != nil {
		return nil, err
	}

	height, _ := keeper.GetRequiredConfirmationHeight(ctx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, req.BurnerAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, burnerInfo.TokenAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
		),
	)

	return &types.ConfirmDepositResponse{}, nil
}

// ConfirmTransferKey handles transfer ownership/operatorship confirmations
func (s msgServer) ConfirmTransferKey(c context.Context, req *types.ConfirmTransferKeyRequest) (*types.ConfirmTransferKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	var keyRole tss.KeyRole
	switch req.TransferType {
	case types.Ownership:
		keyRole = tss.MasterKey
	case types.Operatorship:
		keyRole = tss.SecondaryKey
	default:
		return nil, fmt.Errorf("invalid transfer type %s", req.TransferType.SimpleString())
	}

	_, ok = s.signer.GetNextKeyID(ctx, chain, keyRole)
	if !ok {
		return nil, fmt.Errorf("next %s key for chain %s not set yet", keyRole.SimpleString(), chain.Name)
	}

	keeper := s.ForChain(chain.Name)

	gatewayAddr, ok := keeper.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	period, ok := keeper.GetRevoteLockingPeriod(ctx)
	if !ok {
		return nil, fmt.Errorf("could not retrieve revote locking period for chain %s", chain.Name)
	}

	votingThreshold, ok := keeper.GetVotingThreshold(ctx)
	if !ok {
		return nil, fmt.Errorf("voting threshold for chain %s not found", chain.Name)
	}

	minVoterCount, ok := keeper.GetMinVoterCount(ctx)
	if !ok {
		return nil, fmt.Errorf("min voter count for chain %s not found", chain.Name)
	}

	pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s_%s", req.TxID.Hex(), req.TransferType.SimpleString(), req.KeyID))
	if err := s.voter.InitializePoll(
		ctx,
		pollKey,
		s.nexus.GetChainMaintainers(ctx, chain),
		vote.ExpiryAt(ctx.BlockHeight()+period),
		vote.Threshold(votingThreshold),
		vote.MinVoterCount(minVoterCount),
		vote.RewardPool(chain.Name),
		vote.GracePeriod(keeper.GetParams(ctx).VotingGracePeriod),
	); err != nil {
		return nil, err
	}

	height, _ := keeper.GetRequiredConfirmationHeight(ctx)

	event := sdk.NewEvent(types.EventTypeTransferKeyConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
		sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
		sdk.NewAttribute(types.AttributeKeyTransferKeyType, req.TransferType.SimpleString()),
		sdk.NewAttribute(types.AttributeKeyKeyType, chain.KeyType.SimpleString()),
		sdk.NewAttribute(types.AttributeKeyGatewayAddress, gatewayAddr.Hex()),
		sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
	)
	defer func() { ctx.EventManager().EmitEvent(event) }()

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

	if _, nextMasterKeyAssigned := s.signer.GetNextKeyID(ctx, chain, tss.MasterKey); nextMasterKeyAssigned {
		return nil, fmt.Errorf("next %s key already assigned for chain %s, rotate key first", tss.MasterKey.SimpleString(), chain.Name)
	}

	masterKeyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	token, err := keeper.CreateERC20Token(ctx, req.Asset.Name, req.TokenDetails, req.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to initialize token %s(%s) for chain %s", req.TokenDetails.TokenName, req.TokenDetails.Symbol, chain.Name)
	}

	cmd, err := token.CreateDeployCommand(masterKeyID)
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

	if _, nextSecondaryKeyAssigned := s.signer.GetNextKeyID(ctx, chain, tss.SecondaryKey); nextSecondaryKeyAssigned {
		return nil, s.newErrRotationInProgress(chain, tss.SecondaryKey)
	}

	secondaryKeyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("no %s key for chain %s found", tss.SecondaryKey.SimpleString(), chain.Name)
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

		cmd, err := types.CreateBurnTokenCommand(chainID, secondaryKeyID, ctx.BlockHeight(), *burnerInfo, token.IsExternal())
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

func (s msgServer) newErrRotationInProgress(chain nexus.Chain, key tss.KeyRole) error {
	return sdkerrors.Wrapf(types.ErrRotationInProgress, "finish rotating to next %s key for chain %s first", key.SimpleString(), chain.Name)
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

	if _, nextSecondaryKeyAssigned := s.signer.GetNextKeyID(ctx, chain, tss.SecondaryKey); nextSecondaryKeyAssigned {
		return nil, s.newErrRotationInProgress(chain, tss.SecondaryKey)
	}

	secondaryKeyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("no %s key for chain %s found", tss.SecondaryKey.SimpleString(), chain.Name)
	}

	for _, transfer := range pendingTransfers {
		token := keeper.GetERC20TokenByAsset(ctx, transfer.Asset.Denom)
		if !token.Is(types.Confirmed) {
			s.Logger(ctx).Debug(fmt.Sprintf("token %s is not confirmed on %s", token.GetAsset(), chain.Name))
			continue
		}

		cmd, err := token.CreateMintCommand(secondaryKeyID, transfer)
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

func (s msgServer) createTransferKeyCommand(ctx sdk.Context, keeper types.ChainKeeper, transferKeyType types.TransferKeyType, chainStr string, nextKeyID tss.KeyID) (types.Command, error) {
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

	var keyRole tss.KeyRole
	switch transferKeyType {
	case types.Ownership:
		keyRole = tss.MasterKey
	case types.Operatorship:
		keyRole = tss.SecondaryKey
	default:
		return types.Command{}, fmt.Errorf("invalid transfer key type %s", transferKeyType.SimpleString())
	}

	// don't allow any transfer key if the next master/secondary key is already assigned
	if _, nextMasterKeyAssigned := s.signer.GetNextKeyID(ctx, chain, tss.MasterKey); nextMasterKeyAssigned {
		return types.Command{}, s.newErrRotationInProgress(chain, tss.MasterKey)
	}
	if _, nextSecondaryKeyAssigned := s.signer.GetNextKeyID(ctx, chain, tss.SecondaryKey); nextSecondaryKeyAssigned {
		return types.Command{}, s.newErrRotationInProgress(chain, tss.SecondaryKey)
	}

	if err := s.signer.AssertMatchesRequirements(ctx, s.snapshotter, chain, nextKeyID, keyRole); err != nil {
		return types.Command{}, sdkerrors.Wrapf(err, "key %s does not match requirements for role %s", nextKeyID, keyRole.SimpleString())
	}

	if err := s.signer.AssignNextKey(ctx, chain, keyRole, nextKeyID); err != nil {
		return types.Command{}, err
	}

	currMasterKeyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return types.Command{}, fmt.Errorf("current %s key not set for chain %s", tss.MasterKey, chain.Name)
	}

	nextKey, ok := s.signer.GetKey(ctx, nextKeyID)
	if !ok {
		return types.Command{}, fmt.Errorf("could not find threshold key '%s'", nextKeyID)
	}

	switch chain.KeyType {
	case tss.Threshold:
		pk, err := nextKey.GetECDSAPubKey()
		if err != nil {
			return types.Command{}, err
		}

		address := crypto.PubkeyToAddress(pk)
		s.Logger(ctx).Debug(fmt.Sprintf("creating command %s for chain %s to transfer to address %s", transferKeyType.SimpleString(), chain.Name, address))

		return types.CreateSinglesigTransferCommand(transferKeyType, chainID, currMasterKeyID, crypto.PubkeyToAddress(pk))
	case tss.Multisig:
		addresses, threshold, err := types.GetMultisigAddresses(nextKey)
		if err != nil {
			return types.Command{}, err
		}

		addressStrs := make([]string, len(addresses))
		for i, address := range addresses {
			addressStrs[i] = address.Hex()
		}

		s.Logger(ctx).Debug(fmt.Sprintf("creating command %s for chain %s to transfer to addresses %s", transferKeyType.SimpleString(), chain.Name, strings.Join(addressStrs, ",")))

		return types.CreateMultisigTransferCommand(transferKeyType, chainID, currMasterKeyID, threshold, addresses...)
	default:
		return types.Command{}, fmt.Errorf("invalid key type '%s'", chain.KeyType.SimpleString())
	}
}

func (s msgServer) CreateTransferOwnership(c context.Context, req *types.CreateTransferOwnershipRequest) (*types.CreateTransferOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keeper := s.ForChain(req.Chain)

	if _, ok := keeper.GetGatewayAddress(ctx); !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	cmd, err := s.createTransferKeyCommand(ctx, keeper, types.Ownership, req.Chain, req.KeyID)
	if err != nil {
		return nil, err
	}

	if err := keeper.EnqueueCommand(ctx, cmd); err != nil {
		return nil, err
	}

	return &types.CreateTransferOwnershipResponse{}, nil
}

func (s msgServer) CreateTransferOperatorship(c context.Context, req *types.CreateTransferOperatorshipRequest) (*types.CreateTransferOperatorshipResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keeper := s.ForChain(req.Chain)

	if _, ok := keeper.GetGatewayAddress(ctx); !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	cmd, err := s.createTransferKeyCommand(ctx, keeper, types.Operatorship, req.Chain, req.KeyID)
	if err != nil {
		return nil, err
	}

	if err := keeper.EnqueueCommand(ctx, cmd); err != nil {
		return nil, err
	}

	return &types.CreateTransferOperatorshipResponse{}, nil
}

func getCommandBatchToSign(ctx sdk.Context, keeper types.ChainKeeper, signer types.Signer) (types.CommandBatch, error) {
	latest := keeper.GetLatestCommandBatch(ctx)

	switch latest.GetStatus() {
	case types.BatchSigning:
		return types.CommandBatch{}, sdkerrors.Wrapf(types.ErrSignCommandsInProgress, "command batch '%s'", hex.EncodeToString(latest.GetID()))
	case types.BatchAborted:
		return latest, nil
	default:
		return keeper.CreateNewBatchToSign(ctx, signer)
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

	commandBatch, err := getCommandBatchToSign(ctx, keeper, s.signer)
	if err != nil {
		return nil, err
	}
	if len(commandBatch.GetCommandIDs()) == 0 {
		return &types.SignCommandsResponse{CommandCount: 0, BatchedCommandsID: nil}, nil
	}

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, commandBatch.GetKeyID())
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", commandBatch.GetKeyID())
	}

	sigMetadata, err := codectypes.NewAnyWithValue(&types.SigMetadata{
		Type:  types.SigCommand,
		Chain: chain.Name,
	})
	if err != nil {
		return nil, err
	}

	batchedCommandsIDHex := hex.EncodeToString(commandBatch.GetID())
	err = s.signer.StartSign(ctx, tss.SignInfo{
		KeyID:           commandBatch.GetKeyID(),
		SigID:           batchedCommandsIDHex,
		Msg:             commandBatch.GetSigHash().Bytes(),
		SnapshotCounter: counter,
		RequestModule:   types.ModuleName,
		ModuleMetadata:  sigMetadata,
	}, s.snapshotter, s.voter)
	if err != nil {
		return nil, err
	}

	if !commandBatch.SetStatus(types.BatchSigning) {
		return nil, fmt.Errorf("failed setting status of command batch %s to be signing", hex.EncodeToString(commandBatch.GetID()))
	}

	commandList := types.CommandIDsToStrings(commandBatch.GetCommandIDs())
	for _, commandID := range commandList {
		s.Logger(ctx).Info(
			fmt.Sprintf("signing command %s in batch %s for chain %s using key %s", commandID, batchedCommandsIDHex, chain.Name, string(commandBatch.GetKeyID())),
			types.AttributeKeyChain, chain.Name,
			tsstypes.AttributeKeyKeyID, string(commandBatch.GetKeyID()),
			"commandBatchID", batchedCommandsIDHex,
			"commandID", commandID,
		)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSign,
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
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

	if !tsstypes.TSSEnabled && req.KeyType == tss.Threshold {
		return nil, fmt.Errorf("TSS is disabled")
	}

	chain := nexus.Chain{Name: req.Name, SupportsForeignAssets: true, KeyType: req.KeyType, Module: types.ModuleName}
	s.nexus.SetChain(ctx, chain)
	s.ForChain(chain.Name).SetParams(ctx, req.Params)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeNewChain,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueUpdate),
			sdk.NewAttribute(types.AttributeKeyChain, req.Name),
		),
	)

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
