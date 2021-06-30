package main

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gSpera/ex-manager"
)

type Server struct {
	Session *ex.Session `json:"ExConfig"`
	Config  struct {
		Address string
	}
	*http.ServeMux `json:"-"`

	Users map[string]string

	log       *log.Logger
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (s *Server) HandleServerFunc(path string, fn func(s *Server, rw http.ResponseWriter, r *http.Request)) {
	wrap := func(rw http.ResponseWriter, r *http.Request) {
		fn(s, rw, r)
	}
	wrap = loginMiddleware(s, wrap)

	s.HandleFunc(
		path,
		wrap,
	)
}
