package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
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

func (s msgServer) Link(c context.Context, req *types.LinkRequest) (*types.LinkResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	senderChain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	keeper := s.ForChain(ctx, senderChain.Name)

	gatewayAddr, ok := keeper.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	recipientChain, ok := s.nexus.GetChain(ctx, req.RecipientChain)
	if !ok {
		return nil, fmt.Errorf("unknown recipient chain")
	}

	found := s.nexus.IsAssetRegistered(ctx, recipientChain.Name, req.Asset)
	if !found {
		return nil, fmt.Errorf("asset '%s' not registered for chain '%s'", req.Asset, recipientChain.Name)
	}

	tokenAddr, err := keeper.GetTokenAddress(ctx, req.Asset, gatewayAddr)
	if err != nil {
		return nil, err
	}

	burnerAddr, salt, err := keeper.GetBurnerAddressAndSalt(ctx, tokenAddr, req.RecipientAddr, gatewayAddr)
	if err != nil {
		return nil, err
	}

	symbol, ok := keeper.GetTokenSymbol(ctx, req.Asset)
	if !ok {
		return nil, fmt.Errorf("Could not retrieve symbol for token %s", req.Asset)
	}

	s.nexus.LinkAddresses(ctx,
		nexus.CrossChainAddress{Chain: senderChain, Address: burnerAddr.String()},
		nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr})

	burnerInfo := types.BurnerInfo{
		TokenAddress:     types.Address(tokenAddr),
		DestinationChain: req.RecipientChain,
		Symbol:           symbol,
		Salt:             types.Hash(salt),
	}
	keeper.SetBurnerInfo(ctx, burnerAddr, &burnerInfo)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyBurnAddress, burnerAddr.String()),
			sdk.NewAttribute(types.AttributeKeyAddress, req.RecipientAddr),
		),
	)

	return &types.LinkResponse{DepositAddr: burnerAddr.Hex()}, nil
}

// ConfirmToken handles token deployment confirmation
func (s msgServer) ConfirmToken(c context.Context, req *types.ConfirmTokenRequest) (*types.ConfirmTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	originChain, ok := s.nexus.GetChain(ctx, req.OriginChain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.OriginChain)
	}

	if s.nexus.IsAssetRegistered(ctx, chain.Name, originChain.NativeAsset) {
		return nil, fmt.Errorf("token %s is already registered", originChain.NativeAsset)
	}

	keeper := s.ForChain(ctx, chain.Name)

	gatewayAddr, ok := keeper.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	tokenAddr, err := keeper.GetTokenAddress(ctx, originChain.NativeAsset, gatewayAddr)
	if err != nil {
		return nil, err
	}

	keyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	seqNo, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot seqNo for key ID %s registered", keyID)
	}

	pollKey := vote.NewPollKey(types.ModuleName, req.TxID.Hex()+"_"+originChain.NativeAsset)

	period, ok := keeper.GetRevoteLockingPeriod(ctx)
	if !ok {
		return nil, fmt.Errorf("Could not retrieve revote locking period for chain %s", req.Chain)
	}

	if err := s.voter.InitializePoll(ctx, pollKey, seqNo, vote.ExpiryAt(ctx.BlockHeight()+period)); err != nil {
		return nil, err
	}

	symbol, ok := keeper.GetTokenSymbol(ctx, originChain.NativeAsset)
	if !ok {
		return nil, fmt.Errorf("Could not retrieve symbol for token %s", originChain.NativeAsset)
	}

	deploy := types.ERC20TokenDeployment{
		Asset:        originChain.NativeAsset,
		TokenAddress: types.Address(tokenAddr),
	}
	keeper.SetPendingTokenDeployment(ctx, pollKey, deploy)

	height, _ := keeper.GetRequiredConfirmationHeight(ctx)

	telemetry.NewLabel("eth_token_addr", tokenAddr.String())

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTokenConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyGatewayAddress, gatewayAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, tokenAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeySymbol, symbol),
			sdk.NewAttribute(types.AttributeKeyAsset, originChain.NativeAsset),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
		),
	)

	return &types.ConfirmTokenResponse{}, nil
}

func (s msgServer) ConfirmChain(c context.Context, req *types.ConfirmChainRequest) (*types.ConfirmChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if _, found := s.nexus.GetChain(ctx, req.Name); found {
		return &types.ConfirmChainResponse{}, fmt.Errorf("chain '%s' is already confirmed", req.Name)
	}

	if _, ok := s.GetPendingChain(ctx, req.Name); !ok {
		return &types.ConfirmChainResponse{}, fmt.Errorf("'%s' has not been added yet", req.Name)
	}

	seqNo := s.snapshotter.GetLatestCounter(ctx)
	if seqNo < 0 {
		_, _, err := s.snapshotter.TakeSnapshot(ctx, 0, tss.WeightedByStake)
		if err != nil {
			return nil, fmt.Errorf("unable to take snapshot: %v", err)
		}
		seqNo = s.snapshotter.GetLatestCounter(ctx)
	}
	keeper := s.ForChain(ctx, req.Name)

	period, ok := keeper.GetRevoteLockingPeriod(ctx)
	if !ok {
		return nil, fmt.Errorf("Could not retrieve revote locking period for chain %s", req.Name)
	}

	pollKey := vote.NewPollKey(types.ModuleName, req.Name)
	if err := s.voter.InitializePoll(ctx, pollKey, seqNo, vote.ExpiryAt(ctx.BlockHeight()+period)); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeChainConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyChain, req.Name),
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

	keeper := s.ForChain(ctx, chain.Name)

	_, state, ok := keeper.GetDeposit(ctx, common.Hash(req.TxID), common.Address(req.BurnerAddress))
	switch {
	case !ok:
		break
	case state == types.CONFIRMED:
		return nil, fmt.Errorf("already confirmed")
	case state == types.BURNED:
		return nil, fmt.Errorf("already burned")
	}

	burnerInfo := keeper.GetBurnerInfo(ctx, common.Address(req.BurnerAddress))
	if burnerInfo == nil {
		return nil, fmt.Errorf("no burner info found for address %s", req.BurnerAddress)
	}

	keyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	seqNo, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot seqNo for key ID %s registered", keyID)
	}

	period, ok := keeper.GetRevoteLockingPeriod(ctx)
	if !ok {
		return nil, fmt.Errorf("Could not retrieve revote locking period for chain %s", req.Chain)
	}

	pollKey := vote.NewPollKey(types.ModuleName, req.TxID.Hex()+"_"+req.BurnerAddress.Hex())
	if err := s.voter.InitializePoll(ctx, pollKey, seqNo, vote.ExpiryAt(ctx.BlockHeight()+period)); err != nil {
		return nil, err
	}

	erc20Deposit := types.ERC20Deposit{
		TxID:             req.TxID,
		Amount:           req.Amount,
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
			sdk.NewAttribute(types.AttributeKeyBurnAddress, req.BurnerAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, burnerInfo.TokenAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
		),
	)

	return &types.ConfirmDepositResponse{}, nil
}

// ConfirmTransferOwnership handles transfer ownership confirmations
func (s msgServer) ConfirmTransferOwnership(c context.Context, req *types.ConfirmTransferOwnershipRequest) (*types.ConfirmTransferOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	pk, ok := s.signer.GetKey(ctx, req.KeyID)
	if !ok {
		return nil, fmt.Errorf("key %s does not exist (yet)", req.KeyID)
	}

	_, ok = s.signer.GetNextKey(ctx, chain, tss.MasterKey)
	if ok {
		return nil, fmt.Errorf("next %s key for chain %s already set", tss.MasterKey.SimpleString(), chain.Name)
	}

	keeper := s.ForChain(ctx, chain.Name)

	gatewayAddr, ok := keeper.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	currentKeyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	seqNo, ok := s.signer.GetSnapshotCounterForKeyID(ctx, currentKeyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot seqNo for key ID %s registered", currentKeyID)
	}

	period, ok := keeper.GetRevoteLockingPeriod(ctx)
	if !ok {
		return nil, fmt.Errorf("Could not retrieve revote locking period for chain %s", req.Chain)
	}

	pollKey := vote.NewPollKey(types.ModuleName, req.TxID.Hex()+"_"+req.KeyID)
	if err := s.voter.InitializePoll(ctx, pollKey, seqNo, vote.ExpiryAt(ctx.BlockHeight()+period)); err != nil {
		return nil, err
	}

	transferOwnership := types.TransferOwnership{
		TxID:      req.TxID,
		NextKeyID: pk.ID,
	}
	keeper.SetPendingTransferOwnership(ctx, pollKey, &transferOwnership)

	height, _ := keeper.GetRequiredConfirmationHeight(ctx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTransferOwnershipConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyGatewayAddress, gatewayAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeyAddress, crypto.PubkeyToAddress(pk.Value).Hex()),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
		),
	)

	return &types.ConfirmTransferOwnershipResponse{}, nil
}

func (s msgServer) VoteConfirmChain(c context.Context, req *types.VoteConfirmChainRequest) (*types.VoteConfirmChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	registeredChain, registered := s.nexus.GetChain(ctx, req.Name)
	if registered {
		return &types.VoteConfirmChainResponse{Log: fmt.Sprintf("chain %s already confirmed", registeredChain.Name)}, nil
	}
	chain, ok := s.GetPendingChain(ctx, req.Name)
	if !ok {
		return nil, fmt.Errorf("unknown chain %s", req.Name)
	}

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	if err := poll.Vote(voter, &gogoprototypes.BoolValue{Value: req.Confirmed}); err != nil {
		return nil, err
	}

	if poll.Is(vote.Pending) {
		return &types.VoteConfirmChainResponse{Log: fmt.Sprintf("not enough votes to confirm chain in %s yet", req.Name)}, nil
	}

	if poll.Is(vote.Failed) {
		s.DeletePendingChain(ctx, req.Name)
		return &types.VoteConfirmChainResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	s.Logger(ctx).Info(fmt.Sprintf("EVM chain confirmation result is %s", poll.GetResult()))
	s.DeletePendingChain(ctx, req.Name)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeChainConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, req.Name),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))))

	if !confirmed.Value {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &types.VoteConfirmChainResponse{
			Log: fmt.Sprintf("chain %s was rejected", req.Name),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	s.nexus.SetChain(ctx, chain)
	s.nexus.RegisterAsset(ctx, chain.Name, chain.NativeAsset)

	return &types.VoteConfirmChainResponse{}, nil
}

// VoteConfirmDeposit handles votes for deposit confirmations
func (s msgServer) VoteConfirmDeposit(c context.Context, req *types.VoteConfirmDepositRequest) (*types.VoteConfirmDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	keeper := s.ForChain(ctx, chain.Name)

	pendingDeposit, pollFound := keeper.GetPendingDeposit(ctx, req.PollKey)

	destChain, ok := s.nexus.GetChain(ctx, pendingDeposit.DestinationChain)
	if !ok {
		return nil, fmt.Errorf("destination chain %s is not a registered chain", pendingDeposit.DestinationChain)
	}

	confirmedDeposit, state, depositFound := keeper.GetDeposit(ctx, common.Hash(req.TxID), common.Address(req.BurnAddress))

	switch {
	// a malicious user could try to delete an ongoing poll by providing an already confirmed deposit,
	// so we need to check that it matches the poll before deleting
	case depositFound && pollFound && confirmedDeposit == pendingDeposit:
		keeper.DeletePendingDeposit(ctx, req.PollKey)
		fallthrough
	// If the voting threshold has been met and additional votes are received they should not return an error
	case depositFound:
		switch state {
		case types.CONFIRMED:
			return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("deposit in %s to address %s already confirmed", pendingDeposit.TxID.Hex(), pendingDeposit.BurnerAddress.Hex())}, nil
		case types.BURNED:
			return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("deposit in %s to address %s already spent", pendingDeposit.TxID.Hex(), pendingDeposit.BurnerAddress.Hex())}, nil
		}
	case !pollFound:
		return nil, fmt.Errorf("no deposit found for poll %s", req.PollKey.String())
	case pendingDeposit.BurnerAddress != req.BurnAddress || pendingDeposit.TxID != req.TxID:
		return nil, fmt.Errorf("deposit in %s to address %s does not match poll %s", req.TxID, req.BurnAddress.Hex(), req.PollKey.String())
	default:
		// assert: the deposit is known and has not been confirmed before
	}

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	if err := poll.Vote(voter, &gogoprototypes.BoolValue{Value: req.Confirmed}); err != nil {
		return nil, err
	}

	if poll.Is(vote.Pending) {
		return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("not enough votes to confirm deposit in %s to %s yet", req.TxID, req.BurnAddress.Hex())}, nil
	}

	if poll.Is(vote.Failed) {
		keeper.DeletePendingDeposit(ctx, req.PollKey)
		return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	s.Logger(ctx).Info(fmt.Sprintf("%s deposit confirmation result is %s", chain.Name, poll.GetResult()))
	keeper.DeletePendingDeposit(ctx, req.PollKey)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeDepositConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))))

	if !confirmed.Value {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &types.VoteConfirmDepositResponse{
			Log: fmt.Sprintf("deposit in %s to %s was discarded", req.TxID, req.BurnAddress.Hex()),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	depositAddr := nexus.CrossChainAddress{Address: pendingDeposit.BurnerAddress.Hex(), Chain: chain}
	amount := sdk.NewInt64Coin(destChain.NativeAsset, pendingDeposit.Amount.BigInt().Int64())
	if err := s.nexus.EnqueueForTransfer(ctx, depositAddr, amount); err != nil {
		return nil, err
	}
	keeper.SetDeposit(ctx, pendingDeposit, types.CONFIRMED)

	return &types.VoteConfirmDepositResponse{}, nil
}

// VoteConfirmToken handles votes for token deployment confirmations
func (s msgServer) VoteConfirmToken(c context.Context, req *types.VoteConfirmTokenRequest) (*types.VoteConfirmTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	keeper := s.ForChain(ctx, chain.Name)

	// is there an ongoing poll?
	token, pollFound := keeper.GetPendingTokenDeployment(ctx, req.PollKey)
	registered := s.nexus.IsAssetRegistered(ctx, chain.Name, req.Asset)
	switch {
	// a malicious user could try to delete an ongoing poll by providing an already confirmed token,
	// so we need to check that it matches the poll before deleting
	case registered && pollFound && token.Asset == req.Asset:
		keeper.DeletePendingToken(ctx, req.PollKey)
		fallthrough
	// If the voting threshold has been met and additional votes are received they should not return an error
	case registered:
		return &types.VoteConfirmTokenResponse{Log: fmt.Sprintf("token %s already confirmed", req.Asset)}, nil
	case !pollFound:
		return nil, fmt.Errorf("no token found for poll %s", req.PollKey.String())
	case token.Asset != req.Asset:
		return nil, fmt.Errorf("token %s does not match poll %s", req.Asset, req.PollKey.String())
	default:
		// assert: the token is known and has not been confirmed before
	}

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	if err := poll.Vote(voter, &gogoprototypes.BoolValue{Value: req.Confirmed}); err != nil {
		return nil, err
	}

	if poll.Is(vote.Pending) {
		return &types.VoteConfirmTokenResponse{Log: fmt.Sprintf("not enough votes to confirm token %s yet", req.Asset)}, nil
	}

	if poll.Is(vote.Failed) {
		keeper.DeletePendingToken(ctx, req.PollKey)
		return &types.VoteConfirmTokenResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	s.Logger(ctx).Info(fmt.Sprintf("token deployment confirmation result is %s", poll.GetResult()))
	keeper.DeletePendingToken(ctx, req.PollKey)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeTokenConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))))

	if !confirmed.Value {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &types.VoteConfirmTokenResponse{
			Log: fmt.Sprintf("token %s was discarded", req.Asset),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	s.nexus.RegisterAsset(ctx, chain.Name, token.Asset)

	return &types.VoteConfirmTokenResponse{
		Log: fmt.Sprintf("token %s deployment confirmed", token.Asset)}, nil
}

// VoteConfirmTransferOwnership handles votes for transfer ownership confirmations
func (s msgServer) VoteConfirmTransferOwnership(c context.Context, req *types.VoteConfirmTransferOwnershipRequest) (*types.VoteConfirmTransferOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	keeper := s.ForChain(ctx, chain.Name)

	pendingTransferOwnership, pendingTransferFound := keeper.GetPendingTransferOwnership(ctx, req.PollKey)
	archivedTransferOwnership, archivedtransferFound := keeper.GetArchivedTransferOwnership(ctx, req.PollKey)

	switch {
	case !pendingTransferFound && !archivedtransferFound:
		return nil, fmt.Errorf("no transfer ownership found for poll %s", req.PollKey.String())
	// If the voting threshold has been met and additional votes are received they should not return an error
	case archivedtransferFound:
		return &types.VoteConfirmTransferOwnershipResponse{Log: fmt.Sprintf("transfer ownership in %s to keyID %s already confirmed", archivedTransferOwnership.TxID.Hex(), archivedTransferOwnership.NextKeyID)}, nil
	case pendingTransferFound:
		pk, ok := s.signer.GetKey(ctx, pendingTransferOwnership.NextKeyID)
		if !ok {
			return nil, fmt.Errorf("key %s cannot be found", pendingTransferOwnership.NextKeyID)
		}
		if crypto.PubkeyToAddress(pk.Value) != common.Address(req.NewOwnerAddress) || pendingTransferOwnership.TxID != req.TxID {
			return nil, fmt.Errorf("transfer ownership in %s to address %s does not match poll %s", req.TxID, req.NewOwnerAddress.Hex(), req.PollKey.String())
		}
	default:
		// assert: the transfer ownership is known and has not been confirmed before
	}

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	if err := poll.Vote(voter, &gogoprototypes.BoolValue{Value: req.Confirmed}); err != nil {
		return nil, err
	}

	if poll.Is(vote.Pending) {
		return &types.VoteConfirmTransferOwnershipResponse{Log: fmt.Sprintf("not enough votes to confirm transfer ownership in %s to %s yet", req.TxID, req.NewOwnerAddress.Hex())}, nil
	}

	if poll.Is(vote.Failed) {
		keeper.DeletePendingTransferOwnership(ctx, req.PollKey)
		return &types.VoteConfirmTransferOwnershipResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	// TODO: handle rejected case

	s.Logger(ctx).Info(fmt.Sprintf("%s transfer ownership confirmation result is %s", chain.Name, poll.GetResult()))
	keeper.ArchiveTransferOwnership(ctx, req.PollKey)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeTransferOwnershipConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))))

	if !confirmed.Value {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &types.VoteConfirmTransferOwnershipResponse{
			Log: fmt.Sprintf("transfer ownership in %s to %s was discarded", req.TxID, req.NewOwnerAddress.Hex()),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	if err := s.signer.AssignNextKey(ctx, chain, tss.MasterKey, pendingTransferOwnership.NextKeyID); err != nil {
		return nil, err
	}
	return &types.VoteConfirmTransferOwnershipResponse{}, nil
}

func (s msgServer) SignDeployToken(c context.Context, req *types.SignDeployTokenRequest) (*types.SignDeployTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	chainID := s.getChainID(ctx, req.Chain)
	if chainID == nil {
		return nil, fmt.Errorf("Could not find chain ID for '%s'", req.Chain)
	}

	originChain, found := s.nexus.GetChain(ctx, req.OriginChain)
	if !found {
		return nil, fmt.Errorf("%s is not a registered chain", req.OriginChain)
	}

	commandID := getCommandID([]byte(req.TokenName), chainID)

	data, err := types.CreateDeployTokenCommandData(chainID, commandID, req.TokenName, req.Symbol, req.Decimals, req.Capacity)
	if err != nil {
		return nil, err
	}

	keyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	keeper := s.ForChain(ctx, chain.Name)

	commandIDHex := common.Bytes2Hex(commandID[:])
	s.Logger(ctx).Info(fmt.Sprintf("storing data for deploy-token command %s", commandIDHex))
	keeper.SetCommandData(ctx, commandID, data)

	signHash := types.GetSignHash(data)

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = s.signer.StartSign(ctx, s.voter, keyID, commandIDHex, signHash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	keeper.SetTokenInfo(ctx, originChain.NativeAsset, req)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)
	return &types.SignDeployTokenResponse{CommandID: commandID[:]}, nil
}

func (s msgServer) SignBurnTokens(c context.Context, req *types.SignBurnTokensRequest) (*types.SignBurnTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}
	keeper := s.ForChain(ctx, chain.Name)

	deposits := keeper.GetConfirmedDeposits(ctx)

	if len(deposits) == 0 {
		return &types.SignBurnTokensResponse{}, nil
	}

	chainID := s.getChainID(ctx, req.Chain)
	if chainID == nil {
		return nil, fmt.Errorf("Could not find chain ID for '%s'", req.Chain)
	}

	var burnerInfos []types.BurnerInfo
	seen := map[string]bool{}
	for _, deposit := range deposits {
		if seen[deposit.BurnerAddress.Hex()] {
			continue
		}
		burnerInfo := keeper.GetBurnerInfo(ctx, common.Address(deposit.BurnerAddress))
		if burnerInfo == nil {
			return nil, fmt.Errorf("no burner info found for address %s", deposit.BurnerAddress.Hex())
		}
		burnerInfos = append(burnerInfos, *burnerInfo)
		seen[deposit.BurnerAddress.Hex()] = true
	}

	data, err := types.CreateBurnCommandData(chainID, ctx.BlockHeight(), burnerInfos)
	if err != nil {
		return nil, err
	}

	commandID := getCommandID(data, chainID)

	keyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	s.Logger(ctx).Info(fmt.Sprintf("storing data for burn command %s", commandIDHex))
	keeper.SetCommandData(ctx, commandID, data)

	s.Logger(ctx).Info(fmt.Sprintf("signing burn command [%s] for token deposits to chain %s", commandIDHex, chain.Name))
	signHash := types.GetSignHash(data)

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = s.signer.StartSign(ctx, s.voter, keyID, commandIDHex, signHash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	for _, deposit := range deposits {
		keeper.DeleteDeposit(ctx, deposit)
		keeper.SetDeposit(ctx, deposit, types.BURNED)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)
	return &types.SignBurnTokensResponse{CommandID: commandID[:]}, nil
}

func (s msgServer) SignTx(c context.Context, req *types.SignTxRequest) (*types.SignTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	tx := req.UnmarshaledTx()
	txID := tx.Hash().String()
	keeper := s.ForChain(ctx, chain.Name)

	keeper.SetUnsignedTx(ctx, txID, tx)
	s.Logger(ctx).Info(fmt.Sprintf("storing raw tx %s", txID))
	hash, err := keeper.GetHashToSign(ctx, txID)
	if err != nil {
		return nil, err
	}

	s.Logger(ctx).Info(fmt.Sprintf("%s tx [%s] to sign: %s", chain.Name, txID, hash.Hex()))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyTxID, txID),
		),
	)

	keyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = s.signer.StartSign(ctx, s.voter, keyID, txID, hash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	byteCodes, ok := keeper.GetGatewayByteCodes(ctx)
	if !ok {
		return nil, fmt.Errorf("Could not retrieve gateway bytecodes for chain %s", req.Chain)
	}

	// if this is the transaction that is deploying Axelar Gateway, calculate and save address
	// TODO: this is something that should be done after the signature has been successfully confirmed
	if tx.To() == nil && bytes.Equal(tx.Data(), byteCodes) {

		pub, ok := s.signer.GetCurrentKey(ctx, chain, tss.MasterKey)
		if !ok {
			return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
		}

		addr := crypto.CreateAddress(crypto.PubkeyToAddress(pub.Value), tx.Nonce())
		keeper.SetGatewayAddress(ctx, addr)

		telemetry.NewLabel("eth_factory_addr", addr.String())
	}

	return &types.SignTxResponse{TxID: txID}, nil
}

func (s msgServer) SignPendingTransfers(c context.Context, req *types.SignPendingTransfersRequest) (*types.SignPendingTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	pendingTransfers := s.nexus.GetTransfersForChain(ctx, chain, nexus.Pending)

	if len(pendingTransfers) == 0 {
		return &types.SignPendingTransfersResponse{}, nil
	}

	chainID := s.getChainID(ctx, req.Chain)
	if chainID == nil {
		return nil, fmt.Errorf("Could not find chain ID for '%s'", req.Chain)
	}

	keeper := s.ForChain(ctx, chain.Name)

	getRecipientAndAsset := func(transfer nexus.CrossChainTransfer) string {
		return fmt.Sprintf("%s-%s", transfer.Recipient.Address, transfer.Asset.Denom)
	}

	retrieveSymbol := func(denom string) (string, bool) {
		return keeper.GetTokenSymbol(ctx, denom)
	}

	data, err := types.CreateMintCommandData(chainID, nexus.MergeTransfersBy(pendingTransfers, getRecipientAndAsset), retrieveSymbol)
	if err != nil {
		return nil, err
	}

	commandID := getCommandID(data, chainID)

	keyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])

	s.Logger(ctx).Info(fmt.Sprintf("storing data for mint command %s", commandIDHex))
	keeper.SetCommandData(ctx, commandID, data)

	s.Logger(ctx).Info(fmt.Sprintf("signing mint command [%s] for pending transfers to chain %s", commandIDHex, chain.Name))
	signHash := types.GetSignHash(data)

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = s.signer.StartSign(ctx, s.voter, keyID, commandIDHex, signHash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	// TODO: Archive pending transfers after signing is completed
	for _, pendingTransfer := range pendingTransfers {
		s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)

	return &types.SignPendingTransfersResponse{CommandID: commandID[:]}, nil
}

func (s msgServer) SignTransferOwnership(c context.Context, req *types.SignTransferOwnershipRequest) (*types.SignTransferOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	chainID := s.getChainID(ctx, req.Chain)
	if chainID == nil {
		return nil, fmt.Errorf("Could not find chain ID for '%s'", req.Chain)
	}

	key, ok := s.signer.GetKey(ctx, req.KeyID)
	if !ok {
		return nil, fmt.Errorf("unkown key %s", req.KeyID)
	}

	next, nextAssigned := s.signer.GetNextKey(ctx, chain, tss.MasterKey)
	if nextAssigned {
		return nil, fmt.Errorf("key %s already assigned as the next %s key for chain %s", next.ID, tss.MasterKey.SimpleString(), chain.Name)
	}

	if err := s.signer.AssertMatchesRequirements(ctx, s.snapshotter, chain, key.ID, tss.MasterKey); err != nil {
		return nil, sdkerrors.Wrapf(err, "key %s does not match requirements for role %s", key.ID, tss.MasterKey.SimpleString())
	}

	newOwner := crypto.PubkeyToAddress(key.Value)

	commandID := getCommandID(newOwner.Bytes(), chainID)

	data, err := types.CreateTransferOwnershipCommandData(chainID, commandID, newOwner)
	if err != nil {
		return nil, err
	}

	keyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	keeper := s.ForChain(ctx, chain.Name)

	s.Logger(ctx).Info(fmt.Sprintf("storing data for transfer-ownership command %s", commandIDHex))
	keeper.SetCommandData(ctx, commandID, data)

	signHash := types.GetSignHash(data)

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = s.signer.StartSign(ctx, s.voter, keyID, commandIDHex, signHash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)

	return &types.SignTransferOwnershipResponse{CommandID: commandID[:]}, nil
}

func (s msgServer) AddChain(c context.Context, req *types.AddChainRequest) (*types.AddChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, found := s.nexus.GetChain(ctx, req.Name); found {
		return &types.AddChainResponse{}, fmt.Errorf("chain '%s' is already registered", req.Name)
	}

	s.SetPendingChain(ctx, nexus.Chain{Name: req.Name, NativeAsset: req.NativeAsset, SupportsForeignAssets: true})
	s.tss.SetKeyRequirement(ctx, req.KeyRequirement)
	s.SetParams(ctx, req.Params)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeNewChain,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueUpdate),
			sdk.NewAttribute(types.AttributeKeyChain, req.Name),
			sdk.NewAttribute(types.AttributeKeyNativeAsset, req.NativeAsset),
		),
	)

	return &types.AddChainResponse{}, nil
}

func (s msgServer) getChainID(ctx sdk.Context, chain string) (chainID *big.Int) {
	for _, p := range s.GetParams(ctx) {
		if strings.ToLower(p.Chain) == strings.ToLower(chain) {
			chainID = s.ForChain(ctx, chain).GetChainIDByNetwork(ctx, p.Network)
		}
	}

	return
}

func getCommandID(data []byte, chainID *big.Int) types.CommandID {
	concat := append(data, chainID.Bytes()...)
	hash := crypto.Keccak256(concat)[:32]

	var commandID types.CommandID
	copy(commandID[:], hash)

	return commandID
}
