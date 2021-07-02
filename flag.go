package ex

import (
	"bufio"
	"io"

	log "github.com/sirupsen/logrus"
)

// SubmittedFlagStatus contains the response of the flag
type SubmittedFlagStatus = string

const (
	// FlagNotSubmitted is A flag not yet submitted, we now nothing about it's state
	FlagNotSubmitted SubmittedFlagStatus = "NOT SUBMITTED"
	// FlagSubmittedSuccessfully is a flag that has been submitted correctly
	FlagSubmittedSuccesfully = "SUCCESS"
	// FlagExpired is a flag that has expired
	FlagExpired = "EXPIRED"
	// FlagInvalid is an invalid flag
	FlagInvalid = "INVALID"
	// FlagAlreadySubmitted is a flag that has alredy been submitted by the submitter
	FlagAlredySubmitted = "ALREDY-SUBMITTED"
	// Flagown is a flag that was obtained by the submitter team
	FlagOwn = "TEAM-OWN"
	// FlagNOP is a flag obtained by the NOP team
	FlagNOP = "TEAM-NOP"
	// FlagOffline is a flag that has been submitted when the ctf was closed
	FlagOffline = "OFFLINE-CTF"
	// FlaggServiceOffline is a flag that has been submitted when the service was offline
	FlagServiceOffline = "OFFLINE-SERVICE"
)

func SubmittedFlagStatusFromString(value string) (SubmittedFlagStatus, bool) {
	switch value {
	case FlagNotSubmitted:
		return FlagNotSubmitted, true
	case FlagSubmittedSuccesfully:
		return FlagSubmittedSuccesfully, true
	case FlagExpired:
		return FlagExpired, true
	case FlagInvalid:
		return FlagInvalid, true
	case FlagAlredySubmitted:
		return FlagAlredySubmitted, true
	case FlagOwn:
		return FlagOwn, true
	case FlagNOP:
		return FlagNOP, true
	case FlagOffline:
		return FlagOffline, true
	case FlagServiceOffline:
		return FlagServiceOffline, true
	}
	return "", false
}

// FlagRetriveWriter creates a io.Writer, when wrote the content is logged and flags are searched
func FlagRetriveWriter(l *log.Entry, t Target, e *Exploit) io.Writer {
	pr, pw := io.Pipe()

	go func() {
		r := bufio.NewReader(pr)
		for {
			line, err := r.ReadBytes('\n')
			if err == io.EOF {
				return
			}

			if err != nil {
				log.Error("Cannot read:", err)
				return
			}

			f := e.service.session.SearchFlagsInText(string(line))
			e.foundFlag(t, f...)

			l.Println("Program Stdout:", string(line))
			if len(f) > 0 {
				l.Println("Found Flags: ", f)
			}
		}
	}()

	return pw
}

func (s *Session) SearchFlagsInText(str string) []string {
	return s.flagRegex.FindAllString(str, -1)
}
