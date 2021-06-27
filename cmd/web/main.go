package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gSpera/ex-manager"
	log "github.com/sirupsen/logrus"
)

const (
	maxUploadSize = 10 << 20 //10Megabibyte
	uploadRoot    = "exploits"
)

type Server struct {
	Session *ex.Session `json:"ExConfig"`
	Config  struct {
		Address string
	}

	Users map[string]string

	log       *log.Logger
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		ctx:       ctx,
		ctxCancel: cancel,
		log:       log.New(),
	}
	s.Config.Address = ":8080"

	config, err := ioutil.ReadFile("exm.json")
	err = json.Unmarshal(config, &s)
	if err != nil {
		log.Errorln("Cannot decode config")
		os.Exit(1)
	}

	exs := s.Session

	exs.WorkAdd(10)
	go exs.WorkSubmitter(ctx.Done())

	httpMux := http.NewServeMux()
	httpServer := &http.Server{
		Addr: s.Config.Address,

		Handler:     httpMux,
		BaseContext: func(net.Listener) context.Context { return s.ctx },
	}

	quit := make(chan struct{})
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt)
	go func() {
		<-stopCh
		s.log.Println("Closing, Interrupt detected")
		go func() {
			time.Sleep(10 * time.Second)
			log.Errorln("Force exiting timeout")
			os.Exit(1)
		}()

		s.ctxCancel()
		httpServer.Close()
		s.Session.Stop()
		fl, err := os.Create("exm.json")
		if err != nil {
			s.log.Errorln("Couldn't open json:", err)
			return
		}

		jsonEncoder := json.NewEncoder(fl)
		jsonEncoder.SetIndent("", "\t")
		err = jsonEncoder.Encode(s)
		if err != nil {
			s.log.Errorln("Could not encode:", err)
			return
		}
		fl.Close()

		s.log.Writer().Close()
		log.Println("Done closing")
		quit <- struct{}{}
	}()

	// http.HandleFunc("/", serverHandler(s, handleHome))

	httpMux.HandleFunc("/login", func(rw http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()

		if !ok {
			rw.Header().Add("WWW-Authenticate", "Basic realm=\"Fai schifo\"")
			http.Error(rw, "Not logged in", http.StatusUnauthorized)
			return
		}

		hash, ok := s.Users[user]
		if !ok {
			http.Error(rw, "User not found", http.StatusUnauthorized)
			return
		}

		h := sha1.New()
		h.Write([]byte(password))
		hashed := fmt.Sprintf("%x", h.Sum(nil))

		if hash != hashed {
			fmt.Printf("%q %q", password, hashed)
			s.log.Warnln("User attempted to autheticate:", user, password, hash, hashed)
			http.Error(rw, "Not valid", http.StatusUnauthorized)
			return
		}

		fmt.Fprintf(rw, "Ok")
		return
	})

	httpMux.HandleFunc("/unlog", func(rw http.ResponseWriter, r *http.Request) {
		http.Error(rw, "Unlogged", http.StatusUnauthorized)
	})

	httpMux.Handle("/", http.FileServer(http.Dir("cmd/web/asset")))

	httpMux.HandleFunc("/targets", serverHandler(s, handleApiTarget))

	httpMux.HandleFunc("/api/newService", serverHandler(s, handleApiNewService))
	httpMux.HandleFunc("/api/exploitChangeState", serverHandler(s, handleApiExploitSetState))
	httpMux.HandleFunc("/api/sessionStatus", serverHandler(s, handleApiSessionStatus))
	httpMux.HandleFunc("/flags", serverHandler(s, handleApiFlags))
	httpMux.HandleFunc("/api/uploadExploit", serverHandler(s, handleApiUploadExploit))
	httpMux.HandleFunc("/api/serviceStatus", serverHandler(s, handleApiServiceStatus))
	httpMux.HandleFunc("/api/exploitStatus", serverHandler(s, handleApiExploitStatus))
	httpMux.HandleFunc("/api/name", serverHandler(s, handleApiName))

	log.Infoln("Listening on", httpServer.Addr)
	err = httpServer.ListenAndServe()
	log.Errorln("Listening ", err)
	<-quit
}
