package ex

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct{ *sql.DB }

// Put(...Flag) error
// GetByName(serviceName string, exploitName string) ([]Flag, error)
// UpdateState(flagValue string, flagState SubmittedFlagStatus) error
// GetValueToSubmit(limit int) ([]string, error)
// GetFlagsSubmittedDuring(from time.Time, to time.Time) ([]Flag, error)

// name() string
// json.Marshaler
// json.Unmarshaler
func NewSqliteStore(file string) (SQLiteStore, error) {
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return SQLiteStore{}, err
	}
	return SQLiteStore{db}, nil
}

func (s SQLiteStore) CreateTables() error {
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

func (s SQLiteStore) InsertRow(f Flag) error {
	_, err := s.Exec(`INSERT INTO flags VALUES (?, ?, ?, ?, ?, ?, ?)`, f.Value, f.ServiceName, f.ExploitName, f.From, string(f.Status), f.TakenAt, f.SubmittedAt)
	if err != nil {
		return fmt.Errorf("cannot insert flag, is the value unique??: %w", err)
	}

	return nil
}
