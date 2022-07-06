package types

//go:generate moq -out ./mock/expected_keepers.go -pkg mock . Keeper

// Keeper is implemented by this module's keeper
type Keeper interface{}
