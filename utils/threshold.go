package utils

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IsMet returns true if share > threshold * total
func (m Threshold) IsMet(share sdk.Int, total sdk.Int) bool {
	return share.MulRaw(m.Denominator).GT(total.MulRaw(m.Numerator))
}

func (m Threshold) GT(t Threshold) bool {
	return sdk.NewInt(m.Numerator).MulRaw(t.Denominator).GT(sdk.NewInt(m.Denominator).MulRaw(t.Numerator))
}

func (m Threshold) GTE(t Threshold) bool {
	return sdk.NewInt(m.Numerator).MulRaw(t.Denominator).GTE(sdk.NewInt(m.Denominator).MulRaw(t.Numerator))
}

func (m Threshold) LT(t Threshold) bool {
	return sdk.NewInt(m.Numerator).MulRaw(t.Denominator).LT(sdk.NewInt(m.Denominator).MulRaw(t.Numerator))
}

func (m Threshold) LTE(t Threshold) bool {
	return sdk.NewInt(m.Numerator).MulRaw(t.Denominator).LTE(sdk.NewInt(m.Denominator).MulRaw(t.Numerator))
}

func (m Threshold) Validate() error {
	if m.Denominator == 0 {
		return fmt.Errorf("Denominator must not be 0")
	}

	return nil
}
