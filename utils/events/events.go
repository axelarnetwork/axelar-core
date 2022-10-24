package events

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/axelarnetwork/utils/funcs"
)

// Emit provides a shorthand to emit an event through the context's event manager
func Emit(ctx sdk.Context, evs ...proto.Message) {
	// the tests in this package ensure this will never panic
	if len(evs) == 1 {
		funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(evs[0]))
	} else {
		funcs.MustNoErr(ctx.EventManager().EmitTypedEvents(evs...))
	}
}
