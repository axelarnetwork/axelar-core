package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Threshold struct {
	// split threshold into Numerator and denominator to avoid floating point errors down the line
	Numerator   int64
	Denominator int64
}

func (t Threshold) IsMet(accept sdk.Int, total sdk.Int) bool {
	return accept.MulRaw(t.Denominator).GT(total.MulRaw(t.Numerator))
}

func (t Threshold) IsMetBy(total sdk.Int) (part sdk.Int) {
	return sdk.NewDecFromInt(total.MulRaw(t.Numerator)).QuoRoundUp(sdk.NewDec(t.Denominator)).RoundInt()
}
