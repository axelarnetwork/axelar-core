package types

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

//go:generate moq -out ./mock/types.go -pkg mock -stub . Client Broadcaster AccountRetriever Keyring Info Pipeline
// go:generate moq -out ./mock/interfaces.go -pkg mock -stub . Keybase Client KVStore Info Msg

// interface wraps for testing purposes
type (
	AccountRetriever sdkClient.AccountRetriever
	Client           rpcclient.Client
	Keyring          keyring.Keyring
	Info             keyring.Info
)

// Broadcaster interface allows the submission of messages to the axelar network
type Broadcaster interface {
	Broadcast(msgs ...sdk.Msg) error
}

// Pipeline represents an execution pipeline
type Pipeline interface {
	Push(func() error) error
	Close()
}
