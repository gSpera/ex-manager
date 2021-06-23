package main

import (
	"encoding/json"
	"net/http"
	"os"

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
	exs, err := ex.NewSessionFromFile("ex.json")
	if err != nil {
		log.Fatalln("Cannot create session:", err)
		return
	}

	exs.WorkAdd(10)
	go exs.WorkSubmitter()

	data, err := json.MarshalIndent(exs, "", "\t")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("ex.json", data, 0666)
	if err != nil {
		panic(err)
	}

	s := &Server{
		session: exs,
		log:     log.New(),
	}

	// http.HandleFunc("/", serverHandler(s, handleHome))
	http.Handle("/", http.FileServer(http.Dir("cmd/web/asset")))
	http.HandleFunc("/targets", serverHandler(s, handleApiTarget))
	http.HandleFunc("/api/newService", serverHandler(s, handleApiNewService))
	http.HandleFunc("/api/exploitChangeState", serverHandler(s, handleApiExploitSetState))
	http.HandleFunc("/api/sessionStatus", serverHandler(s, handleApiSessionStatus))
	http.HandleFunc("/flags", serverHandler(s, handleApiFlags))
	http.HandleFunc("/api/serviceStatus", serverHandler(s, handleApiServiceStatus))
	http.HandleFunc("/api/exploitStatus", serverHandler(s, handleApiExploitStatus))
	http.HandleFunc("/api/name", serverHandler(s, handleApiName))

	log.Infoln("Listening on :8080")
	err = http.ListenAndServe(":8080", nil)
	log.Fatalln("Listening ", err)
}
