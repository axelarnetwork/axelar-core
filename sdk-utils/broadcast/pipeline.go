package broadcast

import (
	"context"
	"fmt"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/utils"
	"github.com/axelarnetwork/utils/log"
)

// retryPipeline manages serialized execution of functions with retry on error
type retryPipeline struct {
	c          chan func()
	backOff    utils.BackOff
	maxRetries int
}

// newPipelineWithRetry returns a retryPipeline with the given configuration
func newPipelineWithRetry(cap int, maxRetries int, backOffStrategy utils.BackOff) *retryPipeline {
	p := &retryPipeline{
		c:          make(chan func(), cap),
		backOff:    backOffStrategy,
		maxRetries: maxRetries,
	}

	go func() {
		for f := range p.c {
			f()
		}
	}()

	return p
}

// Push adds the given function to the serialized execution retryPipeline
func (p *retryPipeline) Push(ctx context.Context, f func(context.Context) error, retryOnError func(error) bool) error {
	e := make(chan error, 1)
	p.c <- func() { e <- p.retry(ctx, f, retryOnError) }
	return <-e
}

func (p retryPipeline) retry(ctx context.Context, f func(context.Context) error, retryOnError func(error) bool) error {
	var err error
	for i := 0; i <= p.maxRetries; i++ {
		ctx = log.Append(ctx, "num_attempts", i+1)
		err = f(ctx)
		if err == nil {
			if i > 0 {
				log.FromCtx(ctx).Info("successful execution after backoff")
			}
			return nil
		}

		if !retryOnError(err) {
			return err
		}

		if i < p.maxRetries {
			timeout := p.backOff(i)
			log.FromCtx(ctx).Infof("backing off (retry in %v )", timeout)
			time.Sleep(timeout)
		}
	}
	return sdkerrors.Wrap(err, fmt.Sprintf("aborting after %d retries", p.maxRetries))
}

// Close closes the retryPipeline
func (p *retryPipeline) Close() {
	close(p.c)
}
