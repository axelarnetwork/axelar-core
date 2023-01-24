package multisig

import "github.com/axelarnetwork/axelar-core/x/tss/tofnd"

// Client defines the interface of a grpc client to communicate with tofnd Multisig service
//
//go:generate moq -pkg mock -out ./mock/rpcClient.go . Client
type Client interface {
	tofnd.MultisigClient
}
