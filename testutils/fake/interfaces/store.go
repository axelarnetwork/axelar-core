package interfaces

import sdkTypes "github.com/cosmos/cosmos-sdk/types"

//go:generate moq -out ./mock/store.go -pkg mock . Multistore KVStore

// Interface wrappers for mocking
type (
	Multistore sdkTypes.MultiStore
	KVStore    sdkTypes.KVStore
)
