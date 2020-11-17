package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

type Vote struct {
	Validator sdk.ValAddress
	Confirms  bool
}

type Poll struct {
	Tx    exported.ExternalTx
	Votes []Vote
}

type VotingThreshold struct {
	// split threshold into Numerator and denominator to avoid floating point errors down the line
	Numerator   int64
	Denominator int64
}

func (t VotingThreshold) IsMet(accept sdk.Int, total sdk.Int) bool {
	return accept.MulRaw(t.Denominator).GT(total.MulRaw(t.Numerator))
}
