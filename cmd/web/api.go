package main

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
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
		log.Errorf("API /api/targets:%s", err)
	}
}
