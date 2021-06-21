package ex

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

// Session is the main struct of ex-manager
// a session contains multiple services with their exploits
type Session struct {
	name      string
	log       *log.Entry
	ctx       context.Context
	sleepTime time.Duration
	cancel    context.CancelFunc
	targets   []string
	flagRegex *regexp.Regexp
	services  []*Service
}

func NewSession(name string, flagRegex string, targets ...Target) (*Session, error) {
	s := &Session{}
	var err error

	s.name = name
	s.sleepTime = time.Second
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.targets = targets
	s.log = log.New().WithField("session", name)
	s.services = []*Service{}

	s.flagRegex, err = regexp.Compile(flagRegex)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func NewSessionFromFile(fl string) (*Session, error) {
	bd, err := ioutil.ReadFile(fl)
	if err != nil {
		return &Session{}, fmt.Errorf("Cannot read: %w", err)
	}

	s := &Session{}
	json.Unmarshal(bd, s)
	return s, nil
}

func (s Session) Name() string {
	return s.name
}

func (s Session) ListTargets() []string {
	return s.targets
}

func (s *Session) ListServices() []*Service {
	return s.services
}

func (s *Session) Work() error {
	for {
		if err := s.ctx.Err(); err != nil {
			return err
		}

		e, ok := s.getExploit()
		if !ok {
			s.log.Warnln("Cannot find exploit")
			time.Sleep(1 * time.Second)
			continue
		}

		e.Execute()
		time.Sleep(s.sleepTime)
	}
}

func (s *Session) WorkAdd(n int) {
	for i := 0; i < n; i++ {
		go s.Work()
	}
}

func (s *Session) AddTarget(t Target) {
	s.targets = append(s.targets, t)
}

func (s *Session) AddService(service *Service) {
	service.session = s
	s.services = append(s.services, service)
}

func (s *Session) GetServiceByName(name string) *Service {
	for _, service := range s.services {
		if service.name == name {
			return service
		}
	}

	return nil
}

func (s *Session) Stop() {
	s.cancel()
}

func (s *Session) getExploit() (*Exploit, bool) {
	if len(s.services) == 0 {
		return nil, false
	}

	n := rand.Intn(len(s.services))
	es := s.services[n]
	if len(es.exploits) == 0 {
		// a loop??
		return nil, false
	}
	n = rand.Intn(len(es.exploits))

	if es.exploits[n].state != Running {
		return nil, false
	}

	return es.exploits[n], true
}

func (s *Session) randomTarget() string {
	if len(s.targets) == 0 {
		panic("No targets")
	}
	return s.targets[rand.Intn(len(s.targets))]
}
