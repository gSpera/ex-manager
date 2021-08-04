package ex

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// Session is one of the main structs of ex-manager, it manages
// the Services, the Submitter, the FlagStore and the Workers
//
// A new Session can be built by unmarshaling or using NewSession.
type Session struct {
	name          string
	log           *log.Entry
	ctx           context.Context
	sleepTime     time.Duration
	cancel        context.CancelFunc
	flagRegex     *regexp.Regexp
	services      []*Service
	servicesMutex sync.Mutex
	submitter     *Submitter

	targets      []string
	targetsMutex sync.Mutex

	// Workers informations
	workers      []*WorkerInfo
	workersId    int64
	workersMutex sync.Mutex

	flags  FlagStore
	dumper ExecutionDumper
}

// NewSession creates a new Session and it's internal submitter
func NewSession(name string, flagRegex string, submitCommand string, submitLimit int, flagStore FlagStore, dumper ExecutionDumper, targets ...Target) (*Session, error) {
	s := &Session{}
	var err error

	s.name = name
	s.sleepTime = time.Second
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.targets = targets
	s.log = log.New().WithField("session", name)
	s.services = []*Service{}
	s.submitter = NewSubmitter(submitCommand, defaultSubmitTime, s.log.WithField("what", "submitter"), submitLimit, flagRegex, flagStore)

	s.flagRegex, err = regexp.Compile(flagRegex)
	if err != nil {
		return nil, err
	}

	s.flags = flagStore
	s.dumper = dumper

	return s, nil
}

func NewSessionFromFile(fl string) (*Session, error) {
	bd, err := ioutil.ReadFile(fl)
	if err != nil {
		return &Session{}, fmt.Errorf("Cannot read: %w", err)
	}

	s := &Session{}
	err = json.Unmarshal(bd, s)
	if err != nil {
		return &Session{}, err
	}
	return s, nil
}

func (s *Session) Name() string {
	return s.name
}

func (s *Session) ListTargets() []string {
	return s.targets
}

func (s *Session) ListServices() []*Service {
	return s.services
}

func (s *Session) newWorkerId() int64 {
	return atomic.AddInt64(&s.workersId, 1)
}

// getWorkerKit creates necessary resources for a new worker
func (s *Session) getWorkerKit() (int64, context.Context) {
	// maybe use a context??
	// ctx, close := context.WithCancel(s.ctx)
	id := s.newWorkerId()

	return id, s.ctx
}

func (s *Session) addWorker(w *WorkerInfo) {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	s.workers = append(s.workers, w)
}

func (s *Session) WorkersInfo() []*WorkerInfo {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	return s.workers
}

// Work adds current gorutine to the working pool,
// this functions return when the gorutine stops working
func (s *Session) Work() error {
	w := NewWorkerForSession(s)
	return w.Work()
}

func (s *Session) WorkAdd(n int) {
	for i := 0; i < n; i++ {
		go func() {
			log.Error(s.Work())
		}()
	}
}

func (s *Session) AddTarget(t Target) {
	s.targetsMutex.Lock()
	defer s.targetsMutex.Unlock()

	s.targets = append(s.targets, t)
}

func (s *Session) AddService(service *Service) {
	s.servicesMutex.Lock()
	defer s.servicesMutex.Unlock()

	if s.GetServiceByName(service.name) != nil {
		s.log.Warnln("Cannot add service:", service.name, "another service with same name")
		return
	}

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

func (s *Session) AllExploits() []*Exploit {
	s.servicesMutex.Lock()
	defer s.servicesMutex.Unlock()

	exs := make([]*Exploit, 0, len(s.services)*2)
	for _, service := range s.ListServices() {
		for _, exploit := range service.exploits {
			exs = append(exs, exploit)
		}
	}

	return exs
}

func (s *Session) WorkSubmitter(ctxDone <-chan struct{}) {
	for {
		select {
		case <-s.submitter.ticker.C:
			s.submitter.Submit()
		case <-ctxDone:
			return
		}
	}
}

func (s *Session) getExploit() (*Exploit, bool) {
	if len(s.services) == 0 {
		return nil, false
	}

	exs := s.AllExploits()
	exsRunnable := make([]*Exploit, 0, len(exs))
	for _, ex := range exs {
		if ex.state != Runnable {
			continue
		}

		exsRunnable = append(exsRunnable, ex)
	}

	if len(exsRunnable) == 0 {
		// no runnable exploit
		return nil, false
	}

	n := rand.Intn(len(exsRunnable))
	es := exsRunnable[n]

	return es, true
}

func (s *Session) randomTarget() string {
	if len(s.targets) == 0 {
		panic("No targets")
	}
	return s.targets[rand.Intn(len(s.targets))]
}

func (s *Session) AddFlags(flags ...Flag) error {
	// Set TakenAt time
	timeNow := time.Now()
	for i := range flags {
		flags[i].TakenAt = timeNow
	}

	return s.flags.Put(flags...)
}

// Wrappers for FlagStore

func (s *Session) FlagsByExploitName(serviceName string, exploitName string) ([]Flag, error) {
	return s.flags.GetByName(serviceName, exploitName)
}
func (s *Session) GetFlagsSubmittedDuring(from time.Time, to time.Time) ([]Flag, error) {
	return s.flags.GetFlagsSubmittedDuring(from, to)
}

// Wrappers for ExecutionDumperStore

func (s *Session) LogForExeuctionId(execID ExecutionID) ([]ExecutionLog, error) {
	return s.dumper.LogsFromExecID(execID)
}

func (s *Session) LatestExecIDTimeFromServiceExploitTarget(serviceName string, exploitName string, target Target) (ExecutionID, time.Time, bool, error) {
	return s.dumper.LatestExecIDTimeFromServiceExploitTarget(serviceName, exploitName, target)
}
