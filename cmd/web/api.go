package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/gSpera/ex-manager"
)

func serverHandler(s *Server, fn func(*Server, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		fn(s, rw, r)
	}
}
func handleApiName(s *Server, rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(s.Session.Name()))
}

func handleApiTarget(s *Server, rw http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(rw).Encode(s.Session.ListTargets())
	if err != nil {
		s.log.Errorln("API /targets:", err)
	}
}

func handleApiNewService(s *Server, rw http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	s.log.Debugf("Adding service: %q\n", name)

	if name == "" {
		http.Error(rw, "{\"Ok\": false}", http.StatusBadRequest)
		return
	}

	if s.Session.GetServiceByName(name) != nil {
		http.Error(rw, "{\"Ok\": false}", http.StatusBadRequest)
		return
	}

	service := ex.NewService(name)
	s.Session.AddService(service)
	fmt.Fprint(rw, "{\"Ok\": true}")
}

func handleApiExploitSetState(s *Server, rw http.ResponseWriter, r *http.Request) {
	exploitName := r.FormValue("exploit")
	serviceName := r.FormValue("service")
	state, ok := ex.ExploitStateFromString(r.FormValue("state"))
	if !ok {
		http.Error(rw, "{\"ok\": false}", http.StatusBadRequest)
		return
	}

	service := s.Session.GetServiceByName(serviceName)
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

	m.Name = s.Session.Name()
	m.Services = []string{}

	for _, service := range s.Session.ListServices() {
		m.Services = append(m.Services, service.Name())
	}

	json.NewEncoder(rw).Encode(m)
}

func handleApiServiceStatus(s *Server, rw http.ResponseWriter, r *http.Request) {
	serviceName := r.FormValue("service")
	service := s.Session.GetServiceByName(serviceName)
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

	service := s.Session.GetServiceByName(serviceName)
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
		Fixed bool
	}

	m := struct {
		Targets []targetStruct
		State   ex.ExploitState
	}{}

	targets := s.Session.ListTargets()
	m.Targets = make([]targetStruct, len(targets))
	flags, err := s.Session.FlagsByExploitName(serviceName, exploitName)
	if err != nil {
		s.log.Errorln("Cannot retrieve flags")
		flags = []ex.Flag{}
	}
	for i := range m.Targets {
		m.Targets[i].Name = targets[i]
		m.Targets[i].Fixed = true

		m.Targets[i].Flags = []ex.Flag{}
		for _, f := range flags {
			if f.From == targets[i] {
				m.Targets[i].Flags = append(m.Targets[i].Flags, f)
			}
		}
	}

	m.State = exploit.CurrentStateString()

	json.NewEncoder(rw).Encode(m)
}

func handleApiFlags(s *Server, rw http.ResponseWriter, r *http.Request) {
	// serviceName := r.FormValue("serviceName")
	// exploitName := r.FormValue("exploitName")
	list := make(map[string][]ex.Flag)
	for _, service := range s.Session.ListServices() {
		for _, e := range service.Exploits() {
			flags, err := s.Session.FlagsByExploitName(service.Name(), e.Name())
			if err != nil {
				s.log.Errorln("Cannot get flags:", err)
				continue
			}
			list[service.Name()+"-"+e.Name()] = flags
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
	var filename string
	var directory string

	res := struct {
		Ok     bool
		Reason string
		Name   string
		code   int
	}{}
	res.Ok = false
	res.Name = ""

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		res.Reason = "no content"
		res.code = http.StatusBadRequest
		goto done
	}

	if fileHeader.Size > maxUploadSize {
		res.Reason = "too big"
		res.code = http.StatusRequestEntityTooLarge
		goto done
	}

	if strings.TrimSpace(fileHeader.Filename) == "" {
		res.Reason = "no name"
		res.code = http.StatusBadRequest
		goto done
	}

	if strings.TrimSpace(exploitName) == "" {
		res.Reason = "invalid exploit"
		res.code = http.StatusBadRequest
		goto done
	}

	if strings.TrimSpace(cmdName) == "" {
		res.Reason = "no command"
		res.code = http.StatusBadRequest
		goto done
	}

	service = s.Session.GetServiceByName(serviceName)
	if service == nil {
		res.Reason = "cannot find service"
		res.code = http.StatusNotFound
		goto done
	}

	if service.GetExploitByName(exploitName) != nil {
		res.Reason = "not unique name"
		res.code = http.StatusNotFound
		goto done
	}

	filename, err = s.UploadFile(serviceName, exploitName, fileHeader.Filename, file)

	if err != nil {
		res.Reason = "upload"
		goto done
	}

	directory = path.Dir(filename)
	exploit = ex.NewExploit(exploitName, cmdName, directory)
	service.AddExploit(exploit)

	res.Ok = true
	res.Reason = "done"
	res.Name = filename
	goto done

done:
	err = json.NewEncoder(rw).Encode(res)
	if err != nil {
		s.log.Errorln("Cannot encode json:", err)
	}
}
