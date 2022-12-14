package types

const (
	// ModuleName is the name of the module
	ModuleName = "evm"

	// ChainNamespace differentiated the key space for chains from the default evm namespace
	ChainNamespace = "chain"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for legacy query routing
	QuerierRoute = ModuleName

	// RestRoute to be used for rest routing
	RestRoute = ModuleName
)
