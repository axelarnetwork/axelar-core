package mock

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

//go:generate moq -pkg mock -out ./alias_mocks.go . Client AccountRetriever Keyring Info

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
