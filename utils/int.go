package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/math"
)

var (
	// MaxInt specifies the max sdk.Int value
	MaxInt = sdk.NewIntFromBigInt(math.MaxBig256)

	// MaxUint specifies the max sdk.Uint value
	MaxUint = sdk.NewUintFromBigInt(math.MaxBig256)
)
