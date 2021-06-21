package ex

import (
	"bufio"
	"io"

	log "github.com/sirupsen/logrus"
)

func FlagRetriveWriter(l *log.Entry, e *Exploit) io.Writer {
	pr, pw := io.Pipe()

	go func() {
		r := bufio.NewReader(pr)
		for {
			line, err := r.ReadBytes('\n')
			if err == io.EOF {
				return
			}

			if err != nil {
				log.Error("Cannot read:", err)
			}

			f := e.service.session.SearchFlagsInText(string(line))
			e.foundFlag(f...)

			l.Println("Output:", string(line))
			if len(f) > 0 {
				l.Println("Found: ", f)
			}
		}
	}()

	return pw
}

func (s *Session) SearchFlagsInText(str string) []string {
	return s.flagRegex.FindAllString(str, -1)
}
