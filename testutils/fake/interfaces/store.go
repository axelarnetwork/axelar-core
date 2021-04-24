package interfaces

import sdkTypes "github.com/cosmos/cosmos-sdk/types"

//go:generate moq -out ./mock/store.go -pkg mock . MultiStore KVStore

// Interface wrappers for mocking
type (
	// MultiStore wrapper for github.com/cosmos/cosmos-sdk/types.MultiStore
	MultiStore sdkTypes.MultiStore
	// KVStore wrapper for github.com/cosmos/cosmos-sdk/types.KVStore
	KVStore sdkTypes.KVStore
)
