package keeper

import (
	"sync/atomic"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/monads/cached"
)

// TestCache_InterfaceDispatchMustCacheReads proves that HasVoted and HasVotedCorrectly
// called through the exported.Poll interface only load tallied votes from storage once.
// With value receivers, each interface dispatch copies the poll struct, defeating the cache.
// With pointer receivers, the cache persists across calls.
func TestCache_InterfaceDispatchMustCacheReads(t *testing.T) {
	voter1 := sdk.ValAddress("voter1______________")
	voter2 := sdk.ValAddress("voter2______________")

	var loadCount int64

	p := &poll{
		PollMetadata: exported.PollMetadata{
			State: exported.Completed,
		},
		talliedVotes: cached.New(func() []types.TalliedVote {
			atomic.AddInt64(&loadCount, 1)
			return []types.TalliedVote{
				{
					Tally:       math.NewUint(10),
					IsVoterLate: map[string]bool{voter1.String(): false},
				},
			}
		}),
		passingWeight: cached.New(func() math.Uint {
			return math.OneUint()
		}),
	}

	var pollIface exported.Poll = p

	const numCalls = 20

	atomic.StoreInt64(&loadCount, 0)
	for range numCalls {
		pollIface.HasVoted(voter1)
		pollIface.HasVoted(voter2)
	}

	hasVotedLoads := atomic.LoadInt64(&loadCount)
	t.Logf("HasVoted: %d cache loads for %d calls", hasVotedLoads, numCalls*2)
	assert.Equal(t, int64(1), hasVotedLoads,
		"talliedVotes should be loaded exactly once across all HasVoted calls, got %d", hasVotedLoads)

	p.talliedVotes = cached.New(func() []types.TalliedVote {
		atomic.AddInt64(&loadCount, 1)
		return []types.TalliedVote{
			{
				Tally:       math.NewUint(10),
				IsVoterLate: map[string]bool{voter1.String(): false},
			},
		}
	})
	atomic.StoreInt64(&loadCount, 0)
	for range numCalls {
		pollIface.HasVotedCorrectly(voter1)
		pollIface.HasVotedCorrectly(voter2)
	}

	hasVotedCorrectlyLoads := atomic.LoadInt64(&loadCount)
	t.Logf("HasVotedCorrectly: %d cache loads for %d calls", hasVotedCorrectlyLoads, numCalls*2)
	assert.Equal(t, int64(1), hasVotedCorrectlyLoads,
		"talliedVotes should be loaded exactly once across all HasVotedCorrectly calls, got %d", hasVotedCorrectlyLoads)
}
