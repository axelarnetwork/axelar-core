// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"sync"
)

// Ensure, that PollMock does implement exported.Poll.
// If this is not the case, regenerate this file with moq.
var _ exported.Poll = &PollMock{}

// PollMock is a mock implementation of exported.Poll.
//
// 	func TestSomethingThatUsesPoll(t *testing.T) {
//
// 		// make and configure a mocked exported.Poll
// 		mockedPoll := &PollMock{
// 			GetIDFunc: func() exported.PollID {
// 				panic("mock out the GetID method")
// 			},
// 			GetModuleFunc: func() string {
// 				panic("mock out the GetModule method")
// 			},
// 			GetResultFunc: func() codec.ProtoMarshaler {
// 				panic("mock out the GetResult method")
// 			},
// 			GetRewardPoolNameFunc: func() (string, bool) {
// 				panic("mock out the GetRewardPoolName method")
// 			},
// 			GetStateFunc: func() exported.PollState {
// 				panic("mock out the GetState method")
// 			},
// 			GetVotersFunc: func() []sdk.ValAddress {
// 				panic("mock out the GetVoters method")
// 			},
// 			HasVotedFunc: func(voter sdk.ValAddress) bool {
// 				panic("mock out the HasVoted method")
// 			},
// 			HasVotedCorrectlyFunc: func(voter sdk.ValAddress) bool {
// 				panic("mock out the HasVotedCorrectly method")
// 			},
// 			VoteFunc: func(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) (exported.VoteResult, error) {
// 				panic("mock out the Vote method")
// 			},
// 		}
//
// 		// use mockedPoll in code that requires exported.Poll
// 		// and then make assertions.
//
// 	}
type PollMock struct {
	// GetIDFunc mocks the GetID method.
	GetIDFunc func() exported.PollID

	// GetModuleFunc mocks the GetModule method.
	GetModuleFunc func() string

	// GetResultFunc mocks the GetResult method.
	GetResultFunc func() codec.ProtoMarshaler

	// GetRewardPoolNameFunc mocks the GetRewardPoolName method.
	GetRewardPoolNameFunc func() (string, bool)

	// GetStateFunc mocks the GetState method.
	GetStateFunc func() exported.PollState

	// GetVotersFunc mocks the GetVoters method.
	GetVotersFunc func() []sdk.ValAddress

	// HasVotedFunc mocks the HasVoted method.
	HasVotedFunc func(voter sdk.ValAddress) bool

	// HasVotedCorrectlyFunc mocks the HasVotedCorrectly method.
	HasVotedCorrectlyFunc func(voter sdk.ValAddress) bool

	// VoteFunc mocks the Vote method.
	VoteFunc func(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) (exported.VoteResult, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetID holds details about calls to the GetID method.
		GetID []struct {
		}
		// GetModule holds details about calls to the GetModule method.
		GetModule []struct {
		}
		// GetResult holds details about calls to the GetResult method.
		GetResult []struct {
		}
		// GetRewardPoolName holds details about calls to the GetRewardPoolName method.
		GetRewardPoolName []struct {
		}
		// GetState holds details about calls to the GetState method.
		GetState []struct {
		}
		// GetVoters holds details about calls to the GetVoters method.
		GetVoters []struct {
		}
		// HasVoted holds details about calls to the HasVoted method.
		HasVoted []struct {
			// Voter is the voter argument value.
			Voter sdk.ValAddress
		}
		// HasVotedCorrectly holds details about calls to the HasVotedCorrectly method.
		HasVotedCorrectly []struct {
			// Voter is the voter argument value.
			Voter sdk.ValAddress
		}
		// Vote holds details about calls to the Vote method.
		Vote []struct {
			// Voter is the voter argument value.
			Voter sdk.ValAddress
			// BlockHeight is the blockHeight argument value.
			BlockHeight int64
			// Data is the data argument value.
			Data codec.ProtoMarshaler
		}
	}
	lockGetID             sync.RWMutex
	lockGetModule         sync.RWMutex
	lockGetResult         sync.RWMutex
	lockGetRewardPoolName sync.RWMutex
	lockGetState          sync.RWMutex
	lockGetVoters         sync.RWMutex
	lockHasVoted          sync.RWMutex
	lockHasVotedCorrectly sync.RWMutex
	lockVote              sync.RWMutex
}

// GetID calls GetIDFunc.
func (mock *PollMock) GetID() exported.PollID {
	if mock.GetIDFunc == nil {
		panic("PollMock.GetIDFunc: method is nil but Poll.GetID was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetID.Lock()
	mock.calls.GetID = append(mock.calls.GetID, callInfo)
	mock.lockGetID.Unlock()
	return mock.GetIDFunc()
}

// GetIDCalls gets all the calls that were made to GetID.
// Check the length with:
//     len(mockedPoll.GetIDCalls())
func (mock *PollMock) GetIDCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetID.RLock()
	calls = mock.calls.GetID
	mock.lockGetID.RUnlock()
	return calls
}

// GetModule calls GetModuleFunc.
func (mock *PollMock) GetModule() string {
	if mock.GetModuleFunc == nil {
		panic("PollMock.GetModuleFunc: method is nil but Poll.GetModule was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetModule.Lock()
	mock.calls.GetModule = append(mock.calls.GetModule, callInfo)
	mock.lockGetModule.Unlock()
	return mock.GetModuleFunc()
}

// GetModuleCalls gets all the calls that were made to GetModule.
// Check the length with:
//     len(mockedPoll.GetModuleCalls())
func (mock *PollMock) GetModuleCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetModule.RLock()
	calls = mock.calls.GetModule
	mock.lockGetModule.RUnlock()
	return calls
}

// GetResult calls GetResultFunc.
func (mock *PollMock) GetResult() codec.ProtoMarshaler {
	if mock.GetResultFunc == nil {
		panic("PollMock.GetResultFunc: method is nil but Poll.GetResult was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetResult.Lock()
	mock.calls.GetResult = append(mock.calls.GetResult, callInfo)
	mock.lockGetResult.Unlock()
	return mock.GetResultFunc()
}

// GetResultCalls gets all the calls that were made to GetResult.
// Check the length with:
//     len(mockedPoll.GetResultCalls())
func (mock *PollMock) GetResultCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetResult.RLock()
	calls = mock.calls.GetResult
	mock.lockGetResult.RUnlock()
	return calls
}

// GetRewardPoolName calls GetRewardPoolNameFunc.
func (mock *PollMock) GetRewardPoolName() (string, bool) {
	if mock.GetRewardPoolNameFunc == nil {
		panic("PollMock.GetRewardPoolNameFunc: method is nil but Poll.GetRewardPoolName was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetRewardPoolName.Lock()
	mock.calls.GetRewardPoolName = append(mock.calls.GetRewardPoolName, callInfo)
	mock.lockGetRewardPoolName.Unlock()
	return mock.GetRewardPoolNameFunc()
}

// GetRewardPoolNameCalls gets all the calls that were made to GetRewardPoolName.
// Check the length with:
//     len(mockedPoll.GetRewardPoolNameCalls())
func (mock *PollMock) GetRewardPoolNameCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetRewardPoolName.RLock()
	calls = mock.calls.GetRewardPoolName
	mock.lockGetRewardPoolName.RUnlock()
	return calls
}

// GetState calls GetStateFunc.
func (mock *PollMock) GetState() exported.PollState {
	if mock.GetStateFunc == nil {
		panic("PollMock.GetStateFunc: method is nil but Poll.GetState was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetState.Lock()
	mock.calls.GetState = append(mock.calls.GetState, callInfo)
	mock.lockGetState.Unlock()
	return mock.GetStateFunc()
}

// GetStateCalls gets all the calls that were made to GetState.
// Check the length with:
//     len(mockedPoll.GetStateCalls())
func (mock *PollMock) GetStateCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetState.RLock()
	calls = mock.calls.GetState
	mock.lockGetState.RUnlock()
	return calls
}

// GetVoters calls GetVotersFunc.
func (mock *PollMock) GetVoters() []sdk.ValAddress {
	if mock.GetVotersFunc == nil {
		panic("PollMock.GetVotersFunc: method is nil but Poll.GetVoters was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetVoters.Lock()
	mock.calls.GetVoters = append(mock.calls.GetVoters, callInfo)
	mock.lockGetVoters.Unlock()
	return mock.GetVotersFunc()
}

// GetVotersCalls gets all the calls that were made to GetVoters.
// Check the length with:
//     len(mockedPoll.GetVotersCalls())
func (mock *PollMock) GetVotersCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetVoters.RLock()
	calls = mock.calls.GetVoters
	mock.lockGetVoters.RUnlock()
	return calls
}

// HasVoted calls HasVotedFunc.
func (mock *PollMock) HasVoted(voter sdk.ValAddress) bool {
	if mock.HasVotedFunc == nil {
		panic("PollMock.HasVotedFunc: method is nil but Poll.HasVoted was just called")
	}
	callInfo := struct {
		Voter sdk.ValAddress
	}{
		Voter: voter,
	}
	mock.lockHasVoted.Lock()
	mock.calls.HasVoted = append(mock.calls.HasVoted, callInfo)
	mock.lockHasVoted.Unlock()
	return mock.HasVotedFunc(voter)
}

// HasVotedCalls gets all the calls that were made to HasVoted.
// Check the length with:
//     len(mockedPoll.HasVotedCalls())
func (mock *PollMock) HasVotedCalls() []struct {
	Voter sdk.ValAddress
} {
	var calls []struct {
		Voter sdk.ValAddress
	}
	mock.lockHasVoted.RLock()
	calls = mock.calls.HasVoted
	mock.lockHasVoted.RUnlock()
	return calls
}

// HasVotedCorrectly calls HasVotedCorrectlyFunc.
func (mock *PollMock) HasVotedCorrectly(voter sdk.ValAddress) bool {
	if mock.HasVotedCorrectlyFunc == nil {
		panic("PollMock.HasVotedCorrectlyFunc: method is nil but Poll.HasVotedCorrectly was just called")
	}
	callInfo := struct {
		Voter sdk.ValAddress
	}{
		Voter: voter,
	}
	mock.lockHasVotedCorrectly.Lock()
	mock.calls.HasVotedCorrectly = append(mock.calls.HasVotedCorrectly, callInfo)
	mock.lockHasVotedCorrectly.Unlock()
	return mock.HasVotedCorrectlyFunc(voter)
}

// HasVotedCorrectlyCalls gets all the calls that were made to HasVotedCorrectly.
// Check the length with:
//     len(mockedPoll.HasVotedCorrectlyCalls())
func (mock *PollMock) HasVotedCorrectlyCalls() []struct {
	Voter sdk.ValAddress
} {
	var calls []struct {
		Voter sdk.ValAddress
	}
	mock.lockHasVotedCorrectly.RLock()
	calls = mock.calls.HasVotedCorrectly
	mock.lockHasVotedCorrectly.RUnlock()
	return calls
}

// Vote calls VoteFunc.
func (mock *PollMock) Vote(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) (exported.VoteResult, error) {
	if mock.VoteFunc == nil {
		panic("PollMock.VoteFunc: method is nil but Poll.Vote was just called")
	}
	callInfo := struct {
		Voter       sdk.ValAddress
		BlockHeight int64
		Data        codec.ProtoMarshaler
	}{
		Voter:       voter,
		BlockHeight: blockHeight,
		Data:        data,
	}
	mock.lockVote.Lock()
	mock.calls.Vote = append(mock.calls.Vote, callInfo)
	mock.lockVote.Unlock()
	return mock.VoteFunc(voter, blockHeight, data)
}

// VoteCalls gets all the calls that were made to Vote.
// Check the length with:
//     len(mockedPoll.VoteCalls())
func (mock *PollMock) VoteCalls() []struct {
	Voter       sdk.ValAddress
	BlockHeight int64
	Data        codec.ProtoMarshaler
} {
	var calls []struct {
		Voter       sdk.ValAddress
		BlockHeight int64
		Data        codec.ProtoMarshaler
	}
	mock.lockVote.RLock()
	calls = mock.calls.Vote
	mock.lockVote.RUnlock()
	return calls
}

// Ensure, that VoteHandlerMock does implement exported.VoteHandler.
// If this is not the case, regenerate this file with moq.
var _ exported.VoteHandler = &VoteHandlerMock{}

// VoteHandlerMock is a mock implementation of exported.VoteHandler.
//
// 	func TestSomethingThatUsesVoteHandler(t *testing.T) {
//
// 		// make and configure a mocked exported.VoteHandler
// 		mockedVoteHandler := &VoteHandlerMock{
// 			HandleCompletedPollFunc: func(ctx sdk.Context, poll exported.Poll) error {
// 				panic("mock out the HandleCompletedPoll method")
// 			},
// 			HandleExpiredPollFunc: func(ctx sdk.Context, poll exported.Poll) error {
// 				panic("mock out the HandleExpiredPoll method")
// 			},
// 			HandleResultFunc: func(ctx sdk.Context, result codec.ProtoMarshaler) error {
// 				panic("mock out the HandleResult method")
// 			},
// 			IsFalsyResultFunc: func(result codec.ProtoMarshaler) bool {
// 				panic("mock out the IsFalsyResult method")
// 			},
// 		}
//
// 		// use mockedVoteHandler in code that requires exported.VoteHandler
// 		// and then make assertions.
//
// 	}
type VoteHandlerMock struct {
	// HandleCompletedPollFunc mocks the HandleCompletedPoll method.
	HandleCompletedPollFunc func(ctx sdk.Context, poll exported.Poll) error

	// HandleExpiredPollFunc mocks the HandleExpiredPoll method.
	HandleExpiredPollFunc func(ctx sdk.Context, poll exported.Poll) error

	// HandleResultFunc mocks the HandleResult method.
	HandleResultFunc func(ctx sdk.Context, result codec.ProtoMarshaler) error

	// IsFalsyResultFunc mocks the IsFalsyResult method.
	IsFalsyResultFunc func(result codec.ProtoMarshaler) bool

	// calls tracks calls to the methods.
	calls struct {
		// HandleCompletedPoll holds details about calls to the HandleCompletedPoll method.
		HandleCompletedPoll []struct {
			// Ctx is the ctx argument value.
			Ctx sdk.Context
			// Poll is the poll argument value.
			Poll exported.Poll
		}
		// HandleExpiredPoll holds details about calls to the HandleExpiredPoll method.
		HandleExpiredPoll []struct {
			// Ctx is the ctx argument value.
			Ctx sdk.Context
			// Poll is the poll argument value.
			Poll exported.Poll
		}
		// HandleResult holds details about calls to the HandleResult method.
		HandleResult []struct {
			// Ctx is the ctx argument value.
			Ctx sdk.Context
			// Result is the result argument value.
			Result codec.ProtoMarshaler
		}
		// IsFalsyResult holds details about calls to the IsFalsyResult method.
		IsFalsyResult []struct {
			// Result is the result argument value.
			Result codec.ProtoMarshaler
		}
	}
	lockHandleCompletedPoll sync.RWMutex
	lockHandleExpiredPoll   sync.RWMutex
	lockHandleResult        sync.RWMutex
	lockIsFalsyResult       sync.RWMutex
}

// HandleCompletedPoll calls HandleCompletedPollFunc.
func (mock *VoteHandlerMock) HandleCompletedPoll(ctx sdk.Context, poll exported.Poll) error {
	if mock.HandleCompletedPollFunc == nil {
		panic("VoteHandlerMock.HandleCompletedPollFunc: method is nil but VoteHandler.HandleCompletedPoll was just called")
	}
	callInfo := struct {
		Ctx  sdk.Context
		Poll exported.Poll
	}{
		Ctx:  ctx,
		Poll: poll,
	}
	mock.lockHandleCompletedPoll.Lock()
	mock.calls.HandleCompletedPoll = append(mock.calls.HandleCompletedPoll, callInfo)
	mock.lockHandleCompletedPoll.Unlock()
	return mock.HandleCompletedPollFunc(ctx, poll)
}

// HandleCompletedPollCalls gets all the calls that were made to HandleCompletedPoll.
// Check the length with:
//     len(mockedVoteHandler.HandleCompletedPollCalls())
func (mock *VoteHandlerMock) HandleCompletedPollCalls() []struct {
	Ctx  sdk.Context
	Poll exported.Poll
} {
	var calls []struct {
		Ctx  sdk.Context
		Poll exported.Poll
	}
	mock.lockHandleCompletedPoll.RLock()
	calls = mock.calls.HandleCompletedPoll
	mock.lockHandleCompletedPoll.RUnlock()
	return calls
}

// HandleExpiredPoll calls HandleExpiredPollFunc.
func (mock *VoteHandlerMock) HandleExpiredPoll(ctx sdk.Context, poll exported.Poll) error {
	if mock.HandleExpiredPollFunc == nil {
		panic("VoteHandlerMock.HandleExpiredPollFunc: method is nil but VoteHandler.HandleExpiredPoll was just called")
	}
	callInfo := struct {
		Ctx  sdk.Context
		Poll exported.Poll
	}{
		Ctx:  ctx,
		Poll: poll,
	}
	mock.lockHandleExpiredPoll.Lock()
	mock.calls.HandleExpiredPoll = append(mock.calls.HandleExpiredPoll, callInfo)
	mock.lockHandleExpiredPoll.Unlock()
	return mock.HandleExpiredPollFunc(ctx, poll)
}

// HandleExpiredPollCalls gets all the calls that were made to HandleExpiredPoll.
// Check the length with:
//     len(mockedVoteHandler.HandleExpiredPollCalls())
func (mock *VoteHandlerMock) HandleExpiredPollCalls() []struct {
	Ctx  sdk.Context
	Poll exported.Poll
} {
	var calls []struct {
		Ctx  sdk.Context
		Poll exported.Poll
	}
	mock.lockHandleExpiredPoll.RLock()
	calls = mock.calls.HandleExpiredPoll
	mock.lockHandleExpiredPoll.RUnlock()
	return calls
}

// HandleResult calls HandleResultFunc.
func (mock *VoteHandlerMock) HandleResult(ctx sdk.Context, result codec.ProtoMarshaler) error {
	if mock.HandleResultFunc == nil {
		panic("VoteHandlerMock.HandleResultFunc: method is nil but VoteHandler.HandleResult was just called")
	}
	callInfo := struct {
		Ctx    sdk.Context
		Result codec.ProtoMarshaler
	}{
		Ctx:    ctx,
		Result: result,
	}
	mock.lockHandleResult.Lock()
	mock.calls.HandleResult = append(mock.calls.HandleResult, callInfo)
	mock.lockHandleResult.Unlock()
	return mock.HandleResultFunc(ctx, result)
}

// HandleResultCalls gets all the calls that were made to HandleResult.
// Check the length with:
//     len(mockedVoteHandler.HandleResultCalls())
func (mock *VoteHandlerMock) HandleResultCalls() []struct {
	Ctx    sdk.Context
	Result codec.ProtoMarshaler
} {
	var calls []struct {
		Ctx    sdk.Context
		Result codec.ProtoMarshaler
	}
	mock.lockHandleResult.RLock()
	calls = mock.calls.HandleResult
	mock.lockHandleResult.RUnlock()
	return calls
}

// IsFalsyResult calls IsFalsyResultFunc.
func (mock *VoteHandlerMock) IsFalsyResult(result codec.ProtoMarshaler) bool {
	if mock.IsFalsyResultFunc == nil {
		panic("VoteHandlerMock.IsFalsyResultFunc: method is nil but VoteHandler.IsFalsyResult was just called")
	}
	callInfo := struct {
		Result codec.ProtoMarshaler
	}{
		Result: result,
	}
	mock.lockIsFalsyResult.Lock()
	mock.calls.IsFalsyResult = append(mock.calls.IsFalsyResult, callInfo)
	mock.lockIsFalsyResult.Unlock()
	return mock.IsFalsyResultFunc(result)
}

// IsFalsyResultCalls gets all the calls that were made to IsFalsyResult.
// Check the length with:
//     len(mockedVoteHandler.IsFalsyResultCalls())
func (mock *VoteHandlerMock) IsFalsyResultCalls() []struct {
	Result codec.ProtoMarshaler
} {
	var calls []struct {
		Result codec.ProtoMarshaler
	}
	mock.lockIsFalsyResult.RLock()
	calls = mock.calls.IsFalsyResult
	mock.lockIsFalsyResult.RUnlock()
	return calls
}