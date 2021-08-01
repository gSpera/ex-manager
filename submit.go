package ex

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
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
	flagRegex          string
	flagStore          FlagStore
	limitForEachSubmit int
}

func NewSubmitter(cmd string, submitEvery time.Duration, log *log.Entry, limit int, flagRegex string, flagStore FlagStore) *Submitter {
	s := &Submitter{}
	s.cmdLine = cmd
	s.time = submitEvery
	s.ticker = time.NewTicker(submitEvery)
	s.limitForEachSubmit = limit
	s.flagRegex = flagRegex
	s.flagStore = flagStore
	s.log = log
	s.state = Runnable
	return s
}

// Submit sends the flags
func (s *Submitter) Submit() {
	flags, err := s.flagStore.GetValueToSubmit(s.limitForEachSubmit)
	if err != nil {
		s.log.Errorln("Cannot retrieve flags to submit:", err)
		return
	}
	// support partial sending??
	if len(flags) == 0 {
		s.log.Println("No flags")
		return
	}

	s.log.Println("Sending flags:", flags)

	cmdArgs := strings.Fields(s.cmdLine)
	cmdArgs = append(cmdArgs, flags...)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = SubmitterRetrieverWriter(s.flagRegex, s.log, s.log.Writer(), s.flagStore)
	cmd.Stderr = s.log.Writer()
	err = cmd.Run()
	if err != nil {
		s.log.Errorln("Cannot send flags:", err)
		return
	}

	s.log.Println("Done sending flags")
}

func SubmitterRetrieverWriter(flagRegex string, lo *log.Entry, w io.Writer, flagStore FlagStore) io.Writer {
	pr, pw := io.Pipe()

	r := bufio.NewReader(pr)
	re, err := regexp.Compile(fmt.Sprintf("^(%s)\\s([a-zA-Z0-9\\-]+)", flagRegex))
	if err != nil {
		lo.Errorln("Cannot compile regex:", err)
	}
	go func() {
		for {
			line, err := r.ReadString('\n')
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				lo.Errorln("Cannot read line:", err)
				continue
			}
			_, err = w.Write([]byte(line))
			if err != nil {
				lo.Errorln("Cannot write line:", err)
				continue
			}
			founds := re.FindStringSubmatch(line)
			if len(founds) != 3 {
				lo.Errorln("Invalid line: wrong count:", founds, len(founds))
				continue
			}
			flag := founds[1]
			state := founds[2]
			stateValue, ok := SubmittedFlagStatusFromString(state)
			if !ok {
				lo.Errorln("Wrong state:", state, "for found flag:", flag)
				continue
			}
			err = flagStore.UpdateState(flag, stateValue)
			if err != nil {
				lo.Errorln("Cannot update state, flag:", flag, " updated state:", stateValue, ":", err)
				continue
			}

			lo.Println("Submitted flag:", flag, stateValue)
			lo.Debugln("Out:", line)
		}
	}()
	return pw
}
