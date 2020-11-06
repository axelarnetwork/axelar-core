package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
)

type Vote struct {
	Tx exported.ExternalTx
	// Using a map instead of an array ensures that validators cannot vote multiple times.
	// The actual validator address is needed frequently,
	// therefore the confirmations not only record the ValAddress string and a bool (emulating a hash set),
	// but the []byte representation as well
	Confirmations map[string]sdk.ValAddress
}

type VotingThreshold struct {
	// split threshold into Numerator and denominator to avoid floating point errors down the line
	Numerator   int64
	Denominator int64
}

func (t VotingThreshold) IsMet(accept sdk.Int, total sdk.Int) bool {
	return accept.MulRaw(t.Numerator).GT(total.MulRaw(t.Denominator))
}
