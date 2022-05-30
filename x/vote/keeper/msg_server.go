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

	poll := s.GetPoll(ctx, req.PollID)
	result, voted, err := poll.Vote(voter, ctx.BlockHeight(), &req.Vote)
	if err != nil {
		return nil, err
	}

	event := sdk.NewEvent(types.EventType,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueVote),
		sdk.NewAttribute(types.AttributeKeyPoll, req.PollID.String()),
		sdk.NewAttribute(types.AttributeKeyVoter, req.Sender.String()),
	)

	if voted {
		defer func() { ctx.EventManager().EmitEvent(event) }()
	}

	switch {
	case poll.Is(vote.Pending):
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyPollState, vote.Pending.String()))

		return &types.VoteResponse{Log: fmt.Sprintf("not enough votes to confirm poll %s yet", poll.GetID().String())}, nil
	case poll.Is(vote.Failed):
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyPollState, vote.Failed.String()))

		return &types.VoteResponse{Log: fmt.Sprintf("poll %s failed", poll.GetID().String())}, nil
	case poll.Is(vote.Expired):
		return &types.VoteResponse{Log: fmt.Sprintf("poll %s expired", poll.GetID().String())}, nil
	case result != nil:
		_, ok := result.(*vote.Vote)
		if !ok {
			return nil, fmt.Errorf("result of poll %s has wrong type, expected *exported.Vote, got %T", poll.GetID().String(), poll.GetResult())
		}

		pollModuleMetadata := poll.GetModuleMetadata()
		voteHandler := s.GetVoteRouter().GetHandler(pollModuleMetadata.Module)
		if voteHandler == nil {
			return nil, fmt.Errorf("unknown module for vote %s", pollModuleMetadata.Module)
		}

		if err := voteHandler.HandleResult(ctx, result); err != nil {
			return nil, err
		}

		fallthrough
	default:
		return &types.VoteResponse{}, nil
	}
}
