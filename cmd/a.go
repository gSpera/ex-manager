package main

import (
	"time"

	"github.com/gSpera/ex-manager"
)

func main() {
	q, err := ex.NewSqliteStore("db")
	if err != nil {
		panic(err)
	}
	err = q.CreateTables()

	err = q.InsertRow(ex.Flag{
		Value:       "XX_{}",
		ServiceName: "A",
		ExploitName: "A",
		From:        "1",
		Status:      ex.FlagAlreadySubmitted,
		TakenAt:     time.Now(),
		SubmittedAt: time.Now(),
	}) // maybe implement some better structure

	if err != nil {
		panic(err)
	}
}
