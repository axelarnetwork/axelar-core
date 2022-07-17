package mock

import (
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

//go:generate moq -out ./mock.go -pkg mock . ValidatorI

// ValidatorI is an alias for sdk staking ValidatorI
type ValidatorI types.ValidatorI
