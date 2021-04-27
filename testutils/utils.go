// Package testutils provides general purpose utility functions for unit/integration testing.
package testutils

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/std"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	ethereum "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/types"
)

var (
	cdc *codec.LegacyAmino
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	cdc = encodingConfig.Amino
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// Add new modules here so tests have access to marshalling the registered ethereum
	vote.RegisterLegacyAminoCodec(cdc)
	bitcoin.RegisterLegacyAminoCodec(cdc)
	bitcoin.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	tss.RegisterLegacyAminoCodec(cdc)
	tss.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	broadcast.RegisterLegacyAminoCodec(cdc)
	broadcast.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	snapshot.RegisterLegacyAminoCodec(cdc)
	ethereum.RegisterLegacyAminoCodec(cdc)
	ethereum.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	return encodingConfig
}

// Func wraps a regular testing function so it can be used as a pointer function receiver
type Func func(t *testing.T)

// Repeat executes the testing function n times
func (f Func) Repeat(n int) Func {
	return func(t *testing.T) {
		for i := 0; i < n; i++ {
			f(t)
		}
	}
}

// Events wraps sdk.Events
type Events []abci.Event

// Filter returns a collection of events filtered by the predicate
func (fe Events) Filter(predicate func(events abci.Event) bool) Events {
	var filtered Events
	for _, event := range fe {
		if predicate(event) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}
