package types

// DefaultGenesisState represents the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

// Validate validates the genesis state
func (g GenesisState) Validate() error {
	return nil
}
