package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type GenesisState struct {
	ConfirmationHeight uint64
}

func DefaultGenesisState() GenesisState {
	return GenesisState{ConfirmationHeight: 6}
}

func ValidateGenesis(state GenesisState) error {
	if state.ConfirmationHeight < 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "transaction confirmation height must be greater than 0")
	}

	return nil
}
