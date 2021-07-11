package ex

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// WorkState is the state in which a Worker can be
type WorkerState int

const (
	WorkerPaused WorkerState = iota
	WorkerSearching
	WorkerRunning
	WorkerSleeping
	WorkerExit
	WorkerDebug
)

func (w WorkerState) String() string {
	switch w {
	case WorkerPaused:
		return "WorkerPaused"
	case WorkerSearching:
		return "WorkerSearching"
	case WorkerRunning:
		return "WorkerRunning"
	case WorkerSleeping:
		return "WorkerSleeping"
	case WorkerExit:
		return "WorkerExit"
	default:
		panic(fmt.Sprintf("WorkerState.String(): Unkown state: %d", w))
	}
}

// WorkerInfo contains the information about a specific worker.
// A worker is bind to a specific Session when it is created, see NewWorkerForSession
type WorkerInfo struct {
	id         int64 //worker id
	session    *Session
	setStateCh chan WorkerState //used for setting state from the Session
	ctx        context.Context
	cancel     context.CancelFunc

	state     WorkerState
	from      time.Time // time of last transition of state
	log       *log.Entry
	sleepTime time.Duration
}

func NewWorkerForSession(s *Session) *WorkerInfo {
	id, ctx := s.getWorkerKit()

	w := &WorkerInfo{
		id:         id,
		state:      WorkerSleeping,
		from:       time.Now(),
		log:        s.log.WithField("worker-id", id),
		ctx:        ctx,
		session:    s,
		setStateCh: make(chan WorkerState),
	}

	s.addWorker(w)
	return w
}

func (w *WorkerInfo) ID() int64 {
	return w.id
}

func (w *WorkerInfo) State() (WorkerState, time.Time) {
	return w.state, w.from
}

// SetState can be used to signal the worker to enter a specific event,
// note that the state will not be changed immediately.
func (w *WorkerInfo) SetState(state WorkerState) {
	w.setStateCh <- state
}

func (w *WorkerInfo) setState(state WorkerState) {
	if state == w.state {
		w.log.WithField("state", w.state).Println("Remaining in state")
		return
	}

	w.log.WithFields(log.Fields{
		"from-state": w.state,
		"to-state":   state,
	}).Println("Updating state to:", state)
	w.state = state
	w.from = time.Now()
}

func (w *WorkerInfo) checkSetStateCh() {
	select {
	case newState := <-w.setStateCh:
		w.setState(newState)
	default:
	}
}

// Work uses the current gorutine to execute exploits for the given Session.
//
// See Session.Work()
func (w *WorkerInfo) Work() error {
	for {
		if err := w.ctx.Err(); err != nil {
			w.SetState(WorkerExit)
			return err
		}

		// check for upcoming state changes
		w.checkSetStateCh()
		switch w.state {
		case WorkerExit:
			w.log.Println("Exiting")
			return nil
		case WorkerPaused:
			w.log.Println("Paused")
			time.Sleep(w.sleepTime)
			continue
		}

		w.log.Println("Searching for a exploit")
		w.setState(WorkerSearching)
		e, ok := w.session.getExploit()
		if !ok {
			w.log.Warnln("Cannot find exploit")
			time.Sleep(1 * time.Second)
			continue
		}
		w.log.Println("Found")

		w.setState(WorkerRunning)
		e.Execute()
		w.setState(WorkerSleeping)
		time.Sleep(w.sleepTime)
	}
}
