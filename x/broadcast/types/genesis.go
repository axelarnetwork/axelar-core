package types

// DefaultGenesisState represents the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

// Validate validates the genesis state
func (m GenesisState) Validate() error {
	return nil
}
