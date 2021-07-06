package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/axelarnetwork/axelar-core/utils"
)

// DefaultGenesisState represents the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		VotingThreshold: utils.Threshold{
			Numerator:   2,
			Denominator: 3,
		},
	}
}

// Validate validates the genesis state
func (m GenesisState) Validate() error {
	if m.VotingThreshold.Numerator < 0 || m.VotingThreshold.Denominator <= 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "voting threshold must contain positive integers")
	}

	if m.VotingThreshold.Numerator > m.VotingThreshold.Denominator {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "voting threshold must be lesser than or equal to 1")
	}

	return nil
}
