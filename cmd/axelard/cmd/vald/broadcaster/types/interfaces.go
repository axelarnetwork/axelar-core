package types

import (
	"context"

	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

//go:generate moq -out ./mock/types.go -pkg mock -stub . Client Broadcaster AccountRetriever Keyring Info Pipeline
// go:generate moq -out ./mock/interfaces.go -pkg mock -stub . Keybase Client KVStore Info Msg

// interface wraps for testing purposes
type (
	// AccountRetriever wrapper for github.com/cosmos/cosmos-sdk/client.AccountRetriever
	AccountRetriever sdkClient.AccountRetriever
	// Client wrapper for github.com/tendermint/tendermint/rpc/client.Client
	Client rpcclient.Client
	// Keyring wrapper for github.com/cosmos/cosmos-sdk/crypto/keyring.Keyring
	Keyring keyring.Keyring
	// Info wrapper for github.com/cosmos/cosmos-sdk/crypto/keyring.Info
	Info keyring.Info
)

// Broadcaster interface allows the submission of messages to the axelar network
type Broadcaster interface {
	Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error)
}

// Pipeline represents an execution pipeline
type Pipeline interface {
	Push(f func() error, retryOnError func(error) bool) error
	Close()
}
