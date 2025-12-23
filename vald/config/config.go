package config

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// ValdConfig contains all necessary vald configurations
type ValdConfig struct {
	tss.TssConfig   `mapstructure:"tss"`
	BroadcastConfig `mapstructure:"broadcast"`
	// BatchSizeLimit is the maximum number of messages to include in a single batch when
	// the broadcaster merges multiple broadcast requests under high traffic.
	BatchSizeLimit int `mapstructure:"max_batch_size"`
	// BatchThreshold is the minimum number of pending broadcast requests in the queue
	// before the broadcaster starts merging them into batches. Below this threshold,
	// messages are broadcast individually.
	BatchThreshold int `mapstructure:"batch_threshold"`
	// MaxBlocksBehindLatest is the maximum number of blocks behind the latest until which
	// the cached height is considered valid. If vald is further behind, it discards the
	// stored height and starts from the latest block.
	MaxBlocksBehindLatest int64 `mapstructure:"max_blocks_behind_latest"`
	// EventNotificationsMaxRetries is the maximum number of retries when fetching blocks
	// from the Tendermint node via the block event subscription.
	EventNotificationsMaxRetries int `mapstructure:"event_notifications_max_retries"`
	// EventNotificationsBackOff is the backoff duration between retries when fetching blocks.
	EventNotificationsBackOff time.Duration `mapstructure:"event_notifications_back_off"`
	// MaxLatestBlockAge is the maximum age of a block for vald to consider it the latest.
	// If the latest block is older than this, vald waits for the node to sync.
	// Should be sufficiently larger than the block production time.
	MaxLatestBlockAge time.Duration `mapstructure:"max_latest_block_age"`
	// NoNewBlockPanicTimeout is the duration after which vald panics if no new block is seen.
	// This is a safety mechanism to detect and recover from stalled states. Once at least
	// one block has been seen, vald will panic if it does not see another before the timeout expires.
	NoNewBlockPanicTimeout time.Duration `mapstructure:"no_new_blocks_timeout"`
	// EVMConfig contains the configuration for each EVM chain bridge.
	EVMConfig []evm.EVMConfig `mapstructure:"axelar_bridge_evm"`
}

// DefaultValdConfig returns a configurations populated with default values
func DefaultValdConfig() ValdConfig {
	return ValdConfig{
		TssConfig:                    tss.DefaultConfig(),
		BroadcastConfig:              DefaultBroadcastConfig(),
		BatchSizeLimit:               250,
		BatchThreshold:               3,
		MaxBlocksBehindLatest:        50,
		MaxLatestBlockAge:            15 * time.Second,
		EVMConfig:                    evm.DefaultConfig(),
		EventNotificationsMaxRetries: 3,
		EventNotificationsBackOff:    1 * time.Second,
		NoNewBlockPanicTimeout:       2 * time.Minute,
	}
}

// BroadcastConfig is the configuration for transaction broadcasting
type BroadcastConfig struct {
	// MaxRetries is the maximum number of times to retry a failed broadcast.
	// All broadcasts go through a serialized queue, so retries block other broadcasts.
	MaxRetries int `mapstructure:"max_retries"`
	// MinSleepBeforeRetry is the base sleep duration for backoff between retries.
	MinSleepBeforeRetry time.Duration `mapstructure:"min_sleep_before_retry"`
	// MaxTimeout is the maximum time to wait for a transaction to be included in a block
	// after it has been broadcast.
	MaxTimeout time.Duration `mapstructure:"max_timeout"`
	// FeeGranter is the address of the fee granter account. Currently unused.
	FeeGranter sdk.AccAddress `mapstructure:"fee_granter"`
	// ConfirmationPollingInterval is how often to poll the node to check if a broadcast
	// transaction has been included in a block.
	ConfirmationPollingInterval time.Duration `mapstructure:"confirmation_polling_interval"`
}

// DefaultBroadcastConfig returns a configurations populated with default values
func DefaultBroadcastConfig() BroadcastConfig {
	return BroadcastConfig{
		MaxRetries:                  5,
		MinSleepBeforeRetry:         1 * time.Second,
		MaxTimeout:                  10 * time.Second,
		ConfirmationPollingInterval: 400 * time.Millisecond,
	}
}
