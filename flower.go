package flower

import (
	"sync"
)

const defaultIDLen = 32

type Manager struct {
	IDLen int
	lock  sync.RWMutex
	ons   map[string][]func(j *Job)
	jobs  map[*Job]struct{}
}

func (m *Manager) On(event string, handler func(j *Job)) {
	m.lock.Lock()
	m.ons[event] = append(m.ons[event], handler)
	m.lock.Unlock()
}
func (m *Manager) New(data interface{}, path ...string) (*Job, error) {

	j := &Job{
		Data:     data,
		id:       randomKey(m.IDLen),
		stopChan: make(chan struct{}),
	}

	go func(job *Job) {

		for _, event := range path {
			m.lock.RLock()
			handlers := m.ons[event]
			m.lock.RUnlock()
			for _, handler := range handlers {
				handler(job)
			}
		}

		// job is finished
		job.setFinished()

	}(j)

	return j, nil
}

func New() *Manager {
	return &Manager{
		IDLen: defaultIDLen,
		ons:   make(map[string][]func(j *Job)),
	}
}

type Job struct {
	Data       interface{}
	Err        error
	lock       sync.RWMutex
	id         string
	stopChan   chan struct{}
	finished   bool
	shouldStop bool
}

func (j *Job) setFinished() {
	j.lock.Lock()
	j.finished = true
	j.lock.Unlock()
	close(j.stopChan)
}

func (j *Job) ShouldStop() bool {
	j.lock.RLock()
	s := j.shouldStop
	j.lock.RUnlock()
	return s
}

func (j *Job) Abort() {
	j.lock.Lock()
	j.shouldStop = true
	j.lock.Unlock()
}

func (j *Job) ID() string {
	return j.id
}

func (j *Job) Wait() {
	<-j.stopChan
}
