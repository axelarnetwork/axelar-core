package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/math"
)

var (
	// MaxUint specifies the max sdk.Uint value
	MaxUint = sdk.NewUintFromBigInt(math.MaxBig256)
)
