package main

import (
	"crypto/sha1"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func loginMiddleware(s *Server, fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()

		// No auth provided
		if !ok {
			rw.Header().Add("WWW-Authenticate", "Basic realm=\"Password is required\"")
			http.Error(rw, "Not logged in", http.StatusUnauthorized)
			return
		}

		hash, ok := s.Users[user]
		if !ok {
			s.log.WithFields(logrus.Fields{
				"action":   "login",
				"error":    "wrong-username",
				"username": user,
				"password": password,
			}).Warnln("User attempted to authenticate: Wrong username:", user, password)
			http.Error(rw, "Not valid", http.StatusUnauthorized)
			return
		}

		// Hash
		h := sha1.New()
		h.Write([]byte(password))
		hashed := fmt.Sprintf("%x", h.Sum(nil))

		if hash != hashed {
			fmt.Printf("%q %q", password, hashed)
			s.log.WithFields(logrus.Fields{
				"action":   "login",
				"error":    "wrong-password",
				"username": user,
				"password": password,
			}).Warnln("User attempted to autheticate: Wrong password:", user, password)
			http.Error(rw, "Not valid", http.StatusUnauthorized)
			return
		}

		// Call the wrapped function
		fn(rw, r)
	}
}
