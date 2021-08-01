package ex

import (
	"bufio"
	"encoding/json"
	"io"

	log "github.com/sirupsen/logrus"
)

// OutputStream identifies the stream where the output goes
type OutputStream string

const (
	StreamStdout = "STDOUT"
	StreamStderr = "STDERR"
)

var registeredExecutionDumper map[string]func() ExecutionDumper = map[string]func() ExecutionDumper{}

// RegisterFlagStore registers a new FlagStore
func RegisterExecutionDumper(fl func() ExecutionDumper) {
	registeredExecutionDumper[fl().name()] = fl
}

func init() {
	RegisterExecutionDumper(func() ExecutionDumper { return new(SQLiteStore) })
}

// ExecutionDumper writes the log of the execution,
// logs can be stored in a volatile media, as a console,
// or even a persisten one like a file or a database
type ExecutionDumper interface {
	Dump(serviceName string, exploitName string, execID ExecutionID, stream OutputStream, content []byte) error

	json.Marshaler
	json.Unmarshaler

	//maybe this(and FlagStore) should be exported, let's find a name for this method
	name() string
}

type executionDumperWriter struct {
	service string
	exploit string
	dumper  ExecutionDumper
	id      ExecutionID
	stream  OutputStream
}

func (e executionDumperWriter) Write(body []byte) (int, error) {
	err := e.dumper.Dump(e.service, e.exploit, e.id, e.stream, body)
	if err != nil {
		return 0, err
	}

	return len(body), nil
}

func ExecutionDumperToWriter(dumper ExecutionDumper, serviceName string, exploitName string, id ExecutionID, stream OutputStream) io.Writer {
	return executionDumperWriter{
		service: serviceName,
		exploit: exploitName,
		dumper:  dumper,
		id:      id,
		stream:  stream,
	}
}

func ExploitOutputWriter(logger io.Writer, stream OutputStream, dumper ExecutionDumper, t Target, e *Exploit, execID ExecutionID) io.Writer {
	retriever := FlagRetriveWriter(t, e, execID)
	dump := ExecutionDumperToWriter(dumper, e.service.Name(), e.Name(), execID, stream)
	return io.MultiWriter(
		logger,
		retriever,
		dump,
	)
}

// FlagRetriveWriter returns a io.Writer, when wrote the content is searched for flags
func FlagRetriveWriter(t Target, e *Exploit, execId ExecutionID) io.Writer {
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
			e.foundFlag(t, execId, f...)
		}
	}()

	return pw
}
