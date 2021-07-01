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
	time          time.Duration
	ticker        *time.Ticker
	flagsToSubmit []Flag
}

func NewSubmitter(cmd string, submitEvery time.Duration) *Submitter {
	s := &Submitter{}
	s.cmdLine = cmd
	s.time = submitEvery
	s.ticker = time.NewTicker(submitEvery)
	s.state = Paused

	return s
}

// AddFlags adds to the interal buffer the flag to submit
func (s *Submitter) AddFlags(flags ...Flag) {
	s.flagsToSubmit = append(s.flagsToSubmit, flags...)
}

// Submit sends the flags
func (s *Submitter) Submit() {
	// support partial sending??
	lo := log.New()
	if len(s.flagsToSubmit) == 0 {
		lo.Println("No flags")
		return
	}

	lo.Println("Sending flags:", s.flagsToSubmit)
	flagString := make([]FlagValue, len(s.flagsToSubmit))
	for i := range flagString {
		flagString[i] = s.flagsToSubmit[i].Value
	}

	cmd := exec.Command(s.cmdLine, flagString...)
	cmd.Stdout = lo.Writer()
	cmd.Stderr = lo.Writer()
	err := cmd.Run()
	if err != nil {
		lo.Errorln("Cannot send flags:", err)
		return
	}
	s.flagsToSubmit = []Flag{}
}
