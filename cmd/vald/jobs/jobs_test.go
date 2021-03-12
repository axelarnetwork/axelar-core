package jobs

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

func TestJobManager_Wait(t *testing.T) {
	jobCount := rand.I64Between(0, 100)
	var jobs []Job
	var expectedErrCount int64
	for i := int64(0); i < jobCount; i++ {
		errCount := rand.I64Between(0, 10000)
		expectedErrCount += errCount
		job := randomJob(errCount, i)
		jobs = append(jobs, job)
	}
	var actualErrCount int64
	errHandler := func(err error) { actualErrCount++ }
	mgr := NewMgr(errHandler)

	mgr.AddJobs(jobs...)
	mgr.Wait()
	assert.Equal(t, expectedErrCount, actualErrCount)
}

// this extraction is needed to close over the loop counter i
func randomJob(errCount int64, i int64) Job {
	return func(e chan<- error) {
		for j := int64(0); j < errCount; j++ {
			e <- fmt.Errorf("error %d by job %d", j, i)
		}
		duration := time.Duration(rand.I64Between(0, 100)) * time.Millisecond
		time.Sleep(duration)
	}
}
