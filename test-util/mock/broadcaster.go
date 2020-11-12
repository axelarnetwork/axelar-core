package mock

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
)

type broadcaster struct {
	in        chan<- sdk.Msg
	sender    sdk.AccAddress
	val2Proxy map[string]sdk.AccAddress
	proxy2Val map[string]sdk.ValAddress
	principal sdk.ValAddress
	cdc       *codec.Codec
}

// NewBroadcaster creates a new broadcaster mock that sends messages to the blockchainIn channel.
// Messages are sent from the sender account, while the local validator account is given by localPrincipal.
func NewBroadcaster(cdc *codec.Codec, sender sdk.AccAddress, localPrincipal sdk.ValAddress, blockchainIn chan<- sdk.Msg) exported.Broadcaster {
	return broadcaster{
		cdc:       cdc,
		in:        blockchainIn,
		sender:    sender,
		val2Proxy: make(map[string]sdk.AccAddress),
		proxy2Val: make(map[string]sdk.ValAddress),
		principal: localPrincipal,
	}
}

func (b broadcaster) Broadcast(_ sdk.Context, msgs []exported.ValidatorMsg) error {
	for _, msg := range msgs {
		msg.SetSender(b.sender)

		/*
			exported.ValidatorMsg is usually implemented by a pointer.
			However, handler expect to receive the message by value and do a switch on the message type.
			If they receive the pointer they won't recognize the correct message type.
			By marshalling and unmarshalling into sdk.Msg we get the message by value.
		*/
		bz := b.cdc.MustMarshalBinaryLengthPrefixed(msg)
		var sentMsg sdk.Msg
		b.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &sentMsg)

		b.in <- sentMsg
	}
	return nil
}

func (b broadcaster) RegisterProxy(_ sdk.Context, principal sdk.ValAddress, proxy sdk.AccAddress) error {
	b.val2Proxy[principal.String()] = proxy
	b.proxy2Val[proxy.String()] = principal
	return nil
}

func (b broadcaster) GetPrincipal(_ sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
	return b.proxy2Val[proxy.String()]
}

func (b broadcaster) GetProxyCount(_ sdk.Context) uint32 {
	return uint32(len(b.val2Proxy))
}

func (b broadcaster) GetLocalPrincipal(_ sdk.Context) sdk.ValAddress {
	return b.principal
}
