package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/axelarnetwork/axelar-core/utils"
)

type GenesisState struct {
	VotingInterval  int64
	VotingThreshold utils.Threshold
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		VotingInterval: 10,
		VotingThreshold: utils.Threshold{
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
