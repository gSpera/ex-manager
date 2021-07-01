package ex

import (
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

const defaultSubmitTime = 1 * time.Minute

type Submitter struct {
	cmdLine string
	state   ExploitState
	// time is read-only, it is only used for referencing
	// it may be changed with ticker
	time               time.Duration
	log                *log.Entry
	ticker             *time.Ticker
	flagStore          FlagStore
	limitForEachSubmit int
}

func NewSubmitter(cmd string, submitEvery time.Duration, log *log.Entry, limit int, flagStore FlagStore) *Submitter {
	s := &Submitter{}
	s.cmdLine = cmd
	s.time = submitEvery
	s.ticker = time.NewTicker(submitEvery)
	s.limitForEachSubmit = limit
	s.flagStore = flagStore
	s.log = log
	s.state = Running
	return s
}

// Submit sends the flags
func (s *Submitter) Submit() {
	flags, err := s.flagStore.GetValueToSubmit(s.limitForEachSubmit)
	if err != nil {
		s.log.Errorln("Cannot retrieve flags:", err)
		return
	}
	// support partial sending??
	if len(flags) == 0 {
		s.log.Println("No flags")
		return
	}

	s.log.Println("Sending flags:", flags)

	cmd := exec.Command(s.cmdLine, flags...)
	cmd.Stdout = s.log.Writer()
	cmd.Stderr = s.log.Writer()
	err = cmd.Run()
	if err != nil {
		s.log.Errorln("Cannot send flags:", err)
		return
	}

	for _, f := range flags {
		s.flagStore.UpdateState(f, FlagSubmittedSuccesfully)
	}
}
