package types

// NewGenesisState is the constructor for GenesisState
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{Params: params}
}

// DefaultGenesisState returns a genesis state with default parameters
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams())
}

// Validate performs a validation check on the genesis parameters
func (m GenesisState) Validate() error {
	return m.Params.Validate()
}
