package tss

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProcessNewBlockHeader handles timeout on new block header
func (mgr *Mgr) ProcessNewBlockHeader(blockHeight int64, _ []sdk.Attribute) error {
	for {
		session := mgr.timeoutQueue.Top()

		if session == nil {
			return nil
		}

		mgr.Logger.Debug(fmt.Sprintf("session ID: %s, session timeout: %d, block height: %d", session.ID, session.TimeoutAt, blockHeight))

		if session.TimeoutAt > blockHeight {
			return nil
		}

		mgr.timeoutQueue.Dequeue()
		session.Timeout()
	}
}
