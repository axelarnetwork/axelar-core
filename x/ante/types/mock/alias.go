package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

//go:generate moq -pkg mock -out ./alias_mock.go . Tx Msg

// Tx alias for mocking
type Tx sdk.Tx

// Msg alias for mocking
type Msg interface {
	sdk.Msg
	descriptor.Message
}
