package ex

// Service rapresent a vulnerable service, contains the exploits
type Service struct {
	name     string
	session  *Session
	exploits []*Exploit
}

func (s *Service) AddExploit(e *Exploit) {
	e.service = s
	s.exploits = append(s.exploits, e)
}

func NewService(name string) *Service {
	service := &Service{}
	service.name = name
	return service
}
