package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

func handleApiExploitSetState(s *Server, rw http.ResponseWriter, r *http.Request) {
	exploitName := r.FormValue("exploit")
	serviceName := r.FormValue("service")
	state, ok := ex.ExploitStateFromString(r.FormValue("state"))
	if !ok {
		http.Error(rw, "{\"ok\": false}", http.StatusBadRequest)
		return
	}

	service := s.session.GetServiceByName(serviceName)
	if serviceName == "" || exploitName == "" {
		http.Error(rw, "{\"ok\": false}", http.StatusBadRequest)
		return
	}

	if service == nil {
		http.Error(rw, "{\"ok\": false}", http.StatusNotFound)
		return
	}

	e := service.GetExploitByName(exploitName)
	if e == nil {
		http.Error(rw, "{\"ok\": false}", http.StatusNotFound)
		return
	}

	e.NewState(state)
	fmt.Fprint(rw, "{\"ok\": true}")
}

func handleApiSessionStatus(s *Server, rw http.ResponseWriter, r *http.Request) {
	m := struct {
		Name     string
		Services []string
	}{}

	m.Name = s.session.Name()
	m.Services = []string{}

	for _, service := range s.session.ListServices() {
		m.Services = append(m.Services, service.Name())
	}

	json.NewEncoder(rw).Encode(m)
}

func handleApiServiceStatus(s *Server, rw http.ResponseWriter, r *http.Request) {
	serviceName := r.FormValue("service")
	service := s.session.GetServiceByName(serviceName)
	if service == nil {
		http.Error(rw, "No service", http.StatusNotFound)
		return
	}

	m := struct {
		Exploits []string
	}{}
	m.Exploits = []string{}

	for _, e := range service.Exploits() {
		m.Exploits = append(m.Exploits, e.Name())
	}

	json.NewEncoder(rw).Encode(m)
}

func handleApiExploitStatus(s *Server, rw http.ResponseWriter, r *http.Request) {
	serviceName := r.FormValue("service")
	exploitName := r.FormValue("exploit")

	service := s.session.GetServiceByName(serviceName)
	if service == nil {
		http.Error(rw, "No service", http.StatusNotFound)
		return
	}

	exploit := service.GetExploitByName(exploitName)
	if exploit == nil {
		http.Error(rw, "No exploit", http.StatusNotFound)
		return
	}

	type targetStruct struct {
		Name  string
		Flags []ex.Flag
		State ex.ExploitState
		Fixed bool
	}

	m := struct {
		Targets []targetStruct
		State   ex.ExploitState
	}{}

	m.Targets = []targetStruct{}
	m.State = exploit.CurrentStateString()

	for target, flags := range exploit.Flags() {
		t := targetStruct{}
		t.Name = target
		t.Flags = flags
		t.Fixed = true
		m.Targets = append(m.Targets, t)
	}

	json.NewEncoder(rw).Encode(m)
}

func handleApiFlags(s *Server, rw http.ResponseWriter, r *http.Request) {
	// serviceName := r.FormValue("serviceName")
	// exploitName := r.FormValue("exploitName")
	list := make(map[string]map[ex.Target][]ex.Flag)
	for _, service := range s.session.ListServices() {
		for _, e := range service.Exploits() {
			list[service.Name()+"-"+e.Name()] = e.Flags()
		}
	}

	e := json.NewEncoder(rw)
	e.SetIndent("", "\t")
	err := e.Encode(list)

	if err != nil {
		s.log.Errorf("Encode flags:", err)
	}
}

func handleApiUploadExploit(s *Server, rw http.ResponseWriter, r *http.Request) {
	serviceName := r.FormValue("service")
	exploitName := r.FormValue("exploit")
	cmdName := r.FormValue("cmd")
	var service *ex.Service
	var exploit *ex.Exploit

	res := struct {
		Ok     bool
		Reason string
	}{}

	if strings.TrimSpace(exploitName) == "" {
		res.Ok = false
		res.Reason = "invalid name"
		goto done
	}

	if strings.TrimSpace(cmdName) == "" {
		res.Ok = false
		res.Reason = "no command"
		goto done
	}

	service = s.session.GetServiceByName(serviceName)
	if service == nil {
		res.Ok = false
		res.Reason = "cannot find service"
		goto done
	}

	if service.GetExploitByName(exploitName) != nil {
		res.Ok = false
		res.Reason = "not unique name"
		goto done
	}

	exploit = ex.NewExploit(exploitName, cmdName)
	service.AddExploit(exploit)

	res.Ok = true
	res.Reason = "done"
	goto done

done:
	err := json.NewEncoder(rw).Encode(res)
	if err != nil {
		s.log.Errorln("Cannot encode json:", err)
	}
}
