package broadcaster

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/reward/types"
)

// RefundableBroadcaster only sends RefundMsgRequest msgs
type RefundableBroadcaster struct {
	broadcaster *Broadcaster
}

// Broadcast wraps all given msgs into RefundMsgRequest msgs before broadcasting them
func (b *RefundableBroadcaster) Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	var refundables []sdk.Msg
	for _, msg := range msgs {
		refundables = append(refundables, types.NewRefundMsgRequest(b.broadcaster.clientCtx.FromAddress, msg))
	}
	return b.broadcaster.Broadcast(ctx, refundables...)
}

// WithRefund wraps a broadcaster into a RefundableBroadcaster
func WithRefund(b *Broadcaster) *RefundableBroadcaster {
	return &RefundableBroadcaster{broadcaster: b}
}
