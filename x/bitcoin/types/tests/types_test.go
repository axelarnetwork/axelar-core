package tests

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

func TestOutPointInfo_Equals(t *testing.T) {
	// Take care to have identical slices with different pointers
	var bz1, bz2 []byte
	for _, b := range rand.I64GenBetween(0, 256).Take(chainhash.HashSize) {
		bz1 = append(bz1, byte(b))
		bz2 = append(bz2, byte(b))
	}
	hash1, err := chainhash.NewHash(bz1)
	if err != nil {
		panic(err)
	}
	hash2, err := chainhash.NewHash(bz2)
	if err != nil {
		panic(err)
	}

	op1 := types.OutPointInfo{
		OutPoint: wire.NewOutPoint(hash1, 3),
		Amount:   0,
		Address:  "recipient",
	}

	op2 := types.OutPointInfo{
		OutPoint: wire.NewOutPoint(hash2, 3),
		Amount:   0,
		Address:  "recipient",
	}

	assert.True(t, op1.Equals(op2))
	assert.Equal(t, op1, op2)
}
