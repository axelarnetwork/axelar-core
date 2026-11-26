package utils

import (
	"math/rand"
	"time"
)

// BackOff computes the next back-off duration
type BackOff func(currentRetryCount int) time.Duration

// ExponentialBackOff returns a back-off with uniform jitter between minTimeout
// and minTimeout*(1+2^currentRetryCount). Retry counts start at zero.
func ExponentialBackOff(minTimeout time.Duration) BackOff {
	return func(currentRetryCount int) time.Duration {
		jitter := rand.Float64()
		strategy := 1 << currentRetryCount
		backoff := (1 + float64(strategy)*jitter) * minTimeout.Seconds() * float64(time.Second)

		return time.Duration(backoff)
	}
}

// LinearBackOff returns a back-off with uniform jitter between minTimeout
// and minTimeout*(1+currentRetryCount). Retry counts start at zero.
func LinearBackOff(minTimeout time.Duration) BackOff {
	return func(currentRetryCount int) time.Duration {
		jitter := rand.Float64()
		strategy := float64(currentRetryCount)

		backoff := (1 + strategy*jitter) * minTimeout.Seconds() * float64(time.Second)
		return time.Duration(backoff)
	}
}
