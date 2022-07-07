package testutils

import (
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
)

// RandThreshold returns a random Threshold
func RandThreshold() utils.Threshold {
	return utils.NewThreshold(rand.I64Between(1, 101), 100)
}
