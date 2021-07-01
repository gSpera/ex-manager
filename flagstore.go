package ex

import (
	"encoding/json"
)

// FlagStore stores flags for the session
type FlagStore interface {
	Put(...Flag) error
	GetByName(serviceName string, exploitName string) ([]Flag, error)

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

func (m *MemoryFlagStore) MarshalJSON() ([]byte, error) { return json.Marshal((*[]Flag)(m)) }
func (m *MemoryFlagStore) UnmarshalJSON(body []byte) error {
	return json.Unmarshal(body, (*[]Flag)(m))
}

var registeredFlagStores map[string]func() FlagStore = map[string]func() FlagStore{}

func RegisterFlagStore(fl func() FlagStore) {
	registeredFlagStores[fl().name()] = fl
}

func init() {
	RegisterFlagStore(func() FlagStore { return new(MemoryFlagStore) })
}
