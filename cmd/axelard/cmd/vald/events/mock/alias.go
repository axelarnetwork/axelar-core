package mock

import (
	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
)

//go:generate moq -out ./imported.go -pkg mock . Query Bus Subscriber

type (
	// Query interface alias for mocking
	Query tmpubsub.Query
	// Bus interface alias for mocking
	Bus pubsub.Bus
	// Subscriber interface alias for mocking
	Subscriber pubsub.Subscriber
)
