package ex

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var _ FlagStore = &SQLiteStore{}
var _ ExecutionDumper = &SQLiteStore{}

func init() {
	RegisterFlagStore(func() FlagStore {
		return new(SQLiteStore)
	})
	RegisterExecutionDumper(func() ExecutionDumper {
		return new(SQLiteStore)
	})
}

// SQLiteStore is a FlagStore and a ExecutionDumper.
type SQLiteStore struct {
	url string
	*sql.DB
}

func NewSqliteStore(file string) (*SQLiteStore, error) {
	s := &SQLiteStore{}
	err := s.init(file)
	return s, err
}

func (s *SQLiteStore) init(file string) error {
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return err
	}
	s.DB = db

	err = s.Ping()
	if err != nil {
		return fmt.Errorf("cannot ping: %w", err)
	}

	err = s.CreateTables()
	if err != nil {
		return fmt.Errorf("cannot create tables: %w", err)
	}

	return nil
}

func (s *SQLiteStore) CreateTables() error {
	_, err := s.Exec(`
	CREATE TABLE IF NOT EXISTS flags (
		value string UNIQUE,
		service string,
		exploit string,
		fromTarget string,
		status string,
		takenAt INTEGER,
		submittedAt INTEGER,
		executionId INTEGER
	);
	CREATE TABLE IF NOT EXISTS execlogs (
		"service"	NUMERIC,
		"exploit"	INTEGER,
		"target"    TEXT,
		"execid"	INTEGER,
		"stream_name"	TEXT,
		"content"	TEXT,
		"time" INTEGER
	);`)

	return err
}

func (s *SQLiteStore) InsertRow(f Flag) error {
	_, err := s.Exec(`INSERT INTO flags VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		f.Value,
		f.ServiceName,
		f.ExploitName,
		f.From,
		string(f.Status),
		f.TakenAt.UnixNano(),
		f.SubmittedAt.UnixNano(),
		f.ExecutionID,
	)

	if err != nil {
		return fmt.Errorf("cannot insert flag, is the value unique??: %w", err)
	}

	return nil
}

func (s *SQLiteStore) Put(flags ...Flag) error {
	for _, f := range flags {
		err := s.InsertRow(f)
		if err != nil {
			return fmt.Errorf("cannot send flag %q: %w", f, err)
		}
	}

	return nil
}

func (s *SQLiteStore) GetByName(serviceName string, exploitName string) ([]Flag, error) {
	flags := make([]Flag, 0)
	rows, err := s.Query("SELECT * FROM flags WHERE service=? AND exploit=?", serviceName, exploitName)
	if err != nil {
		return flags, fmt.Errorf("cannot query database: %w", err)
	}

	for rows.Next() {
		f := Flag{}
		err := rows.Scan(&f.Value, &f.ServiceName, &f.ExploitName, &f.From, &f.Status, &timeScan{&f.TakenAt}, &timeScan{&f.SubmittedAt}, &f.ExecutionID)
		if err != nil {
			return flags, fmt.Errorf("cannot scan row: %w", err)
		}

		flags = append(flags, f)
	}

	return flags, nil
}

func (s *SQLiteStore) UpdateState(flagValue string, state SubmittedFlagStatus) error {
	_, err := s.Exec("UPDATE flags SET status=?, submittedAt=? WHERE value=?", state, time.Now().UnixNano(), flagValue)
	if err != nil {
		return fmt.Errorf("cannot update: %w", err)
	}

	return nil
}

func (s *SQLiteStore) GetValueToSubmit(limit int) ([]string, error) {
	flags := make([]string, 0)
	rows, err := s.Query("SELECT value FROM flags WHERE status=\"NOT-SUBMITTED\" ORDER BY takenAt  LIMIT ?", limit)
	if err != nil {
		return flags, fmt.Errorf("cannot query: %w", err)
	}

	for rows.Next() {
		var f string
		err := rows.Scan(&f)
		if err != nil {
			return flags, fmt.Errorf("cannot scan: %w", err)
		}
		flags = append(flags, f)
	}

	return flags, nil
}

func (s *SQLiteStore) GetFlagsSubmittedDuring(from time.Time, to time.Time) ([]Flag, error) {
	flags := make([]Flag, 0)
	rows, err := s.Query("SELECT * FROM flags WHERE submittedAt >= ? and submittedAt <= ?", from.UnixNano(), to.UnixNano())
	if err != nil {
		return flags, fmt.Errorf("cannot query: %w", err)
	}

	for rows.Next() {
		var f Flag
		err := rows.Scan(&f.Value, &f.ServiceName, &f.ExploitName, &f.From, &f.Status, &timeScan{&f.TakenAt}, &timeScan{&f.SubmittedAt}, &f.ExecutionID)
		if err != nil {
			return flags, fmt.Errorf("cannot scan: %w", err)
		}
		flags = append(flags, f)
	}

	return flags, nil
}

func (s *SQLiteStore) name() string { return "SQLITE" }

func (s *SQLiteStore) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.url)
}

func (s *SQLiteStore) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &s.url)
	if err != nil {
		return err
	}

	if strings.TrimSpace(s.url) == "" {
		return fmt.Errorf("sqlite: no url")
	}

	return s.init(s.url)
}

func (s *SQLiteStore) Dump(service string, exploit string, target Target, execID ExecutionID, stream OutputStream, body []byte) error {
	_, err := s.Exec(`INSERT INTO execlogs VALUES (?, ?, ?, ?, ?, ?, ?)`,
		service, exploit, target, execID, string(stream), string(body), time.Now().UnixNano())

	return err
}

func (s *SQLiteStore) LogsFromExecID(execID ExecutionID) ([]ExecutionLog, error) {
	rows, err := s.Query(`SELECT * FROM execlogs WHERE execID = ?`, execID)

	if err != nil {
		return []ExecutionLog{}, fmt.Errorf("Cannot query logs for execution: %q: %w", execID, err)
	}

	logs := make([]ExecutionLog, 0, 2)

	for rows.Next() {
		var l ExecutionLog
		err := rows.Scan(&l.ServiceName, &l.ExploitName, &l.ExecutionID, &l.Stream, &l.Content, &timeScan{&l.When})
		if err != nil {
			return []ExecutionLog{}, fmt.Errorf("Cannot scan row: %w", err)
		}

		logs = append(logs, l)
	}

	return logs, nil
}

func (s *SQLiteStore) ExecIDsFromServiceExploitTarget(serviceName string, exploitName string, target Target) ([]ExecutionID, error) {
	rows, err := s.Query(`SELECT execID FROM execlogs WHERE service=? AND exploit=? AND target=?`, serviceName, exploitName, target)

	if err != nil {
		return nil, fmt.Errorf("Cannot query logs from service, exploit, target: %w", err)
	}

	ids := []ExecutionID{}
	for rows.Next() {
		var id ExecutionID
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("Cannot scan exec id: %w", err)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

type timeScan struct{ *time.Time }

func (t *timeScan) Scan(v interface{}) error {
	content, ok := v.(int64)
	if !ok {
		panic(v)
	}

	tt := time.Unix(0, int64(content))

	*t.Time = tt

	return nil
}

func (t *timeScan) Value() (driver.Value, error) {
	return t.Format(time.RFC3339Nano), nil
}
