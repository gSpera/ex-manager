package ex

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	_ "modernc.org/sqlite"
)

var _ FlagStore = &SQLiteStore{}

func init() {
	RegisterFlagStore(func() FlagStore {
		return new(SQLiteStore)
	})
}

type SQLiteStore struct {
	url string
	*sql.DB
}

// Put(...Flag) error
// GetByName(serviceName string, exploitName string) ([]Flag, error)
// UpdateState(flagValue string, flagState SubmittedFlagStatus) error
// GetValueToSubmit(limit int) ([]string, error)
// GetFlagsSubmittedDuring(from time.Time, to time.Time) ([]Flag, error)

// name() string
// json.Marshaler
// json.Unmarshaler
func NewSqliteStore(file string) (*SQLiteStore, error) {
	s := SQLiteStore{}
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
		takenAt time,
		submittedAt time
	);`)
	return err
}

func (s *SQLiteStore) InsertRow(f Flag) error {
	_, err := s.Exec(`INSERT INTO flags VALUES (?, ?, ?, ?, ?, ?, ?)`, f.Value, f.ServiceName, f.ExploitName, f.From, string(f.Status), f.TakenAt, f.SubmittedAt)
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
	if s.DB == nil {
		panic(s)
	}
	flags := make([]Flag, 0)
	rows, err := s.Query("SELECT * FROM flags WHERE service=? AND exploit=?", serviceName, exploitName)
	if err != nil {
		return flags, fmt.Errorf("cannot query database: %w", err)
	}

	for rows.Next() {
		f := Flag{}
		err := rows.Scan(&f)
		if err != nil {
			return flags, fmt.Errorf("cannot scan row: %w", err)
		}

		flags = append(flags, f)
	}

	return flags, nil
}

func (s *SQLiteStore) UpdateState(flagValue string, state SubmittedFlagStatus) error {
	_, err := s.Exec("UPDATE flags SET state=? WHERE value=?", state, flagValue)
	if err != nil {
		return fmt.Errorf("cannot update: %w", err)
	}

	return nil
}

func (s *SQLiteStore) GetValueToSubmit(limit int) ([]string, error) {
	flags := make([]string, 0)
	rows, err := s.Query("SELECT value FROM flags WHERE status=NOT-SUBMITTED LIMIT ? ORDER BY takenAt", limit)
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
	rows, err := s.Query("SELECT * FROM FLAGS")
	if err != nil {
		return flags, fmt.Errorf("cannot query: %w", err)
	}

	for rows.Next() {
		var f Flag
		err := rows.Scan(&f)
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

	defer spew.Dump(s)
	return s.init(s.url)
}
