package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type GenesisState struct {
	VotingInterval int64
}

func DefaultGenesisState() GenesisState {
	return GenesisState{VotingInterval: 10}
}

func ValidateGenesis(state GenesisState) error {
	if state.VotingInterval == 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "voting interval must be larger than 0")
	}

	return nil
}
