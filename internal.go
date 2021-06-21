package ex

import (
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

var r uint32 = 0

func getRandomID() int {
	return int(atomic.AddUint32(&r, 1))
}

func newLogger(service string, exploit string, id int) *log.Entry {
	l := log.New()
	entry := l.WithFields(log.Fields{
		"service":    service,
		"exploit":    exploit,
		"exploit-id": id,
	})
	return entry
}

func Dump() Session {
	return Session{
		name: "A",
		targets: []string{
			"1", "2",
		},
		services: []*Service{
			{name: "A"},
			{name: "A"},
		},
	}
}
