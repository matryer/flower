package flower

import (
	"sync"
	"time"

	"github.com/stretchr/slog"
)

const idlen = 16

// Manager keeps track of job handlers and jobs.
type Manager struct {
	lock sync.RWMutex
	ons  map[string][]func(j *Job)
	jobs map[string]*Job

	// PostJobWait is the amount of time to wait before
	// cleaning up after a job once it has finished.
	// By default, jobs disappear immediately.
	PostJobWait time.Duration

	// Logger is the slog.Logger where log information is
	// written to.
	Logger slog.Logger
}

// New makes a new Manager.
func New() *Manager {
	return &Manager{
		ons:    make(map[string][]func(j *Job)),
		jobs:   make(map[string]*Job),
		Logger: slog.NilLogger,
	}
}

// On adds an event handler.
func (m *Manager) On(event string, handler func(j *Job)) {
	m.lock.Lock()
	m.ons[event] = append(m.ons[event], handler)
	m.lock.Unlock()
}

// New makes a new job.
func (m *Manager) New(data interface{}, path ...string) *Job {
	randomID := randomKey(idlen)
	j := &Job{
		Data:     data,
		Logger:   m.Logger.New("job(" + randomID + ")"),
		id:       randomID,
		stopChan: make(chan struct{}),
		state:    JobRunning,
	}
	m.lock.Lock()
	m.jobs[j.id] = j
	m.lock.Unlock()
	if j.Logger.Info() {
		j.Logger.Info("new job")
	}
	go func(job *Job) {

		job.lock.Lock()
		job.started = time.Now()
		job.lock.Unlock()

		if j.Logger.Debug() {
			j.Logger.Debug("started:", job.started)
		}

		for _, event := range path {
			m.lock.RLock()
			handlers := m.ons[event]
			m.lock.RUnlock()
			for _, handler := range handlers {
				if j.Logger.Debug() {
					j.Logger.Debug("running handler", handler)
				}
				handler(job)
				if j.Logger.Debug() {
					j.Logger.Debug("handler complete")
				}
			}
		}

		// job is finished
		if j.Logger.Debug() {
			j.Logger.Debug("finished")
		}
		job.setFinished()

		if m.PostJobWait > 0 {
			if j.Logger.Debug() {
				j.Logger.Debug("waiting before cleanup:", m.PostJobWait)
			}
		}

		// wait a while before removing the job
		time.Sleep(m.PostJobWait)

		// remove it
		m.lock.Lock()
		delete(m.jobs, job.id)
		m.lock.Unlock()

		if j.Logger.Debug() {
			j.Logger.Debug("cleaned up")
		}

	}(j)
	return j
}

// RunningLen gets the number of jobs that are
// currently running.
func (m *Manager) RunningLen() int {
	c := 0
	m.lock.RLock()
	for _, job := range m.jobs {
		if job.State() == JobRunning {
			c++
		}
	}
	m.lock.RUnlock()
	return c
}

// All gets all active jobs.
func (m *Manager) All() []*Job {
	jobs := make([]*Job, len(m.jobs))
	i := 0
	m.lock.RLock()
	for _, job := range m.jobs {
		jobs[i] = job
		i++
	}
	m.lock.RUnlock()
	return jobs
}

// Get gets a job by ID.
func (m *Manager) Get(id string) (*Job, bool) {
	m.lock.RLock()
	j, ok := m.jobs[id]
	m.lock.RUnlock()
	return j, ok
}

// Job is a single job that will flow through the
// system.
type Job struct {
	Data       interface{}
	Logger     slog.Logger
	Err        error
	lock       sync.RWMutex
	id         string
	stopChan   chan struct{}
	state      JobState
	shouldStop bool
	started    time.Time
	finished   time.Time
}

// JobState represents the state of a Job.
type JobState int8

const (
	_ JobState = iota
	// JobScheduled means the job has not yet started.
	// Not used since jobs start running right away.
	JobScheduled
	// JobRunning means the job is running.
	JobRunning
	// JobFinished means the job has finished.
	JobFinished
	// JobErred means the job was aborted due to an error.
	JobErred
)

var jobStateStrs = map[JobState]string{
	JobScheduled: "scheduled",
	JobRunning:   "running",
	JobFinished:  "finished",
	JobErred:     "error",
}

// String gets the JobState as a human readable string.
func (s JobState) String() string {
	return jobStateStrs[s]
}

// setFinished marks the job as finished.
func (j *Job) setFinished() {
	j.lock.Lock()
	j.state = JobFinished
	if j.Err != nil {
		j.state = JobErred
	}
	j.finished = time.Now()
	if j.Logger.Debug() {
		j.Logger.Debug("finished at", j.finished, "with state", j.state)
	}
	j.lock.Unlock()
	close(j.stopChan)
}

// State gets the current state of the job.
func (j *Job) State() JobState {
	j.lock.RLock()
	state := j.state
	j.lock.RUnlock()
	return state
}

// ShouldStop gets whether the job should stop running
// or not.
func (j *Job) ShouldStop() bool {
	j.lock.RLock()
	s := j.shouldStop
	j.lock.RUnlock()
	return s
}

// Abort causes the job to stop. Calls to ShouldStop will
// return true after Abort is called and handlers should stop
// running.
func (j *Job) Abort() {
	j.lock.Lock()
	j.shouldStop = true
	j.lock.Unlock()
}

// ID gets the unique ID for the job.
func (j *Job) ID() string {
	return j.id
}

// Wait blocks until the job has finished.
func (j *Job) Wait() {
	<-j.stopChan
}

// Started gets the time.Time when this Job started.
func (j *Job) Started() time.Time {
	j.lock.RLock()
	s := j.started
	j.lock.RUnlock()
	return s
}

// Finished gets the time.Time when this Job finished.
func (j *Job) Finished() time.Time {
	j.lock.RLock()
	s := j.finished
	j.lock.RUnlock()
	return s
}
