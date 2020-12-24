package types

type GenesisState struct {
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func ValidateGenesis(_ GenesisState) error {
	return nil
}
