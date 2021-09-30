package exported

import sdk "github.com/cosmos/cosmos-sdk/types"

// Refundable interface is used to register refundable message
type Refundable interface {
	sdk.Msg
}
