package ex

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

func (s *Session) UnmarshalJSON(b []byte) error {
	m := struct {
		Name          string
		SleepTime     time.Duration
		Targets       []Target
		FlagRegex     string
		SubmitCommand string
		SubmitTime    time.Duration
		SubmitLimit   int
		Flags         flagStoreMarshal
		Services      []*Service
	}{}

	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	s.name = m.Name
	if s.name == "" {
		return fmt.Errorf("No session name")
	}
	s.sleepTime = m.SleepTime
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.log = log.New().WithField("session", m.Name)
	if m.SubmitTime == 0 {
		m.SubmitTime = defaultSubmitTime
	}
	if m.SubmitCommand == "" {
		return fmt.Errorf("No submit command")
	}
	s.targets = m.Targets
	if m.FlagRegex == "" {
		return fmt.Errorf("No regex flag")
	}
	s.flagRegex, err = regexp.Compile(m.FlagRegex)
	if err != nil {
		return fmt.Errorf("Decode regex: %w", err)
	}
	for _, mm := range m.Services {
		s.AddService(mm)
	}

	s.flags = m.Flags.FlagStore
	s.submitter = NewSubmitter(m.SubmitCommand, m.SubmitTime*time.Second, s.log.WithField("what", "submitter"), m.SubmitLimit, m.FlagRegex, s.flags)

	return nil
}

func (s *Service) UnmarshalJSON(b []byte) error {
	m := struct {
		Name     string
		Exploits []*Exploit
	}{}

	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	s.name = m.Name
	if s.name == "" {
		return fmt.Errorf("No serice name")
	}
	for _, mm := range m.Exploits {
		s.AddExploit(mm)
	}
	return nil
}
func (e *Exploit) UnmarshalJSON(b []byte) error {
	m := struct {
		Name string

		Patched      map[Target]bool
		State        string
		CommandName  string
		ExecutionDir string
	}{}

	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	e.name = m.Name
	if e.name == "" {
		return fmt.Errorf("No exploit name")
	}
	e.cmdName = m.CommandName
	e.executionDir = m.ExecutionDir
	e.state = Paused

	e.state = m.State
	e.ctx, e.stop = context.WithCancel(context.Background())
	e.patched = make(map[string]struct{}, len(m.Patched))
	for target, patched := range m.Patched {
		if patched {
			e.patched[target] = struct{}{}
		}
	}
	return nil
}

func (s *Session) MarshalJSON() ([]byte, error) {
	m := struct {
		Name          string
		Targets       []Target
		FlagRegex     string
		SubmitCommand string
		SubmitTime    time.Duration
		Flags         flagStoreMarshal
		SubmitLimit   int
		Services      []*Service
	}{}

	m.Name = s.name
	m.SubmitCommand = s.submitter.cmdLine
	m.SubmitTime = s.submitter.time / time.Second
	m.FlagRegex = s.flagRegex.String()
	m.Targets = s.targets
	m.Services = s.services
	m.Flags = flagStoreMarshal{s.flags}
	m.SubmitLimit = s.submitter.limitForEachSubmit
	return json.Marshal(m)
}

func (s *Service) MarshalJSON() ([]byte, error) {
	m := struct {
		Name     string
		Exploits []*Exploit
	}{}

	m.Name = s.name
	m.Exploits = s.exploits
	return json.Marshal(m)
}

func (e *Exploit) MarshalJSON() ([]byte, error) {
	m := struct {
		Name         string
		Patched      map[Target]bool `json:",omitempty"`
		State        ExploitState
		CommandName  string
		ExecutionDir string
	}{}

	m.Name = e.name
	m.CommandName = e.cmdName
	m.ExecutionDir = e.executionDir
	m.State = e.state
	return json.Marshal(m)
}

type flagStoreMarshal struct {
	FlagStore
}

func (flagStoreMarshal flagStoreMarshal) MarshalJSON() ([]byte, error) {
	flagStore := flagStoreMarshal.FlagStore

	return json.Marshal(struct {
		Name  string
		Flags FlagStore
	}{Name: flagStore.name(), Flags: flagStore})
}

func (flagStoreMarshal *flagStoreMarshal) UnmarshalJSON(body []byte) error {
	mapp := struct {
		Name  string
		Flags interface{}
	}{}

	err := json.Unmarshal(body, &mapp)
	if err != nil {
		return fmt.Errorf("decode flag store: %w", err)
	}
	if mapp.Name == "" {
		return fmt.Errorf("no Name found")
	}

	flagStore, ok := registeredFlagStores[mapp.Name]
	if !ok {
		return fmt.Errorf("no registered Name")
	}

	content, err := json.Marshal(mapp.Flags)
	if err != nil {
		return fmt.Errorf("Cannot re encode flags: %w", err)
	}

	flag := flagStore()
	err = flag.UnmarshalJSON(content)
	if err != nil {
		return err
	}

	flagStoreMarshal.FlagStore = flag
	return nil
}
