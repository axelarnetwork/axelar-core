package mock

import (
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

//go:generate moq -pkg mock -out ./alias_mocks.go . Client AccountRetriever Keyring

// interface wraps for testing purposes
type (
	// AccountRetriever wrapper for github.com/cosmos/cosmos-sdk/client.AccountRetriever
	AccountRetriever sdkClient.AccountRetriever
	// Client wrapper for github.com/tendermint/tendermint/rpc/client.Client
	Client rpcclient.Client
	// Keyring wrapper for github.com/cosmos/cosmos-sdk/crypto/keyring.Keyring
	Keyring keyring.Keyring
)
