package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns a new msg server instance
func NewMsgServerImpl(keeper Keeper) types.MsgServiceServer {
	return msgServer{
		Keeper: keeper,
	}
}

// Vote handles vote request
func (s msgServer) Vote(c context.Context, req *types.VoteRequest) (*types.VoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	voteHandler := s.GetVoteRouter().GetHandler(req.PollKey.Module)
	if voteHandler == nil {
		return nil, fmt.Errorf("unknown module for vote %s", req.PollKey.Module)
	}

	poll := s.GetPoll(ctx, req.PollKey)

	switch {
	case poll.Is(vote.Expired):
		return &types.VoteResponse{Log: fmt.Sprintf("vote for poll %s already %s", req.PollKey, vote.Expired.String())}, nil
	case poll.Is(vote.Failed), poll.Is(vote.Completed):
		// If the voting threshold has been met and additional votes are received they should not return an error
		return &types.VoteResponse{Log: fmt.Sprintf("vote for poll %s already %s", req.PollKey, vote.Completed.String())}, nil
	default:
	}

	if err := poll.Vote(voter, &req.Vote); err != nil {
		return nil, err
	}

	event := sdk.NewEvent(types.EventType,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueVote),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))),
		sdk.NewAttribute(types.AttributeKeyVoter, req.Sender.String()),
	)
	defer func() { ctx.EventManager().EmitEvent(event) }()

	if poll.Is(vote.Pending) {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyPollState, vote.Pending.String()))
		return &types.VoteResponse{Log: fmt.Sprintf("not enough votes to confirm poll %s yet", poll.GetKey())}, nil
	}

	if poll.Is(vote.Failed) {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyPollState, vote.Failed.String()))
		return &types.VoteResponse{Log: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	_, ok := poll.GetResult().(*vote.Vote)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected VoteConfirmDepositRequest_Vote, got %T", poll.GetKey().String(), poll.GetResult())
	}

	if err := voteHandler(ctx, poll); err != nil {
		return &types.VoteResponse{Log: fmt.Sprintf("vote handler failed %s", err.Error())}, nil
	}

	return &types.VoteResponse{}, nil
}
