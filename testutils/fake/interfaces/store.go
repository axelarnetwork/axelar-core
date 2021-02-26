package interfaces

import sdkTypes "github.com/cosmos/cosmos-sdk/types"

//go:generate moq -out ./mock/store.go -pkg mock . MultiStore KVStore

// Interface wrappers for mocking
type (
	MultiStore sdkTypes.MultiStore
	KVStore    sdkTypes.KVStore
)
