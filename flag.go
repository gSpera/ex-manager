package ex

import (
	"bufio"
	"io"

	log "github.com/sirupsen/logrus"
)

// FlagRetriveWriter creates a io.Writer, when wrote the content is logged and flags are searched
func FlagRetriveWriter(l *log.Entry, t Target, e *Exploit) io.Writer {
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
				return
			}

			f := e.service.session.SearchFlagsInText(string(line))
			e.foundFlag(t, f...)

			l.Println("Program Stdout:", string(line))
			if len(f) > 0 {
				l.Println("Found Flags: ", f)
			}
		}
	}()

	return pw
}

func (s *Session) SearchFlagsInText(str string) []string {
	return s.flagRegex.FindAllString(str, -1)
}
