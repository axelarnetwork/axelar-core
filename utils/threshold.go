package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IsMet returns true if the share is greater than threshold * total
func (m Threshold) IsMet(share sdk.Int, total sdk.Int) bool {
	return share.MulRaw(m.Denominator).GT(total.MulRaw(m.Numerator))
}

// IsMetBy returns the minimum share that meets the threshold of the total
func (m Threshold) IsMetBy(total sdk.Int) (share sdk.Int) {
	return sdk.NewDecFromInt(total.MulRaw(m.Numerator)).QuoRoundUp(sdk.NewDec(m.Denominator)).RoundInt()
}
