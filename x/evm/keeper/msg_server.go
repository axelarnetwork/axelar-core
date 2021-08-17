package keeper

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
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
		return nil, fmt.Errorf("could not retrieve symbol for token %s", req.Asset)
	}

	s.nexus.LinkAddresses(ctx,
		nexus.CrossChainAddress{Chain: senderChain, Address: burnerAddr.String()},
		nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr})

	burnerInfo := types.BurnerInfo{
		TokenAddress:     types.Address(tokenAddr),
		DestinationChain: req.RecipientChain,
		Symbol:           symbol,
		Asset:            req.Asset,
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
		return nil, fmt.Errorf("could not retrieve revote locking period for chain %s", req.Chain)
	}

	if err := s.voter.InitializePoll(ctx, pollKey, seqNo, vote.ExpiryAt(ctx.BlockHeight()+period)); err != nil {
		return nil, err
	}

	symbol, ok := keeper.GetTokenSymbol(ctx, originChain.NativeAsset)
	if !ok {
		return nil, fmt.Errorf("could not retrieve symbol for token %s", originChain.NativeAsset)
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
		return nil, fmt.Errorf("could not retrieve revote locking period for chain %s", req.Name)
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
		return nil, fmt.Errorf("could not retrieve revote locking period for chain %s", req.Chain)
	}

	pollKey := vote.NewPollKey(types.ModuleName, req.TxID.Hex()+"_"+req.BurnerAddress.Hex())
	if err := s.voter.InitializePoll(ctx, pollKey, seqNo, vote.ExpiryAt(ctx.BlockHeight()+period)); err != nil {
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
			sdk.NewAttribute(types.AttributeKeyBurnAddress, req.BurnerAddress.Hex()),
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

	pk, ok := s.signer.GetKey(ctx, req.KeyID)
	if !ok {
		return nil, fmt.Errorf("key %s does not exist (yet)", req.KeyID)
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

	_, ok = s.signer.GetNextKey(ctx, chain, keyRole)
	if !ok {
		return nil, fmt.Errorf("next %s key for chain %s not set yet", keyRole.SimpleString(), chain.Name)
	}

	keeper := s.ForChain(ctx, chain.Name)

	gatewayAddr, ok := keeper.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	currentKeyID, ok := s.signer.GetCurrentKeyID(ctx, chain, keyRole)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	seqNo, ok := s.signer.GetSnapshotCounterForKeyID(ctx, currentKeyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot seqNo for key ID %s registered", currentKeyID)
	}

	period, ok := keeper.GetRevoteLockingPeriod(ctx)
	if !ok {
		return nil, fmt.Errorf("could not retrieve revote locking period for chain %s", req.Chain)
	}

	pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s_%s", req.TxID.Hex(), req.TransferType.SimpleString(), req.KeyID))
	if err := s.voter.InitializePoll(ctx, pollKey, seqNo, vote.ExpiryAt(ctx.BlockHeight()+period)); err != nil {
		return nil, err
	}

	transferKey := types.TransferKey{
		TxID:      req.TxID,
		Type:      req.TransferType,
		NextKeyID: pk.ID,
	}
	keeper.SetPendingTransferKey(ctx, pollKey, &transferKey)

	height, _ := keeper.GetRequiredConfirmationHeight(ctx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTransferKeyConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyTransferKeyType, req.TransferType.SimpleString()),
			sdk.NewAttribute(types.AttributeKeyGatewayAddress, gatewayAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeyAddress, crypto.PubkeyToAddress(pk.Value).Hex()),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
		),
	)

	return &types.ConfirmTransferKeyResponse{}, nil
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

	_, ok = s.nexus.GetChain(ctx, pendingDeposit.DestinationChain)
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
	amount := sdk.NewInt64Coin(pendingDeposit.Asset, pendingDeposit.Amount.BigInt().Int64())
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

// VoteConfirmTransferKey handles votes for transfer ownership/operatorship confirmations
func (s msgServer) VoteConfirmTransferKey(c context.Context, req *types.VoteConfirmTransferKeyRequest) (*types.VoteConfirmTransferKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	keeper := s.ForChain(ctx, chain.Name)

	pendingTransfer, pendingTransferFound := keeper.GetPendingTransferKey(ctx, req.PollKey)
	archivedTransfer, archivedTransferFound := keeper.GetArchivedTransferKey(ctx, req.PollKey)

	var nextKey tss.Key
	switch {
	case !pendingTransferFound && !archivedTransferFound:
		return nil, fmt.Errorf("no transfer ownership found for poll %s", req.PollKey.String())
	// If the voting threshold has been met and additional votes are received they should not return an error
	case archivedTransferFound:
		return &types.VoteConfirmTransferKeyResponse{Log: fmt.Sprintf("%s in %s to keyID %s already confirmed", archivedTransfer.Type.SimpleString(), archivedTransfer.TxID.Hex(), archivedTransfer.NextKeyID)}, nil
	case pendingTransferFound:
		nextKey, ok = s.signer.GetKey(ctx, pendingTransfer.NextKeyID)
		if !ok {
			return nil, fmt.Errorf("key %s cannot be found", pendingTransfer.NextKeyID)
		}

		if crypto.PubkeyToAddress(nextKey.Value) != common.Address(req.NewAddress) || pendingTransfer.Type != req.TransferType || pendingTransfer.TxID != req.TxID {
			return nil, fmt.Errorf("%s in %s to address %s does not match poll %s", pendingTransfer.Type.SimpleString(), req.TxID, req.NewAddress.Hex(), req.PollKey.String())
		}
	default:
		// assert: the transfer ownership/operatorship is known and has not been confirmed before
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
		return &types.VoteConfirmTransferKeyResponse{Log: fmt.Sprintf("not enough votes to confirm %s in %s to %s yet", req.TransferType.SimpleString(), req.TxID, req.NewAddress.Hex())}, nil
	}

	if poll.Is(vote.Failed) {
		keeper.DeletePendingTransferKey(ctx, req.PollKey)
		return &types.VoteConfirmTransferKeyResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	// TODO: handle rejected case

	s.Logger(ctx).Info(fmt.Sprintf("%s transfer ownership confirmation result is %s", chain.Name, poll.GetResult()))
	keeper.ArchiveTransferKey(ctx, req.PollKey)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeTransferKeyConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
		sdk.NewAttribute(types.AttributeKeyTransferKeyType, pendingTransfer.Type.SimpleString()),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))))

	if !confirmed.Value {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))

		return &types.VoteConfirmTransferKeyResponse{
			Log: fmt.Sprintf("transfer ownership in %s to %s was discarded", req.TxID, req.NewAddress.Hex()),
		}, nil
	}

	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	if err := s.signer.RotateKey(ctx, chain, nextKey.Role); err != nil {
		return nil, err
	}

	return &types.VoteConfirmTransferKeyResponse{}, nil
}

func (s msgServer) CreateDeployToken(c context.Context, req *types.CreateDeployTokenRequest) (*types.CreateDeployTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keeper := s.ForChain(ctx, req.Chain)

	if _, ok := keeper.GetGatewayAddress(ctx); !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	chainID := s.getChainID(ctx, req.Chain)
	if chainID == nil {
		return nil, fmt.Errorf("could not find chain ID for '%s'", req.Chain)
	}

	originChain, found := s.nexus.GetChain(ctx, req.OriginChain)
	if !found {
		return nil, fmt.Errorf("%s is not a registered chain", req.OriginChain)
	}

	if _, nextMasterKeyAssigned := s.signer.GetNextKey(ctx, chain, tss.MasterKey); nextMasterKeyAssigned {
		return nil, fmt.Errorf("next %s key already assigned for chain %s, rotate key first", tss.MasterKey.SimpleString(), chain.Name)
	}

	masterKeyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
	}

	command, err := types.CreateDeployTokenCommand(
		chainID,
		masterKeyID,
		req.TokenName,
		req.Symbol,
		req.Decimals,
		req.Capacity.BigInt(),
	)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed create deploy-token command token %s(%s) for chain %s", req.TokenName, req.Symbol, chain.Name)
	}

	keeper.SetTokenInfo(ctx, originChain.NativeAsset, req)
	if err := keeper.SetCommand(ctx, command); err != nil {
		return nil, err
	}

	return &types.CreateDeployTokenResponse{}, nil
}

func (s msgServer) CreateBurnTokens(c context.Context, req *types.CreateBurnTokensRequest) (*types.CreateBurnTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keeper := s.ForChain(ctx, req.Chain)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	deposits := keeper.GetConfirmedDeposits(ctx)
	if len(deposits) == 0 {
		return &types.CreateBurnTokensResponse{}, nil
	}

	chainID := s.getChainID(ctx, req.Chain)
	if chainID == nil {
		return nil, fmt.Errorf("could not find chain ID for '%s'", req.Chain)
	}

	if _, nextSecondaryKeyAssigned := s.signer.GetNextKey(ctx, chain, tss.SecondaryKey); nextSecondaryKeyAssigned {
		return nil, fmt.Errorf("next %s key already assigned for chain %s, rotate key first", tss.SecondaryKey.SimpleString(), chain.Name)
	}

	secondaryKeyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("no %s key for chain %s found", tss.SecondaryKey.SimpleString(), chain.Name)
	}

	seen := map[string]bool{}
	for _, deposit := range deposits {
		if seen[deposit.BurnerAddress.Hex()] {
			continue
		}

		burnerInfo := keeper.GetBurnerInfo(ctx, common.Address(deposit.BurnerAddress))
		if burnerInfo == nil {
			return nil, fmt.Errorf("no burner info found for address %s", deposit.BurnerAddress.Hex())
		}

		command, err := types.CreateBurnTokenCommand(chainID, secondaryKeyID, ctx.BlockHeight(), *burnerInfo)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to create burn-token command to burn token at address %s for chain %s", deposit.BurnerAddress.Hex(), chain.Name)
		}

		if err := keeper.SetCommand(ctx, command); err != nil {
			return nil, err
		}
	}

	return &types.CreateBurnTokensResponse{}, nil
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

	if _, err := s.signer.ScheduleSign(ctx, tss.SignInfo{
		KeyID:           keyID,
		SigID:           txID,
		Msg:             hash.Bytes(),
		SnapshotCounter: snapshot.Counter,
	}); err != nil {
		return nil, err
	}

	byteCode, ok := keeper.GetGatewayByteCodes(ctx)
	if !ok {
		return nil, fmt.Errorf("could not retrieve gateway bytecodes for chain %s", req.Chain)
	}

	secondaryKey, ok := s.signer.GetCurrentKey(ctx, chain, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("no %s key for chain %s found", tss.SecondaryKey.SimpleString(), chain.Name)
	}

	deploymentBytecode, err := types.GetGatewayDeploymentBytecode(byteCode, crypto.PubkeyToAddress(secondaryKey.Value))
	if err != nil {
		return nil, err
	}

	// if this is the transaction that is deploying Axelar Gateway, calculate and save address
	// TODO: this is something that should be done after the signature has been successfully confirmed
	if tx.To() == nil && bytes.Equal(tx.Data(), deploymentBytecode) {
		pub, ok := s.signer.GetCurrentKey(ctx, chain, tss.MasterKey)
		if !ok {
			return nil, fmt.Errorf("no master key for chain %s found", chain.Name)
		}

		addr := crypto.CreateAddress(crypto.PubkeyToAddress(pub.Value), tx.Nonce())
		keeper.SetGatewayAddress(ctx, addr)

		telemetry.NewLabel("eth_factory_addr", addr.String())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyTxID, txID),
		),
	)

	return &types.SignTxResponse{TxID: txID}, nil
}

func transferIDtoCommandID(transferID uint64) types.CommandID {
	var commandID types.CommandID

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, transferID)

	copy(commandID[:], common.LeftPadBytes(bz, 32)[:32])

	return commandID
}

func (s msgServer) CreatePendingTransfers(c context.Context, req *types.CreatePendingTransfersRequest) (*types.CreatePendingTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keeper := s.ForChain(ctx, req.Chain)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	pendingTransfers := s.nexus.GetTransfersForChain(ctx, chain, nexus.Pending)
	if len(pendingTransfers) == 0 {
		return &types.CreatePendingTransfersResponse{}, nil
	}

	chainID := s.getChainID(ctx, req.Chain)
	if chainID == nil {
		return nil, fmt.Errorf("could not find chain ID for '%s'", req.Chain)
	}

	if _, nextSecondaryKeyAssigned := s.signer.GetNextKey(ctx, chain, tss.SecondaryKey); nextSecondaryKeyAssigned {
		return nil, fmt.Errorf("next %s key already assigned for chain %s, rotate key first", tss.SecondaryKey.SimpleString(), chain.Name)
	}

	secondaryKeyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("no %s key for chain %s found", tss.SecondaryKey.SimpleString(), chain.Name)
	}

	getRecipientAndAsset := func(transfer nexus.CrossChainTransfer) string {
		return fmt.Sprintf("%s-%s", transfer.Recipient.Address, transfer.Asset.Denom)
	}
	transfers := nexus.MergeTransfersBy(pendingTransfers, getRecipientAndAsset)

	for _, transfer := range transfers {
		symbol, found := keeper.GetTokenSymbol(ctx, transfer.Asset.Denom)
		if !found {
			return nil, fmt.Errorf("could not find symbol for asset %s", transfer.Asset.Denom)
		}

		command, err := types.CreateMintTokenCommand(
			chainID,
			secondaryKeyID,
			transferIDtoCommandID(transfer.ID),
			symbol,
			common.HexToAddress(transfer.Recipient.Address),
			transfer.Asset.Amount.BigInt(),
		)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed create mint-token command for transfer %d", transfer.ID)
		}

		s.Logger(ctx).Info(fmt.Sprintf("storing data for mint command %s", command.ID.Hex()))

		if err := keeper.SetCommand(ctx, command); err != nil {
			return nil, err
		}
	}

	for _, pendingTransfer := range pendingTransfers {
		s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
	}

	return &types.CreatePendingTransfersResponse{}, nil
}

func (s msgServer) createTransferKeyCommand(ctx sdk.Context, transferKeyType types.TransferKeyType, chainStr string, nextKeyID string) (types.Command, error) {
	chain, ok := s.nexus.GetChain(ctx, chainStr)
	if !ok {
		return types.Command{}, fmt.Errorf("%s is not a registered chain", chainStr)
	}

	chainID := s.getChainID(ctx, chainStr)
	if chainID == nil {
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
	if _, nextMasterKeyAssigned := s.signer.GetNextKey(ctx, chain, tss.MasterKey); nextMasterKeyAssigned {
		return types.Command{}, fmt.Errorf("next %s key already assigned for chain %s, rotate key first", tss.MasterKey.SimpleString(), chain.Name)
	}
	if _, nextSecondaryKeyAssigned := s.signer.GetNextKey(ctx, chain, tss.SecondaryKey); nextSecondaryKeyAssigned {
		return types.Command{}, fmt.Errorf("next %s key already assigned for chain %s, rotate key first", tss.SecondaryKey.SimpleString(), chain.Name)
	}

	nextKey, ok := s.signer.GetKey(ctx, nextKeyID)
	if !ok {
		return types.Command{}, fmt.Errorf("unkown key %s", nextKeyID)
	}

	if err := s.signer.AssertMatchesRequirements(ctx, s.snapshotter, chain, nextKey.ID, keyRole); err != nil {
		return types.Command{}, sdkerrors.Wrapf(err, "key %s does not match requirements for role %s", nextKey.ID, keyRole.SimpleString())
	}

	newAddress := crypto.PubkeyToAddress(nextKey.Value)
	currMasterKeyID, ok := s.signer.GetCurrentKeyID(ctx, chain, tss.MasterKey)
	if !ok {
		return types.Command{}, fmt.Errorf("current %s key not set for chain %s", tss.MasterKey, chain.Name)
	}

	var command types.Command
	var err error

	switch transferKeyType {
	case types.Ownership:
		command, err = types.CreateTransferOwnershipCommand(chainID, currMasterKeyID, newAddress)
	case types.Operatorship:
		command, err = types.CreateTransferOperatorshipCommand(chainID, currMasterKeyID, newAddress)
	default:
		return types.Command{}, fmt.Errorf("invalid transfer key type %s", transferKeyType.SimpleString())
	}

	if err != nil {
		return types.Command{}, sdkerrors.Wrapf(err, "failed create %s command", transferKeyType.SimpleString())
	}

	s.Logger(ctx).Info(fmt.Sprintf("storing data for %s command %s", transferKeyType.SimpleString(), command.ID.Hex()))

	if err := s.signer.AssignNextKey(ctx, chain, keyRole, nextKey.ID); err != nil {
		return types.Command{}, sdkerrors.Wrapf(err, "failed assigning the next %s key for chain %s", keyRole.SimpleString(), chain.Name)
	}

	s.Logger(ctx).Debug(fmt.Sprintf("created command %s for chain %s to transfer to address %s", transferKeyType.SimpleString(), chain.Name, newAddress.Hex()))

	return command, nil
}

func (s msgServer) CreateTransferOwnership(c context.Context, req *types.CreateTransferOwnershipRequest) (*types.CreateTransferOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keeper := s.ForChain(ctx, req.Chain)

	if _, ok := keeper.GetGatewayAddress(ctx); !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	command, err := s.createTransferKeyCommand(ctx, types.Ownership, req.Chain, req.KeyID)
	if err != nil {
		return nil, err
	}

	if err := keeper.SetCommand(ctx, command); err != nil {
		return nil, err
	}

	return &types.CreateTransferOwnershipResponse{}, nil
}

func (s msgServer) CreateTransferOperatorship(c context.Context, req *types.CreateTransferOperatorshipRequest) (*types.CreateTransferOperatorshipResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keeper := s.ForChain(ctx, req.Chain)

	if _, ok := keeper.GetGatewayAddress(ctx); !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	command, err := s.createTransferKeyCommand(ctx, types.Operatorship, req.Chain, req.KeyID)
	if err != nil {
		return nil, err
	}

	if err := keeper.SetCommand(ctx, command); err != nil {
		return nil, err
	}

	return &types.CreateTransferOperatorshipResponse{}, nil
}

func (s msgServer) SignCommands(c context.Context, req *types.SignCommandsRequest) (*types.SignCommandsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.Chain)
	}

	chainID := s.getChainID(ctx, req.Chain)
	if chainID == nil {
		return nil, fmt.Errorf("could not find chain ID for '%s'", req.Chain)
	}

	keeper := s.ForChain(ctx, chain.Name)
	batchedCommands, err := getBatchedCommandsToSign(ctx, keeper, chainID)
	if err != nil {
		return nil, err
	}

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, batchedCommands.KeyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", batchedCommands.KeyID)
	}

	batchedCommandsIDHex := hex.EncodeToString(batchedCommands.ID)
	if _, err := s.signer.ScheduleSign(ctx, tss.SignInfo{
		KeyID:           batchedCommands.KeyID,
		SigID:           batchedCommandsIDHex,
		Msg:             batchedCommands.SigHash.Bytes(),
		SnapshotCounter: counter,
	}); err != nil {
		return nil, err
	}

	keeper.SetUnsignedBatchedCommands(ctx, batchedCommands)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyBatchedCommandsID, batchedCommandsIDHex),
		),
	)

	return &types.SignCommandsResponse{BatchedCommandsID: batchedCommands.ID}, nil
}

func getBatchedCommandsToSign(ctx sdk.Context, keeper types.ChainKeeper, chainID *big.Int) (types.BatchedCommands, error) {
	if unsignedBatchedCommands, ok := keeper.GetUnsignedBatchedCommands(ctx); ok {
		if unsignedBatchedCommands.Is(types.Aborted) {
			return unsignedBatchedCommands, nil
		}

		return types.BatchedCommands{}, fmt.Errorf("signing for batched commands %s is still in progress", hex.EncodeToString(unsignedBatchedCommands.ID))
	}

	var command types.Command
	commandQueue := keeper.GetCommandQueue(ctx)

	if !commandQueue.Dequeue(&command) {
		return types.BatchedCommands{}, fmt.Errorf("no commands are found to sign for chain %s", keeper.GetName())
	}
	// Only batching commands to be signed by the same key
	keyID := command.KeyID
	filter := func(value codec.ProtoMarshaler) bool {
		cmd, ok := value.(*types.Command)

		return ok && cmd.KeyID == keyID
	}

	commands := []types.Command{command.Clone()}
	// TODO: limit the number of commands that are signed each time to avoid going above the gas limit
	for commandQueue.Dequeue(&command, filter) {
		commands = append(commands, command.Clone())
	}

	return types.NewBatchedCommands(chainID, keyID, commands)
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
		if strings.EqualFold(p.Chain, chain) {
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
