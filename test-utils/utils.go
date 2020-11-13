package test_utils

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	axTypes "github.com/axelarnetwork/axelar-core/x/axelar/types"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	btcTypes "github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

var (
	cdc *codec.Codec
)

// Codec creates a codec for testing with all necessary types registered
func Codec() *codec.Codec {
	if cdc != nil {
		return cdc
	}

	cdc = codec.New()

	sdk.RegisterCodec(cdc)

	// Add new modules here so tests have access to marshalling the registered types
	axTypes.RegisterCodec(cdc)
	btcTypes.RegisterCodec(cdc)
	tssTypes.RegisterCodec(cdc)
	broadcastTypes.RegisterCodec(cdc)

	cdc.Seal()
	return cdc
}

func StartTimeout(t time.Duration) chan struct{} {
	timeOut := make(chan struct{})
	go func() {
		time.Sleep(t)
		close(timeOut)
	}()
	return timeOut
}
