package utils

import (
	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common/math"
)

var (
	// MaxInt specifies the max sdk.Int value
	MaxInt = sdkmath.NewIntFromBigInt(math.MaxBig256)

	// MaxUint specifies the max sdk.Uint value
	MaxUint = sdkmath.NewUintFromBigInt(math.MaxBig256)
)
