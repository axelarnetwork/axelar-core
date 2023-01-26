package rpc

import "github.com/axelarnetwork/axelar-core/x/tss/tofnd"

//go:generate moq -pkg mock -out ./mock/rpcClient.go . MultiSigClient

// MultiSigClient defines the interface of a grpc client to communicate with tofnd Multisig service
type MultiSigClient interface {
	tofnd.MultisigClient
}
