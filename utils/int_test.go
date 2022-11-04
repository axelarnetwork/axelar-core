package utils_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/utils"
)

func TestMaxUint(t *testing.T) {
	x := utils.MaxUint
	assert.Equal(t, 256, x.BigInt().BitLen())
	assert.Panics(t, func() { x.AddUint64(1) })
	assert.Equal(t, 1, x.BigInt().Sign())

	y := big.NewInt(1)
	y.Lsh(y, 256)
	y.Sub(y, big.NewInt(1))
	assert.Equal(t, y, x.BigInt())
}
