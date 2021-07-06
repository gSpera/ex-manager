package ex

import (
	"encoding/json"
	"fmt"
	"time"
)

var registeredFlagStores map[string]func() FlagStore = map[string]func() FlagStore{}

// RegisterFlagStore registers a new FlagStore
func RegisterFlagStore(fl func() FlagStore) {
	registeredFlagStores[fl().name()] = fl
}

// flag store
func init() {
	RegisterFlagStore(func() FlagStore { return new(MemoryFlagStore) })
}

// FlagStore stores flags for the session
type FlagStore interface {
	Put(...Flag) error
	GetByName(serviceName string, exploitName string) ([]Flag, error)
	UpdateState(flagValue string, flagState SubmittedFlagStatus) error
	GetValueToSubmit(limit int) ([]string, error)
	GetFlagsSubmittedDuring(from time.Time, to time.Time) ([]Flag, error)

	name() string
	json.Marshaler
	json.Unmarshaler
}

// MemoryFlagStore implements FlagStore
type MemoryFlagStore []Flag

func (m *MemoryFlagStore) Put(flag ...Flag) error {
	*m = append(*m, flag...)
	return nil
}

func (m *MemoryFlagStore) name() string {
	return "MEMORY"
}
func (m *MemoryFlagStore) GetByName(serviceName string, exploitName string) ([]Flag, error) {
	v := make([]Flag, 0)
	for _, flag := range *m {
		if flag.ServiceName == serviceName && flag.ExploitName == exploitName {
			v = append(v, flag)
		}
	}

	return v, nil
}

func (m *MemoryFlagStore) UpdateState(flag string, state SubmittedFlagStatus) error {
	timeNow := time.Now()

	for i := range *m {
		if (*m)[i].Value == flag {
			(*m)[i].Status = state
			// Update SubmittedAt time
			// this would happen only once
			(*m)[i].SubmittedAt = timeNow
			return nil
		}
	}

	return fmt.Errorf("no flag found")
}

func (m *MemoryFlagStore) GetValueToSubmit(limit int) ([]string, error) {
	res := make([]string, 0, limit)

	for _, f := range *m {
		if f.Status == FlagNotSubmitted {
			res = append(res, f.Value)
			if len(res) == limit {
				break
			}
		}
	}

	return res, nil
}

func (m *MemoryFlagStore) GetFlagsSubmittedDuring(from time.Time, to time.Time) ([]Flag, error) {
	res := make([]Flag, 0)

	for _, f := range *m {
		if f.SubmittedAt.After(from) && f.SubmittedAt.Before(to) {
			res = append(res, f)
		}
	}

	return res, nil
}

func (m *MemoryFlagStore) MarshalJSON() ([]byte, error)    { return json.Marshal((*[]Flag)(m)) }
func (m *MemoryFlagStore) UnmarshalJSON(body []byte) error { return json.Unmarshal(body, (*[]Flag)(m)) }
