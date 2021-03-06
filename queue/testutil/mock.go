package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/deciduosity/amboy"
	"github.com/deciduosity/amboy/dependency"
	"github.com/deciduosity/amboy/job"
	"github.com/deciduosity/amboy/registry"
	"github.com/google/uuid"
)

type Counters interface {
	Reset()
	Inc()
	Count() int
}

var MockJobCounters Counters

type mockJobRunEnv struct {
	runCount int
	mu       sync.Mutex
}

func (e *mockJobRunEnv) Inc() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.runCount++
}

func (e *mockJobRunEnv) Count() int {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.runCount
}

func (e *mockJobRunEnv) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.runCount = 0
}

func init() {
	MockJobCounters = &mockJobRunEnv{}
	registry.AddJobType("mock", NewMockJob)
	registry.AddJobType("sleep", func() amboy.Job { return newSleepJob() })
}

//
type mockJob struct {
	job.Base `bson:"job_base" json:"job_base" yaml:"job_base"`
}

func MakeMockJob(id string) amboy.Job {
	j := NewMockJob().(*mockJob)
	j.SetID(id)
	return j
}

func NewMockJob() amboy.Job {
	j := &mockJob{
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    "mock",
				Version: 0,
			},
		},
	}
	j.SetDependency(dependency.NewAlways())
	return j
}

func (j *mockJob) Run(_ context.Context) {
	defer j.MarkComplete()

	MockJobCounters.Inc()
}

type sleepJob struct {
	Sleep time.Duration
	job.Base
}

func NewSleepJob(dur time.Duration) amboy.Job {
	j := newSleepJob()
	j.Sleep = dur
	return j
}

func newSleepJob() *sleepJob {
	j := &sleepJob{
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    "sleep",
				Version: 0,
			},
		},
	}
	j.SetDependency(dependency.NewAlways())
	j.SetID(uuid.New().String())
	return j
}

func (j *sleepJob) Run(ctx context.Context) {
	defer j.MarkComplete()

	if j.Sleep == 0 {
		return
	}

	timer := time.NewTimer(j.Sleep)
	defer timer.Stop()

	select {
	case <-timer.C:
		return
	case <-ctx.Done():
		return
	}
}
