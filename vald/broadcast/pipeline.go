package broadcast

import (
	"fmt"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/utils"
)

// retryPipeline manages serialized execution of functions with retry on error
type retryPipeline struct {
	c          chan func()
	backOff    utils.BackOff
	maxRetries int
	logger     log.Logger
}

// newPipelineWithRetry returns a retryPipeline with the given configuration
func newPipelineWithRetry(cap int, maxRetries int, backOffStrategy utils.BackOff, logger log.Logger) *retryPipeline {
	p := &retryPipeline{
		c:          make(chan func(), cap),
		backOff:    backOffStrategy,
		maxRetries: maxRetries,
		logger:     logger,
	}

	go func() {
		for f := range p.c {
			f()
		}
	}()

	return p
}

// Push adds the given function to the serialized execution retryPipeline
func (p *retryPipeline) Push(f func() error, retryOnError func(error) bool) error {
	e := make(chan error, 1)
	p.c <- func() { e <- p.retry(f, retryOnError) }
	return <-e
}

func (p retryPipeline) retry(f func() error, retryOnError func(error) bool) error {
	var err error
	logger := p.logger
	for i := 0; i <= p.maxRetries; i++ {
		logger = logger.With("num_attempts", i+1)
		err = f()
		if err == nil {
			if i > 0 {
				logger.Info("successful execution after backoff")
			}
			return nil
		}

		if !retryOnError(err) {
			return err
		}

		if i < p.maxRetries {
			timeout := p.backOff(i)
			logger.Info(sdkerrors.Wrapf(err, "backing off (retry in %v )", timeout).Error())
			time.Sleep(timeout)
		}
	}
	return sdkerrors.Wrap(err, fmt.Sprintf("aborting after %d retries", p.maxRetries))
}

// Close closes the retryPipeline
func (p *retryPipeline) Close() {
	close(p.c)
}
