package types

type GenesisState struct {
}

func NewGenesisState() GenesisState {
	return GenesisState{}
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func ValidateGenesis(data GenesisState) error {
	return nil
}
