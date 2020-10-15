package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
)

type PreVote struct {
	Tx          exported.ExternalTx
	LocalAccept bool
}

type Vote struct {
	Tx exported.ExternalTx
	// using a map instead of an array ensures that validators cannot vote multiple times
	Confirmations map[string]sdk.ValAddress
}
