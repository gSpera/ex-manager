package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gSpera/ex-manager"
)

func serverHandler(s *Server, fn func(*Server, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		fn(s, rw, r)
	}
}
func handleApiName(s *Server, rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(s.session.Name()))
}

func handleApiTarget(s *Server, rw http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(rw).Encode(s.session.ListTargets())
	if err != nil {
		s.log.Errorln("API /targets:", err)
	}
}

func handleApiNewService(s *Server, rw http.ResponseWriter, r *http.Request) {
	name := r.FormValue("serviceName")
	s.log.Debugf("Adding service: %q\n", name)

	if name == "" {
		http.Error(rw, "{\"ok\": false}", http.StatusBadRequest)
		return
	}

	service := ex.NewService(name)
	s.session.AddService(service)
	fmt.Fprint(rw, "{\"ok\": true}")
}

func handleApiExploitStart(s *Server, rw http.ResponseWriter, r *http.Request) {
	exploitName := r.FormValue("exploitName")
	serviceName := r.FormValue("serviceName")
	service := s.session.GetServiceByName(serviceName)
	if serviceName == "" || exploitName == "" {
		http.Error(rw, "{\"ok\": false}", http.StatusBadRequest)
		return
	}

	if service == nil {
		http.Error(rw, "{\"ok\": false}", http.StatusNotFound)
		return
	}

	for _, e := range service.Exploits() {
		if e.Name() == exploitName {
			e.NewState(ex.Running)
			fmt.Fprint(rw, "{\"ok\": true}")
			return
		}
	}
}

func handleApiFlags(s *Server, rw http.ResponseWriter, r *http.Request) {
	// serviceName := r.FormValue("serviceName")
	// exploitName := r.FormValue("exploitName")
	list := make(map[string]map[ex.Target][]ex.Flag)
	for _, service := range s.session.ListServices() {
		for _, e := range service.Exploits() {
			list[service.Name() + "-" +e.Name()] = e.Flags()
		}
	}

	e := json.NewEncoder(rw)
	e.SetIndent("", "\t")
	err := e.Encode(list)

	if err != nil {
		s.log.Errorf("Encode flags:", err)
	}
}
