package types

// ModuleName must be different from "staking" otherwise we conflict with the sdk staking module
const (
	// ModuleName is the name of the module
	ModuleName = "snapshot"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName
)
