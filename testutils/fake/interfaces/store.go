package interfaces

import sdkTypes "github.com/cosmos/cosmos-sdk/types"

// Interface wrappers for mocking
//
//go:generate moq -out ./mock/store.go -pkg mock . MultiStore CacheMultiStore KVStore
type (
	// MultiStore wrapper for github.com/cosmos/cosmos-sdk/types.MultiStore
	MultiStore sdkTypes.MultiStore
	// CacheMultiStore wrapper for github.com/cosmos/cosmos-sdk/types.CacheMultiStore
	CacheMultiStore sdkTypes.CacheMultiStore
	// KVStore wrapper for github.com/cosmos/cosmos-sdk/types.KVStore
	KVStore sdkTypes.KVStore
)
