package flower_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cheekybits/is"
	"github.com/matryer/flower"
)

func TestSimple(t *testing.T) {
	is := is.New(t)

	// make a manager
	manager := flower.New()
	is.OK(manager)

	// add a handler
	var calls []*flower.Job
	manager.On("event", func(j *flower.Job) {
		calls = append(calls, j)
	})

	// trigger it with some data
	data := map[string]interface{}{"data": true}
	job := manager.New(data, "event")
	is.OK(job)
	is.OK(job.ID())

	// wait for the job to finish
	job.Wait()
	is.Equal(1, len(calls))
	is.Equal(job, calls[0])
	is.Equal(data, calls[0].Data)

}

func TestPath(t *testing.T) {
	is := is.New(t)

	// make a manager
	manager := flower.New()

	// add three handlers
	var calls []string
	manager.On("one", func(j *flower.Job) {
		calls = append(calls, "one")
	})
	manager.On("two", func(j *flower.Job) {
		calls = append(calls, "two")
	})
	manager.On("three", func(j *flower.Job) {
		calls = append(calls, "three")
	})

	data := map[string]interface{}{"data": true}
	job := manager.New(data, "one", "two", "three")

	is.OK(job)

	job.Wait()
	is.Equal(len(calls), 3)
	is.Equal(calls[0], "one")
	is.Equal(calls[1], "two")
	is.Equal(calls[2], "three")

}

func TestAbort(t *testing.T) {
	is := is.New(t)

	// make a manager
	manager := flower.New()
	is.OK(manager)

	// add a handler
	var ticks []*flower.Job
	manager.On("event", func(j *flower.Job) {
		for {
			ticks = append(ticks, j)
			time.Sleep(100 * time.Millisecond)
			if j.ShouldStop() {
				break
			}
		}
	})

	// trigger it with some data
	data := map[string]interface{}{"data": true}
	job := manager.New(data, "event")
	is.OK(job)
	is.OK(job.ID())

	// tell the job to stop in 100 milliseconds
	go func() {
		time.Sleep(1000 * time.Millisecond)
		job.Abort()
	}()

	// wait for the job to finish
	job.Wait()
	is.Equal(10, len(ticks))

}

func TestAbortByID(t *testing.T) {
	is := is.New(t)

	// make a manager
	manager := flower.New()
	is.OK(manager)

	// add a handler
	var ticks []*flower.Job
	manager.On("event", func(j *flower.Job) {
		for {
			ticks = append(ticks, j)
			time.Sleep(100 * time.Millisecond)
			if j.ShouldStop() {
				break
			}
		}
	})

	// trigger it with some data
	data := map[string]interface{}{"data": true}
	job := manager.New(data, "event")
	is.OK(job)
	is.OK(job.ID())

	// look-up the job
	lookedupJob, ok := manager.Get(job.ID())
	is.Equal(true, ok)
	is.Equal(lookedupJob, job)

	// tell the job to stop in 100 milliseconds
	go func() {
		time.Sleep(1000 * time.Millisecond)
		lookedupJob.Abort()
	}()

	// wait for the job to finish
	job.Wait()
	is.Equal(10, len(ticks))

	_, ok = manager.Get(job.ID())
	is.Equal(false, ok)

}

func TestErrs(t *testing.T) {
	is := is.New(t)

	// make a manager
	manager := flower.New()
	is.OK(manager)

	// add a handler
	err := errors.New("something went wrong")
	manager.On("event", func(j *flower.Job) {
		j.Err = err
	})

	// trigger it with some data
	data := map[string]interface{}{"data": true}
	job := manager.New(data, "event")
	is.OK(job)
	is.OK(job.ID())

	// wait for the job to finish
	job.Wait()
	is.Equal(err, job.Err)

}

func TestAll(t *testing.T) {
	is := is.New(t)

	// make a manager
	manager := flower.New()
	is.OK(manager)

	manager.On("event", func(j *flower.Job) {
		for {
			if j.ShouldStop() {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	})

	job1 := manager.New(1, "event")
	job2 := manager.New(2, "event")
	job3 := manager.New(3, "event")

	jobs := manager.All()
	is.Equal(len(jobs), 3)

	job1.Abort()
	job2.Abort()
	job3.Abort()
	job1.Wait()
	job2.Wait()
	job3.Wait()

}

func TestAfter(t *testing.T) {
	is := is.New(t)

	// make a manager
	manager := flower.New()
	manager.PostJobWait = 500 * time.Millisecond
	is.OK(manager)

	manager.On("event", func(j *flower.Job) {
		for {
			if j.ShouldStop() {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	})

	job1 := manager.New(1, "event")
	job2 := manager.New(2, "event")
	job3 := manager.New(3, "event")

	jobs := manager.All()
	is.Equal(len(jobs), 3)

	is.Equal(flower.JobRunning, job1.State())
	is.Equal(flower.JobRunning, job2.State())
	is.Equal(flower.JobRunning, job3.State())

	job1.Abort()
	job2.Abort()
	job3.Abort()
	job1.Wait()
	job2.Wait()
	job3.Wait()

	time.Sleep(250 * time.Millisecond)

	jobs = manager.All()
	is.Equal(len(jobs), 3) // should still be around
	is.Equal(flower.JobFinished, job1.State())
	is.Equal(flower.JobFinished, job2.State())
	is.Equal(flower.JobFinished, job3.State())

	time.Sleep(550 * time.Millisecond)

	jobs = manager.All()
	is.Equal(len(jobs), 0)

}

func TestStateStrings(t *testing.T) {
	is := is.New(t)

	is.Equal(flower.JobScheduled.String(), "scheduled")
	is.Equal(flower.JobRunning.String(), "running")
	is.Equal(flower.JobFinished.String(), "finished")

}
