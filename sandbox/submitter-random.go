package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	for _, f := range os.Args[1:] {
		fmt.Printf("%s %s\n", f, random())
	}
}

func random() string {
	v := []string{
		"SUCCESS",
		"EXPIRED",
		"INVALID",
		"ALREADY-SUBMITTED",
		"TEAM-OWN",
		"TEAM-NOP",
		"OFFLINE-CTF",
		"OFFLINE-SERVICE",
	}
	return v[rand.Intn(len(v))]
}
