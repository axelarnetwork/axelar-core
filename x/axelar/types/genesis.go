package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type GenesisState struct {
	VotingInterval  int64
	VotingThreshold VotingThreshold
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		VotingInterval: 10,
		VotingThreshold: VotingThreshold{
			Numerator:   2,
			Denominator: 3,
		},
	}
}

func ValidateGenesis(state GenesisState) error {
	if state.VotingInterval == 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "voting interval must be larger than 0")
	}
	if state.VotingThreshold.Numerator < 0 || state.VotingThreshold.Denominator <= 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "voting threshold must contain positive integers")
	}

	if state.VotingThreshold.Numerator > state.VotingThreshold.Denominator {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "voting threshold must be lesser than or equal to 1")
	}

	return nil
}
