package tss

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProcessNewBlockHeader handles timeout on new block header
func (mgr *Mgr) ProcessNewBlockHeader(blockHeight int64, _ []sdk.Attribute) error {
	for {
		session := mgr.timeoutQueue.top()

		if session == nil {
			return nil
		}

		mgr.Logger.Debug(fmt.Sprintf("session ID: %s, session timeout: %d, block height: %d", session.id, session.timeoutAt, blockHeight))

		if session.timeoutAt > blockHeight {
			return nil
		}

		mgr.timeoutQueue.dequeue()
		close(session.timeout)
	}
}
