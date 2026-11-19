package vald

import (
	"testing"
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/stretchr/testify/require"
)

func TestValidateNodeStatus(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	freshStatus := func() *coretypes.ResultStatus {
		return &coretypes.ResultStatus{
			SyncInfo: coretypes.SyncInfo{
				LatestBlockHeight: 100,
				LatestBlockTime:   now.Add(-5 * time.Second),
			},
		}
	}

	tests := []struct {
		name            string
		status          *coretypes.ResultStatus
		maxAge          time.Duration
		wantErrContains string
	}{
		{
			name:   "fresh node",
			status: freshStatus(),
			maxAge: 15 * time.Second,
		},
		{
			name:            "invalid maximum age",
			status:          freshStatus(),
			maxAge:          0,
			wantErrContains: "must be positive",
		},
		{
			name:            "empty status",
			status:          nil,
			maxAge:          15 * time.Second,
			wantErrContains: "response is empty",
		},
		{
			name: "catching up",
			status: &coretypes.ResultStatus{
				SyncInfo: coretypes.SyncInfo{
					LatestBlockHeight: 100,
					LatestBlockTime:   now.Add(-5 * time.Second),
					CatchingUp:        true,
				},
			},
			maxAge:          15 * time.Second,
			wantErrContains: "is catching up",
		},
		{
			name: "missing height",
			status: &coretypes.ResultStatus{
				SyncInfo: coretypes.SyncInfo{LatestBlockTime: now},
			},
			maxAge:          15 * time.Second,
			wantErrContains: "height is not positive",
		},
		{
			name: "missing block time",
			status: &coretypes.ResultStatus{
				SyncInfo: coretypes.SyncInfo{LatestBlockHeight: 100},
			},
			maxAge:          15 * time.Second,
			wantErrContains: "block time is missing",
		},
		{
			name: "stalled consensus",
			status: &coretypes.ResultStatus{
				SyncInfo: coretypes.SyncInfo{
					LatestBlockHeight: 100,
					LatestBlockTime:   now.Add(-16 * time.Second),
				},
			},
			maxAge:          15 * time.Second,
			wantErrContains: "latest block is stale",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNodeStatus(tt.status, tt.maxAge, now)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErrContains)
		})
	}
}
