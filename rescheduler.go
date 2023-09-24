// Package rescheduler implements a mini scheduler to ensure an update activity
// has run after all calls to the Run() function.
//
// This results in long-running synchronous tasks collecting a queue of updates
// which will be merged into a single run of the call function.
//
// r.Run() - starts the first runner
// r.Run() - flags a rerun
// r.Run() - flags a rerun
//
// Once the first runner finishes it will run once more due to a rerun being
// requested.
package rescheduler

import "sync"

const (
	running byte = 0b01
	rerun   byte = 0b10
)

// NewRescheduler creates a new rescheduler to run the call function
func NewRescheduler(call func()) *Rescheduler {
	return &Rescheduler{
		lock: &sync.Mutex{},
		me:   0,
		call: call,
		done: makeClosedChannel(),
	}
}

func makeClosedChannel() chan struct{} {
	a := make(chan struct{}, 1)
	close(a)
	return a
}

// Rescheduler handles the running of synchronous tasks
type Rescheduler struct {
	lock *sync.Mutex
	me   byte
	call func()
	done chan struct{}
}

// Run starts threadRun() if it isn't running or sets the rerun flag
func (r *Rescheduler) Run() {
	r.lock.Lock()
	// check running state
	if r.me&running == 1 {
		// set rerun flag
		r.me |= rerun
		r.lock.Unlock()
		return
	}

	// set to running + no rerun
	r.me = running
	r.done = make(chan struct{}, 1)
	r.lock.Unlock()

	// run background thread
	go r.threadRun()
}

// threadRun starts in a goroutine and calls the internal call() field multiple
// times. After running call() the rerun flag is checked. If it is false then
// the running flag is cleared, the done channel is closed then reopened to
// reuse, then breaks out of the loop. If the rerun flag is true then the
// rerun flag is flipped and the internal call() field gets called again.
func (r *Rescheduler) threadRun() {
	for {
		// run call
		r.call()

		// check if a rerun is required and reuse this thread
		r.lock.Lock()
		if r.me&rerun == 0 {
			// clear the run flag
			r.me = 0
			// close r.done to release waiting code and make a new done channel
			close(r.done)
			r.lock.Unlock()
			break
		}
		// flip the rerun flag
		r.me ^= rerun
		r.lock.Unlock()
	}
}

// Wait holds the goroutine until the last call is run (including reruns).
func (r *Rescheduler) Wait() {
	<-r.done
}
