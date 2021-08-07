package ex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
)

// OutputStream identifies the stream where the output goes
type OutputStream string

const (
	StreamStdout = "STDOUT"
	StreamStderr = "STDERR"
)

type ExecutionLog struct {
	ServiceName string
	ExploitName string
	Target      Target
	ExecutionID ExecutionID
	Stream      OutputStream
	Content     string
	When        time.Time
}

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
	NewExecution(serviceName string, exploitName string, target Target) (ExecutionID, error)
	Dump(serviceName string, exploitName string, target Target, execID ExecutionID, stream OutputStream, content []byte) error
	LogsFromExecID(ExecutionID) ([]ExecutionLog, error)
	LatestExecIDTimeFromServiceExploitTarget(serviceName string, exploitName string, target Target) (ExecutionID, time.Time, bool, error)

	json.Marshaler
	json.Unmarshaler

	//maybe this(and FlagStore) should be exported, let's find a name for this method
	name() string
}

type executionDumperWriter struct {
	service string
	exploit string
	target  Target
	dumper  ExecutionDumper
	id      ExecutionID
	stream  OutputStream
}

func (e executionDumperWriter) Write(body []byte) (int, error) {
	err := e.dumper.Dump(e.service, e.exploit, e.target, e.id, e.stream, body)
	if err != nil {
		return 0, err
	}

	return len(body), nil
}

func ExecutionDumperToWriter(dumper ExecutionDumper, serviceName string, exploitName string, target Target, id ExecutionID, stream OutputStream) io.Writer {
	return executionDumperWriter{
		service: serviceName,
		exploit: exploitName,
		target:  target,
		dumper:  dumper,
		id:      id,
		stream:  stream,
	}
}

type writerAndCloser struct {
	io.Writer
	io.Closer
}

type closerFn func() error

func (fn closerFn) Close() error {
	return fn()
}

func ExploitOutputWriter(logger io.WriteCloser, stream OutputStream, dumper ExecutionDumper, t Target, e *Exploit, execID ExecutionID) io.WriteCloser {
	retriever := FlagRetriveWriter(t, e, execID)
	dump := ExecutionDumperToWriter(dumper, e.service.Name(), e.Name(), t, execID, stream)
	writer := io.MultiWriter(
		logger,
		retriever,
		dump,
	)
	return writerAndCloser{
		Writer: writer,
		Closer: closerFn(func() error {
			err1 := logger.Close()
			err2 := retriever.Close()
			switch {
			case err1 != nil && err2 != nil:
				// cannot wrap??
				return fmt.Errorf("Error while closing both logger and retriever: logger: %v; retriever: %v", err1, err2)
			case err1 != nil:
				return fmt.Errorf("Error while closing logger: %w", err1)
			case err2 != nil:
				return fmt.Errorf("Error while closing retriever: %w", err2)
			}

			return nil
		}),
	}
}

// FlagRetriveWriter returns a io.Writer, when wrote the content is searched for flags
func FlagRetriveWriter(t Target, e *Exploit, execId ExecutionID) io.WriteCloser {
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
