package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
)

type broadcaster struct {
	in        chan<- sdk.Msg
	sender    sdk.AccAddress
	val2Proxy map[string]sdk.AccAddress
	proxy2Val map[string]sdk.ValAddress
	principal sdk.ValAddress
}

// NewBroadcaster creates a new broadcaster mock that sends messages to the blockchainIn channel.
// Messages are sent from the sender account, while the local validator account is given by localPrincipal.
func NewBroadcaster(sender sdk.AccAddress, localPrincipal sdk.ValAddress, blockchainIn chan<- sdk.Msg) exported.Broadcaster {
	return broadcaster{
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
		b.in <- msg
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
