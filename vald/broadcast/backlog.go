package broadcast

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type backlog struct {
	tail chan broadcastTask
	head broadcastTask
}

func (bl *backlog) Pop() broadcastTask {
	bl.loadHead()

	next := bl.head
	bl.head = broadcastTask{}
	return next
}

func (bl *backlog) loadHead() {
	if len(bl.head.Msgs) == 0 {
		bl.head = <-bl.tail
	}
}

func (bl *backlog) Push(task broadcastTask) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		if len(task.Msgs) == 0 {
			task.Callback <- broadcastResult{
				Err: errors.New("task without msg pushed into backlog"),
			}
			return
		}
		bl.tail <- task
	}()

	return done
}

func (bl *backlog) Peek() broadcastTask {
	bl.loadHead()

	return bl.head
}

func (bl *backlog) Len() int {
	// do not block in this function because it might be used to inform other calls like Peek()
	if len(bl.head.Msgs) == 0 {
		// head is not currently loaded
		return len(bl.tail)
	}

	return 1 + len(bl.tail)
}

type broadcastTask struct {
	Ctx      context.Context
	Msgs     []sdk.Msg
	Callback chan<- broadcastResult
}

type broadcastResult struct {
	Response *sdk.TxResponse
	Err      error
}
