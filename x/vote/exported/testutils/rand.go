package testutils

import (
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/slices"
	"github.com/axelarnetwork/utils/test/rand"
)

// RandomPollID generates a random PollID
func RandomPollID() exported.PollID {
	return exported.PollID(rand.PosI64())
}

// RandomPollParticipants generates random PollParticipants
func RandomPollParticipants() exported.PollParticipants {
	return exported.PollParticipants{
		PollID:       RandomPollID(),
		Participants: slices.Expand2(rand2.ValAddr, int(rand.I64Between(1, 20))),
	}
}
