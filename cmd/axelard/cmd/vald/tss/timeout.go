package tss

import (
	"fmt"
)

// ProcessNewBlockHeader handles timeout on new block header
func (mgr *Mgr) ProcessNewBlockHeader(blockHeight int64) {
	for {
		session := mgr.timeoutQueue.Top()

		if session == nil {
			return
		}

		mgr.Logger.Debug(fmt.Sprintf("session ID: %s, session timeout: %d, block height: %d", session.ID, session.TimeoutAt, blockHeight))

		if session.TimeoutAt > blockHeight {
			return
		}

		mgr.timeoutQueue.Dequeue()
		session.Timeout()
	}
}
