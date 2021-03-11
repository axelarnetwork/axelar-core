// Package testutils provides general purpose utility functions for unit/integration testing.
package testutils

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	ethereum "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/types"
)

var (
	cdc *codec.Codec
)

// Codec creates a codec for testing with all necessary types registered.
// This codec is not sealed so tests can add their own mock types.
func Codec() *codec.Codec {
	// Use cache if initialized before
	if cdc != nil {
		return cdc
	}

	cdc = codec.New()

	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	// Add new modules here so tests have access to marshalling the registered ethereum
	vote.RegisterCodec(cdc)
	bitcoin.RegisterCodec(cdc)
	tss.RegisterCodec(cdc)
	broadcast.RegisterCodec(cdc)
	snapshot.RegisterCodec(cdc)
	ethereum.RegisterCodec(cdc)

	return cdc
}
