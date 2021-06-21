package main

import (
	"net/http"

	"github.com/gSpera/ex-manager"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	session *ex.Session
	log     *log.Logger
}
type ServerValue struct {
	SessionName string
	Services    []ServiceValue
}
type ServiceValue struct {
	Name     string
	Exploits []ExploitValue
}
type ExploitValue struct {
	Name  string
	State string
}

func (s *Server) Value() ServerValue {
	return ServerValue{
		SessionName: s.session.Name(),
		Services:    services(s),
	}
}

func services(s *Server) []ServiceValue {
	services := make([]ServiceValue, len(s.session.ListServices()))
	for i, service := range s.session.ListServices() {
		services[i] = ServiceValue{
			Name:     service.Name(),
			Exploits: exploit(service),
		}
	}
	return services
}

func exploit(s *ex.Service) []ExploitValue {
	exploits := make([]ExploitValue, len(s.Exploits()))

	for i, v := range s.Exploits() {
		exploits[i] = ExploitValue{Name: v.Name(), State: v.CurrentStateString()}
	}

	return exploits
}

func main() {
	exs, err := ex.NewSession("Test", "CCIT", "0", "127")
	if err != nil {
		return
	}
	biomarkt := ex.NewService("Biomarkt")
	biomarkt.AddExploit(ex.NewExploit("ExploitName", "cmd"))
	ilbonus := ex.NewService("Il Bonus")
	ilbonus.AddExploit(ex.NewExploit("Il bonus", "ilbonus"))
	exs.AddService(biomarkt)
	exs.AddService(ilbonus)
	s := &Server{
		session: exs,
		log:     log.New(),
	}

	http.HandleFunc("/", serverHandler(s, handleHome))
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/api/targets", serverHandler(s, handleApiTarget))
	http.HandleFunc("/api/name", serverHandler(s, handleApiName))

	log.Infoln("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
