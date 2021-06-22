package main

import (
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseFiles("template/index.html"))

func handleHome(s *Server, rw http.ResponseWriter, r *http.Request) {
	var templates = template.Must(template.ParseFiles("template/index.html"))

	err := templates.ExecuteTemplate(rw, "index.html", s.Value())
	if err != nil {
		s.log.Errorln("Render index.hmtl:", err)
	}
}
