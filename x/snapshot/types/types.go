package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Validator is an interface for a Cosmos validator account
type Validator interface {
	GetOperator() sdk.ValAddress
	GetConsAddr() sdk.ConsAddress
	GetConsensusPower() int64
	IsJailed() bool
}
