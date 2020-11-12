package cli

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
)

func TestGetCmdSignStart_UnmarshalTxHash(t *testing.T) {
	input := "\"1EtL4N8J5dE37bL53qRRbtSW0dQbGYjseMr4ks3lbNM=\""
	bz := []byte(input)
	var realMsg []byte
	cdc := codec.New()
	cdc.MustUnmarshalJSON(bz, &realMsg)
}
