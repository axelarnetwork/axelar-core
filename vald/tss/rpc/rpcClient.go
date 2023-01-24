package rpc

import "github.com/axelarnetwork/axelar-core/x/tss/tofnd"

// MultiSigClient defines the interface of a grpc client to communicate with tofnd Multisig service
//
//go:generate moq -pkg mock -out ./mock/rpcClient.go . MultiSigClient
type MultiSigClient interface {
	tofnd.MultisigClient
}
