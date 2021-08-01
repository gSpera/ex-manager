package ex

import (
	"time"
)

// Flag contains all the information about a flag
type Flag struct {
	ServiceName string
	ExploitName string
	Value       FlagValue
	From        Target
	Status      SubmittedFlagStatus
	TakenAt     time.Time
	SubmittedAt time.Time
	ExecutionID ExecutionID
}

// SubmittedFlagStatus contains the response of the flag
type SubmittedFlagStatus = string

const (
	// FlagNotSubmitted is A flag not yet submitted, we now nothing about it's state
	FlagNotSubmitted SubmittedFlagStatus = "NOT-SUBMITTED"
	// FlagSubmittedSuccessfully is a flag that has been submitted correctly
	FlagSubmittedSuccesfully = "SUCCESS"
	// FlagExpired is a flag that has expired
	FlagExpired = "EXPIRED"
	// FlagInvalid is an invalid flag
	FlagInvalid = "INVALID"
	// FlagAlreadySubmitted is a flag that has alredy been submitted by the submitter
	FlagAlreadySubmitted = "ALREADY-SUBMITTED"
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
	case FlagAlreadySubmitted:
		return FlagAlreadySubmitted, true
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

func (s *Session) SearchFlagsInText(str string) []string {
	return s.flagRegex.FindAllString(str, -1)
}
