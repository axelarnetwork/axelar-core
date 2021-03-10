package jobs

import "sync"

// Job represents a (long-running) process that can be spawned on a separate go-routine.
// When encountering an error, a Job should send that error to the given channel and continue to run.
type Job func(chan<- error)

// JobManager manages multiple concurrent jobs and handles their errors. Can wait for all jobs and error handling to finish.
type JobManager struct {
	wgJobs  *sync.WaitGroup
	wgErr   *sync.WaitGroup
	errChan chan error
}

// NewMgr returns a new JobManager
func NewMgr(errHandler func(err error)) *JobManager {
	wgErr := &sync.WaitGroup{}
	wgErr.Add(1)
	errChan := make(chan error, 1000)
	go func() {
		defer wgErr.Done()
		for err := range errChan {
			errHandler(err)
		}
	}()

	return &JobManager{
		wgErr:   wgErr,
		errChan: errChan,
		wgJobs:  &sync.WaitGroup{},
	}
}

// AddJobs calls AddJob for each of the given jobs
func (mgr *JobManager) AddJobs(jobs ...Job) {
	for _, j := range jobs {
		mgr.AddJob(j)
	}
}

// AddJob spawns a new goroutine for the given job, manages its lifetime and handles its errors
func (mgr *JobManager) AddJob(j Job) {
	mgr.wgJobs.Add(1)
	go func() {
		defer mgr.wgJobs.Done()
		j(mgr.errChan)
	}()
}

// Wait blocks until all jobs and error handling have finished
func (mgr *JobManager) Wait() {
	mgr.wgJobs.Wait()
	close(mgr.errChan)
	mgr.wgErr.Wait()
}
