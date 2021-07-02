package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	maxUploadSize = 10 << 20 //10Megabibyte
	uploadRoot    = "exploits"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		ctx:       ctx,
		ServeMux:  http.NewServeMux(),
		ctxCancel: cancel,
		log:       log.New(),
	}
	s.Config.Address = ":8080"

	config, err := ioutil.ReadFile("exm.json")
	err = json.Unmarshal(config, &s)
	if err != nil {
		log.Fatalln("Cannot decode config:", err)
		os.Exit(1)
	}

	exs := s.Session

	exs.WorkAdd(10)
	go exs.WorkSubmitter(ctx.Done())

	httpServer := &http.Server{
		Addr: s.Config.Address,

		Handler:     s.ServeMux,
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

	s.Handle("/", http.FileServer(http.FS(assetFS)))

	s.HandleServerFunc("/targets", handleApiTarget)

	s.HandleServerFunc("/api/newService", handleApiNewService)
	s.HandleServerFunc("/api/exploitChangeState", handleApiExploitSetState)
	s.HandleServerFunc("/api/sessionStatus", handleApiSessionStatus)
	s.HandleServerFunc("/flags", handleApiFlags)
	s.HandleServerFunc("/api/uploadExploit", handleApiUploadExploit)
	s.HandleServerFunc("/api/serviceStatus", handleApiServiceStatus)
	s.HandleServerFunc("/api/exploitStatus", handleApiExploitStatus)
	s.HandleServerFunc("/api/submitterStatus", handleApiSubmitterInfo)
	s.HandleServerFunc("/api/name", handleApiName)

	log.Infoln("Listening on", httpServer.Addr)
	err = httpServer.ListenAndServe()
	log.Errorln("Listening ", err)
	<-quit
}
