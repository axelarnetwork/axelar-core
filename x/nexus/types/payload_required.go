package types

// AxelarnetModule is the module name of the axelarnet destination route.
const AxelarnetModule = "axelarnet"

// RequiresPayload reports whether the route for the given destination module can only deliver a
// message when the original payload is supplied.
func RequiresPayload(module string) bool {
	return module == AxelarnetModule
}
