package keeper

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/utils"
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

func (s msgServer) ConfirmGatewayDeployment(c context.Context, req *types.ConfirmGatewayDeploymentRequest) (*types.ConfirmGatewayDeploymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	keeper := s.ForChain(chain.Name)

	if _, ok := keeper.GetPendingGatewayAddress(ctx); ok {
		return nil, fmt.Errorf("gateway is in the process of confirmation")
	}

	if _, ok := keeper.GetGatewayAddress(ctx); ok {
		return nil, fmt.Errorf("gateway is already confirmed")
	}

	keeper.SetPendingGateway(ctx, common.Address(req.Address))

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

	pollKey := types.GetConfirmGatewayDeploymentPollKey(chain, req.TxID, req.Address)
	if err := s.voter.InitializePoll(
		ctx,
		pollKey,
		s.nexus.GetChainMaintainers(ctx, chain),
		vote.ExpiryAt(ctx.BlockHeight()+period),
		vote.Threshold(votingThreshold),
		vote.MinVoterCount(minVoterCount),
		vote.RewardPool(chain.Name),
	); err != nil {
		return nil, err
	}

	deploymentBytecode, err := getGatewayDeploymentBytecode(ctx, keeper, s.signer, chain)
	if err != nil {
		return nil, err
	}

	height, _ := keeper.GetRequiredConfirmationHeight(ctx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeGatewayDeploymentConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address.Hex()),
			sdk.NewAttribute(types.AttributeKeyBytecodeHash, hex.EncodeToString(crypto.Keccak256(deploymentBytecode))),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
		),
	)

	return nil, nil
}

func (s msgServer) VoteConfirmGatewayDeployment(c context.Context, req *types.VoteConfirmGatewayDeploymentRequest) (*types.VoteConfirmGatewayDeploymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	keeper := s.ForChain(chain.Name)

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	switch {
	case poll.Is(vote.Expired):
		return &types.VoteConfirmGatewayDeploymentResponse{Log: fmt.Sprintf("vote for poll %s already expired", req.PollKey)}, nil
	case poll.Is(vote.Failed), poll.Is(vote.Completed):
		// If the voting threshold has been met and additional votes are received they should not return an error
		return &types.VoteConfirmGatewayDeploymentResponse{Log: fmt.Sprintf("vote for poll %s already decided", req.PollKey)}, nil
	}

	voteValue := &gogoprototypes.BoolValue{Value: req.Confirmed}
	if err := poll.Vote(voter, voteValue); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeGatewayDeploymentConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueVote),
		sdk.NewAttribute(types.AttributeKeyValue, strconv.FormatBool(voteValue.Value)),
	))

	if poll.Is(vote.Pending) {
		return &types.VoteConfirmGatewayDeploymentResponse{Log: fmt.Sprintf("not enough votes to confirm gateway for chain %s yet", chain.Name)}, nil
	}

	if poll.Is(vote.Failed) {
		if err := keeper.DeletePendingGateway(ctx); err != nil {
			return nil, err
		}

		return &types.VoteConfirmGatewayDeploymentResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	s.Logger(ctx).Info(fmt.Sprintf("%s gateway confirmation result is %t", chain.Name, confirmed.Value))

	address, ok := keeper.GetPendingGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("no pending gateway found")
	}

	// handle poll result
	event := sdk.NewEvent(
		types.EventTypeGatewayDeploymentConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
		sdk.NewAttribute(types.AttributeKeyAddress, address.Hex()),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))))
	defer func() { ctx.EventManager().EmitEvent(event) }()

	if !confirmed.Value {
		poll.AllowOverride()
		event = event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject))

		if err := keeper.DeletePendingGateway(ctx); err != nil {
			return nil, err
		}

		return &types.VoteConfirmGatewayDeploymentResponse{
			Log: fmt.Sprintf("%s gateway was discarded", chain.Name),
		}, nil
	}

	event = event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm))
	keeper.ConfirmPendingGateway(ctx)

	return &types.VoteConfirmGatewayDeploymentResponse{}, nil
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
			sdk.NewAttribute(types.AttributeKeyAsset, req.Asset.Name),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
		),
	)

	return &types.ConfirmTokenResponse{}, nil
}

func (s msgServer) ConfirmChain(c context.Context, req *types.ConfirmChainRequest) (*types.ConfirmChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if _, found := s.nexus.GetChain(ctx, req.Name); found {
		return nil, fmt.Errorf("chain '%s' is already confirmed", req.Name)
	}

	pendingChain, ok := s.GetPendingChain(ctx, req.Name)
	if !ok {
		return nil, fmt.Errorf("'%s' has not been added yet", req.Name)
	}

	keyRequirement, ok := s.tss.GetKeyRequirement(ctx, tss.MasterKey, pendingChain.Chain.KeyType)
	if !ok {
		return nil, fmt.Errorf("key requirement for key role %s type %s not found", tss.MasterKey.SimpleString(), pendingChain.Chain.KeyType)
	}

	snapshot, err := s.snapshotter.TakeSnapshot(ctx, keyRequirement)
	if err != nil {
		return nil, fmt.Errorf("unable to take snapshot: %v", err)
	}

	if err := pendingChain.Params.Validate(); err != nil {
		return nil, err
	}

	period := pendingChain.Params.RevoteLockingPeriod
	votingThreshold := pendingChain.Params.VotingThreshold
	minVoterCount := pendingChain.Params.MinVoterCount

	pollKey := vote.NewPollKey(types.ModuleName, pendingChain.Chain.Name)
	if err := s.voter.InitializePollWithSnapshot(
		ctx,
		pollKey,
		snapshot.Counter,
		vote.ExpiryAt(ctx.BlockHeight()+period),
		vote.Threshold(votingThreshold),
		vote.MinVoterCount(minVoterCount),
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeChainConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyChain, pendingChain.Chain.Name),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
		),
	)

	return &types.ConfirmChainResponse{}, nil
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

	_, state, ok := keeper.GetDeposit(ctx, common.Hash(req.TxID), common.Address(req.BurnerAddress))
	switch {
	case !ok:
		break
	case state == types.DepositStatus_Confirmed:
		return nil, fmt.Errorf("already confirmed")
	case state == types.DepositStatus_Burned:
		return nil, fmt.Errorf("already burned")
	}

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

	pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s_%s", req.TxID.Hex(), req.BurnerAddress.Hex(), req.Amount.String()))
	if err := s.voter.InitializePoll(
		ctx,
		pollKey,
		s.nexus.GetChainMaintainers(ctx, chain),
		vote.ExpiryAt(ctx.BlockHeight()+period),
		vote.Threshold(votingThreshold),
		vote.MinVoterCount(minVoterCount),
		vote.RewardPool(chain.Name),
	); err != nil {
		return nil, err
	}

	erc20Deposit := types.ERC20Deposit{
		TxID:             req.TxID,
		Amount:           req.Amount,
		Asset:            burnerInfo.Asset,
		DestinationChain: burnerInfo.DestinationChain,
		BurnerAddress:    req.BurnerAddress,
	}
	keeper.SetPendingDeposit(ctx, pollKey, &erc20Deposit)

	height, _ := keeper.GetRequiredConfirmationHeight(ctx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyAmount, req.Amount.String()),
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
	); err != nil {
		return nil, err
	}

	transferKey := types.TransferKey{
		TxID:      req.TxID,
		Type:      req.TransferType,
		NextKeyID: req.KeyID,
	}
	keeper.SetPendingTransferKey(ctx, pollKey, &transferKey)

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

	key, ok := s.signer.GetKey(ctx, req.KeyID)
	if !ok {
		return nil, fmt.Errorf("key %s does not exist", req.KeyID)
	}

	switch chain.KeyType {
	case tss.Threshold:
		pk, err := key.GetECDSAPubKey()
		if err != nil {
			return nil, err
		}

		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeyAddress, crypto.PubkeyToAddress(pk).Hex()),
			sdk.NewAttribute(types.AttributeKeyThreshold, ""),
		)
	case tss.Multisig:
		addresses, threshold, err := getMultisigAddresses(key)
		if err != nil {
			return nil, err
		}

		addressStrs := make([]string, len(addresses))
		for i, address := range addresses {
			addressStrs[i] = address.Hex()
		}

		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeyAddress, strings.Join(addressStrs, ",")),
			sdk.NewAttribute(types.AttributeKeyThreshold, strconv.FormatUint(uint64(threshold), 10)),
		)
	default:
		return nil, fmt.Errorf("uknown key type for chain %s", chain.Name)
	}

	return &types.ConfirmTransferKeyResponse{}, nil
}

func (s msgServer) VoteConfirmChain(c context.Context, req *types.VoteConfirmChainRequest) (*types.VoteConfirmChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	switch {
	case poll.Is(vote.Expired):
		return &types.VoteConfirmChainResponse{Log: fmt.Sprintf("vote for poll %s already expired", req.PollKey)}, nil
	case poll.Is(vote.Failed), poll.Is(vote.Completed):
		// if the voting threshold has been met and additional votes are received they should not return an error
		return &types.VoteConfirmChainResponse{Log: fmt.Sprintf("vote for poll %s already decided", req.PollKey)}, nil
	default:
	}

	pendingChain, chainFound := s.GetPendingChain(ctx, req.Name)
	if !chainFound {
		return nil, fmt.Errorf("unknown chain %s", req.Name)
	}

	voteValue := &gogoprototypes.BoolValue{Value: req.Confirmed}
	if err := poll.Vote(voter, voteValue); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeChainConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueVote),
		sdk.NewAttribute(types.AttributeKeyValue, strconv.FormatBool(voteValue.Value)),
	))

	if poll.Is(vote.Pending) {
		return &types.VoteConfirmChainResponse{Log: fmt.Sprintf("not enough votes to confirm chain %s yet", pendingChain.Chain.Name)}, nil
	}

	if poll.Is(vote.Failed) {
		s.DeletePendingChain(ctx, pendingChain.Chain.Name)
		return &types.VoteConfirmChainResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	s.Logger(ctx).Info(fmt.Sprintf("EVM chain %s confirmation result is %t", pendingChain.Chain.Name, confirmed.Value))
	s.DeletePendingChain(ctx, pendingChain.Chain.Name)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeChainConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, pendingChain.Chain.Name),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))))

	if !confirmed.Value {
		poll.AllowOverride()
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &types.VoteConfirmChainResponse{
			Log: fmt.Sprintf("chain %s was rejected", pendingChain.Chain.Name),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	s.nexus.SetChain(ctx, pendingChain.Chain)
	s.ForChain(pendingChain.Chain.Name).SetParams(ctx, pendingChain.Params)

	return &types.VoteConfirmChainResponse{}, nil
}

// VoteConfirmDeposit handles votes for deposit confirmations
func (s msgServer) VoteConfirmDeposit(c context.Context, req *types.VoteConfirmDepositRequest) (*types.VoteConfirmDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	switch {
	case poll.Is(vote.Expired):
		return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("vote for poll %s already %s", req.PollKey, vote.Expired.String())}, nil
	case poll.Is(vote.Failed), poll.Is(vote.Completed):
		// If the voting threshold has been met and additional votes are received they should not return an error
		return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("vote for poll %s already %s", req.PollKey, vote.Completed.String())}, nil
	default:
	}

	keeper := s.ForChain(chain.Name)
	pendingDeposit, _ := keeper.GetPendingDeposit(ctx, req.PollKey)

	// poll is pending
	if pendingDeposit.BurnerAddress != req.BurnAddress || pendingDeposit.TxID != req.TxID {
		return nil, fmt.Errorf("deposit in %s to address %s does not match poll %s", req.TxID.Hex(), req.BurnAddress.Hex(), req.PollKey.String())
	}

	_, ok = s.nexus.GetChain(ctx, pendingDeposit.DestinationChain)
	if !ok {
		return nil, fmt.Errorf("destination chain %s is not a registered chain", pendingDeposit.DestinationChain)
	}

	voteValue := &gogoprototypes.BoolValue{Value: req.Confirmed}
	if err := poll.Vote(voter, voteValue); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDepositConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueVote),
		sdk.NewAttribute(types.AttributeKeyValue, strconv.FormatBool(voteValue.Value)),
	))

	if poll.Is(vote.Pending) {
		return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("not enough votes to confirm deposit in %s to %s yet", req.TxID.Hex(), req.BurnAddress.Hex())}, nil
	}

	if poll.Is(vote.Failed) {
		keeper.DeletePendingDeposit(ctx, req.PollKey)
		return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	s.Logger(ctx).Info(fmt.Sprintf("%s deposit confirmation result for %s to %s is %t", chain.Name, pendingDeposit.TxID.Hex(), pendingDeposit.BurnerAddress.Hex(), confirmed.Value))
	keeper.DeletePendingDeposit(ctx, req.PollKey)

	depositAddr := nexus.CrossChainAddress{Address: pendingDeposit.BurnerAddress.Hex(), Chain: chain}
	recipient, ok := s.nexus.GetRecipient(ctx, depositAddr)
	if !ok {
		return nil, fmt.Errorf("cross-chain sender has no recipient")
	}

	height, ok := keeper.GetRequiredConfirmationHeight(ctx)
	if !ok {
		return nil, fmt.Errorf("could not find EVM subspace")
	}

	// handle poll result
	event := sdk.NewEvent(types.EventTypeDepositConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
		sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
		sdk.NewAttribute(types.AttributeKeyDestinationChain, recipient.Chain.Name),
		sdk.NewAttribute(types.AttributeKeyDestinationAddress, recipient.Address),
		sdk.NewAttribute(types.AttributeKeyAmount, pendingDeposit.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyDepositAddress, depositAddr.Address),
		sdk.NewAttribute(types.AttributeKeyTxID, pendingDeposit.TxID.Hex()),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))))

	burnerInfo := keeper.GetBurnerInfo(ctx, req.BurnAddress)
	if burnerInfo != nil {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyTokenAddress, burnerInfo.TokenAddress.Hex()))
	}

	defer func() { ctx.EventManager().EmitEvent(event) }()

	if !confirmed.Value {
		poll.AllowOverride()
		event = event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject))
		return &types.VoteConfirmDepositResponse{
			Log: fmt.Sprintf("deposit in %s to %s was discarded", pendingDeposit.TxID.Hex(), req.BurnAddress.Hex()),
		}, nil
	}

	event = event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm))

	feeRate, ok := keeper.GetTransactionFeeRate(ctx)
	if !ok {
		return nil, fmt.Errorf("could not retrieve transaction fee rate")
	}

	amount := sdk.NewCoin(pendingDeposit.Asset, sdk.NewIntFromBigInt(pendingDeposit.Amount.BigInt()))
	transferID, err := s.nexus.EnqueueForTransfer(ctx, depositAddr, amount, feeRate)
	if err != nil {
		return nil, err
	}

	event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyTransferID, transferID.String()))

	s.Logger(ctx).Info(fmt.Sprintf("deposit confirmed on chain %s for %s to %s with transfer ID %d and command ID %s", chain.Name, pendingDeposit.TxID.Hex(), depositAddr.Address, transferID, types.TransferIDtoCommandID(transferID).Hex()))
	keeper.SetDeposit(ctx, pendingDeposit, types.DepositStatus_Confirmed)

	return &types.VoteConfirmDepositResponse{}, nil
}

// VoteConfirmToken handles votes for token deployment confirmations
func (s msgServer) VoteConfirmToken(c context.Context, req *types.VoteConfirmTokenRequest) (*types.VoteConfirmTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	keeper := s.ForChain(chain.Name)
	token := keeper.GetERC20TokenByAsset(ctx, req.Asset)

	poll := s.voter.GetPoll(ctx, req.PollKey)
	switch {
	case poll.Is(vote.Expired):
		return &types.VoteConfirmTokenResponse{Log: fmt.Sprintf("vote for poll %s already expired", req.PollKey)}, nil
	case poll.Is(vote.Failed), poll.Is(vote.Completed):
		// If the voting threshold has been met and additional votes are received they should not return an error
		return &types.VoteConfirmTokenResponse{Log: fmt.Sprintf("vote for poll %s already decided", req.PollKey)}, nil
	default:
		if types.GetConfirmTokenKey(token.GetTxID(), token.GetAsset()) != req.PollKey {
			return nil, fmt.Errorf("poll key mismatch (expected %s, got %s)", types.GetConfirmTokenKey(token.GetTxID(), token.GetAsset()).String(), req.PollKey.String())
		}
	}

	voteValue := &gogoprototypes.BoolValue{Value: req.Confirmed}
	if err := poll.Vote(voter, voteValue); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTokenConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueVote),
		sdk.NewAttribute(types.AttributeKeyValue, strconv.FormatBool(voteValue.Value)),
	))

	if poll.Is(vote.Pending) {
		return &types.VoteConfirmTokenResponse{Log: fmt.Sprintf("not enough votes to confirm token %s yet", req.Asset)}, nil
	}

	if poll.Is(vote.Failed) {
		token.RejectDeployment()
		return &types.VoteConfirmTokenResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	s.Logger(ctx).Info(fmt.Sprintf("token %s deployment confirmation result on chain %s is %t", req.Asset, chain.Name, confirmed.Value))

	// handle poll result
	event := sdk.NewEvent(types.EventTypeTokenConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))))

	if !confirmed.Value {
		poll.AllowOverride()
		token.RejectDeployment()
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &types.VoteConfirmTokenResponse{
			Log: fmt.Sprintf("token %s was discarded", req.Asset),
		}, nil
	}

	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	token.ConfirmDeployment()

	return &types.VoteConfirmTokenResponse{
		Log: fmt.Sprintf("token %s deployment confirmed", req.Asset)}, nil
}

// VoteConfirmTransferKey handles votes for transfer ownership/operatorship confirmations
func (s msgServer) VoteConfirmTransferKey(c context.Context, req *types.VoteConfirmTransferKeyRequest) (*types.VoteConfirmTransferKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	if err := validateChainActivated(ctx, s.nexus, chain); err != nil {
		return nil, err
	}

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	switch {
	case poll.Is(vote.Expired):
		return &types.VoteConfirmTransferKeyResponse{Log: fmt.Sprintf("vote for poll %s already expired", req.PollKey)}, nil
	case poll.Is(vote.Failed), poll.Is(vote.Completed):
		// If the voting threshold has been met and additional votes are received they should not return an error
		return &types.VoteConfirmTransferKeyResponse{Log: fmt.Sprintf("vote for poll %s already decided", req.PollKey)}, nil
	default:
	}

	keeper := s.ForChain(chain.Name)
	pendingTransfer, _ := keeper.GetPendingTransferKey(ctx, req.PollKey)

	var keyRole tss.KeyRole
	keyRole = s.signer.GetKeyRole(ctx, pendingTransfer.NextKeyID)
	if keyRole == tss.Unknown {
		return nil, fmt.Errorf("key %s cannot be found", pendingTransfer.NextKeyID)
	}

	voteValue := &gogoprototypes.BoolValue{Value: req.Confirmed}
	if err := poll.Vote(voter, voteValue); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransferKeyConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueVote),
		sdk.NewAttribute(types.AttributeKeyValue, strconv.FormatBool(voteValue.Value)),
	))

	if poll.Is(vote.Pending) {
		return &types.VoteConfirmTransferKeyResponse{Log: fmt.Sprintf("not enough votes to confirm transfer key in poll %s yet", req.PollKey.String())}, nil
	}

	if poll.Is(vote.Failed) {
		keeper.DeletePendingTransferKey(ctx, req.PollKey)
		return &types.VoteConfirmTransferKeyResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	// handle poll result
	event := sdk.NewEvent(types.EventTypeTransferKeyConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
		sdk.NewAttribute(types.AttributeKeyTransferKeyType, pendingTransfer.Type.SimpleString()),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))))

	if !confirmed.Value {
		poll.AllowOverride()
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))

		msg := fmt.Sprintf("failed to confirmed %s key transfer for chain %s", keyRole.SimpleString(), chain.Name)
		s.Logger(ctx).Error(msg)
		return &types.VoteConfirmTransferKeyResponse{Log: msg}, nil

	}

	keeper.ArchiveTransferKey(ctx, req.PollKey)
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	if err := s.signer.RotateKey(ctx, chain, keyRole); err != nil {
		return nil, err
	}

	s.Logger(ctx).Info(fmt.Sprintf("successfully confirmed %s key transfer for chain %s",
		keyRole.SimpleString(), chain.Name), "txID", pendingTransfer.TxID.Hex(), "rotation count", s.tss.GetRotationCount(ctx, chain, keyRole))
	return &types.VoteConfirmTransferKeyResponse{}, nil
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

	if err = s.nexus.RegisterAsset(ctx, chain, nexus.NewAsset(req.Asset.Name, req.MinAmount, false)); err != nil {
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

		cmd, err := types.CreateBurnTokenCommand(chainID, secondaryKeyID, ctx.BlockHeight(), *burnerInfo)
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

func getMultisigThreshold(keyCount int, threshold utils.Threshold) uint8 {
	return uint8(
		sdk.NewDec(int64(keyCount)).
			MulInt64(threshold.Numerator).
			QuoInt64(threshold.Denominator).
			Ceil().
			RoundInt64(),
	)
}

func getMultisigAddresses(key tss.Key) ([]common.Address, uint8, error) {
	multisigPubKeys, err := key.GetMultisigPubKey()
	if err != nil {
		return nil, 0, sdkerrors.Wrapf(types.ErrEVM, err.Error())
	}

	threshold := uint8(key.GetMultisigKey().Threshold)
	return types.KeysToAddresses(multisigPubKeys...), threshold, nil
}

func getGatewayDeploymentBytecode(ctx sdk.Context, k types.ChainKeeper, s types.Signer, chain nexus.Chain) ([]byte, error) {
	externalKeyIDs, ok := s.GetExternalKeyIDs(ctx, chain)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("no %s keys for chain %s found", tss.ExternalKey.SimpleString(), chain.Name))
	}

	externalPubKeys := make([]ecdsa.PublicKey, len(externalKeyIDs))
	for i, externalKeyID := range externalKeyIDs {
		externalKey, ok := s.GetKey(ctx, externalKeyID)
		if !ok {
			return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s key %s for chain %s not found", tss.ExternalKey.SimpleString(), externalKeyID, chain.Name))
		}

		pk, err := externalKey.GetECDSAPubKey()
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
		}

		externalPubKeys[i] = pk
	}
	externalKeyAddresses := types.KeysToAddresses(externalPubKeys...)
	externalKeyThreshold := getMultisigThreshold(len(externalKeyAddresses), s.GetExternalMultisigThreshold(ctx))

	bz, _ := k.GetGatewayByteCode(ctx)

	masterKey, ok := s.GetCurrentKey(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("no %s key for chain %s found", tss.MasterKey.SimpleString(), chain.Name))
	}

	secondaryKey, ok := s.GetCurrentKey(ctx, chain, tss.SecondaryKey)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("no %s key for chain %s found", tss.SecondaryKey.SimpleString(), chain.Name))
	}

	switch chain.KeyType {
	case tss.Threshold:
		masterPubKey, err := masterKey.GetECDSAPubKey()
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
		}

		secondaryPubKey, err := secondaryKey.GetECDSAPubKey()
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
		}

		return types.GetSinglesigGatewayDeploymentBytecode(
			bz,
			externalKeyAddresses,
			uint8(s.GetExternalMultisigThreshold(ctx).Numerator),
			crypto.PubkeyToAddress(masterPubKey),
			crypto.PubkeyToAddress(secondaryPubKey),
		)
	case tss.Multisig:
		masterMultisigAddresses, masterMultisigThreshold, err := getMultisigAddresses(masterKey)
		if err != nil {
			return nil, err
		}

		secondaryMultisigAddresses, secondaryMultisigThreshold, err := getMultisigAddresses(secondaryKey)
		if err != nil {
			return nil, err
		}

		return types.GetMultisigGatewayDeploymentBytecode(
			bz,
			externalKeyAddresses,
			externalKeyThreshold,
			masterMultisigAddresses,
			masterMultisigThreshold,
			secondaryMultisigAddresses,
			secondaryMultisigThreshold,
		)
	default:
		return nil, fmt.Errorf("unknown key type set for chain %s", chain.Name)
	}
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

		s.Logger(ctx).Info(fmt.Sprintf("storing data for mint command %s", cmd.ID.Hex()))

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
		addresses, threshold, err := getMultisigAddresses(nextKey)
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

	sigMetadata := types.SigMetadata{
		Type:  types.SigCommand,
		Chain: chain.Name,
	}

	batchedCommandsIDHex := hex.EncodeToString(commandBatch.GetID())
	err = s.signer.StartSign(ctx, tss.SignInfo{
		KeyID:           commandBatch.GetKeyID(),
		SigID:           batchedCommandsIDHex,
		Msg:             commandBatch.GetSigHash().Bytes(),
		SnapshotCounter: counter,
		RequestModule:   types.ModuleName,
		Metadata:        string(types.ModuleCdc.MustMarshalJSON(&sigMetadata)),
	}, s.snapshotter, s.voter)
	if err != nil {
		return nil, err
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

	s.SetPendingChain(
		ctx,
		nexus.Chain{Name: req.Name, SupportsForeignAssets: true, KeyType: req.KeyType, Module: types.ModuleName},
		req.Params,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeNewChain,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueUpdate),
			sdk.NewAttribute(types.AttributeKeyChain, req.Name),
		),
	)

	return &types.AddChainResponse{}, nil
}
