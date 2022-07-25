package rpc

import "github.com/axelarnetwork/axelar-core/x/tss/tofnd"

//go:generate moq -pkg mock -out ./mock/rpcClient.go . Client MultiSigClient
// Client defines the interface of a grpc client to communicate with tofnd
type Client interface {
	tofnd.GG20Client
}

// MultiSigClient defines the interface of a grpc client to communicate with tofnd Multisig service
type MultiSigClient interface {
	tofnd.MultisigClient
}
