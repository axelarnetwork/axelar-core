package keeper

import (
	"context"
	"fmt"

	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	poll, ok := s.GetPoll(ctx, req.PollID)
	if !ok {
		return nil, fmt.Errorf("poll %s not found", req.PollID)
	}

	voteResult, err := poll.Vote(voter, ctx.BlockHeight(), req.Vote)
	if err != nil {
		return nil, err
	}

	event := sdk.NewEvent(types.EventType,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueVote),
		sdk.NewAttribute(types.AttributeKeyPoll, req.PollID.String()),
		sdk.NewAttribute(types.AttributeKeyVoter, req.Sender.String()),
		sdk.NewAttribute(types.AttributeKeyPollState, poll.GetState().String()),
	)

	if voteResult != vote.NoVote {
		defer func() { ctx.EventManager().EmitEvent(event) }()
	}

	switch poll.GetState() {
	case vote.Pending:
		return &types.VoteResponse{Log: fmt.Sprintf("not enough votes to confirm poll %s yet", req.PollID.String())}, nil
	case vote.Failed:
		return &types.VoteResponse{Log: fmt.Sprintf("poll %s failed", req.PollID.String())}, nil
	case vote.Completed:

		if voteResult == vote.VoteInTime {
			voteHandler := s.GetVoteRouter().GetHandler(poll.GetModule())
			if err := voteHandler.HandleResult(ctx, poll.GetResult()); err != nil {
				return &types.VoteResponse{Log: fmt.Sprintf("vote handler failed %s", err.Error())}, nil
			}
		}

		return &types.VoteResponse{}, nil
	default:
		panic(fmt.Sprintf("unexpected poll state %s", poll.GetState().String()))
	}
}
