package mock

import (
	"io"

	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

//go:generate moq -out ./types.go -pkg mock . SignClient Query Bus Subscriber ReadWriteSeekTruncateCloser

type (
	// SignClient interface alias for mocking
	SignClient rpcclient.SignClient
	// Query interface alias for mocking
	Query tmpubsub.Query
	// Bus interface alias for mocking
	Bus pubsub.Bus
	// Subscriber interface alias for mocking
	Subscriber pubsub.Subscriber
	// ReadWriteSeekTruncateCloser interface for mocking. Duplicated to prevent cyclic dependencies
	ReadWriteSeekTruncateCloser interface {
		io.ReadWriteSeeker
		Truncate(size int64) error
		Close() error
	}
)
