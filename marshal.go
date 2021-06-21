package ex

import (
	"context"
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

func (s *Session) UnmarshalJSON(b []byte) error {
	m := struct {
		Name      string
		SleepTime time.Duration
		Targets   []Target
		Services  []*Service
	}{}

	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	s.name = m.Name
	s.sleepTime = m.SleepTime
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.log = log.New().WithField("session", m.Name)
	s.targets = m.Targets
	for _, mm := range m.Services {
		s.AddService(mm)
	}
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
	for _, mm := range m.Exploits {
		s.AddExploit(mm)
	}
	return nil
}
func (e *Exploit) UnmarshalJSON(b []byte) error {
	m := struct {
		Name string

		Patched     map[Target]bool
		CommandName string
	}{}

	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	e.name = m.Name
	e.cmdName = m.CommandName
	e.state = Paused

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
		Name     string
		Targets  []Target
		Services []*Service
	}{}

	m.Name = s.name
	m.Targets = s.targets
	m.Services = s.services
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
		Name        string
		Patched     map[Target]bool `json:",omitempty"`
		CommandName string
	}{}

	m.Name = e.name
	m.CommandName = e.cmdName
	return json.Marshal(m)
}
