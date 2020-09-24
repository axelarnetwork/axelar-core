package types

type GenesisState struct {
}

func NewGenesisState() GenesisState {
	return GenesisState{}
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func ValidateGenesis(_ GenesisState) error {
	return nil
}
