package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IsMet returns true if share > threshold * total
func (m Threshold) IsMet(share sdk.Int, total sdk.Int) bool {
	return share.MulRaw(m.Denominator).GT(total.MulRaw(m.Numerator))
}
