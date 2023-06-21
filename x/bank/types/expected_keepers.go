package types

import (
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . BankKeeper

// BankKeeper provides functionality to the bank module
type BankKeeper interface {
	bankkeeper.ViewKeeper
	bankkeeper.SendKeeper
	bankkeeper.Keeper
}
