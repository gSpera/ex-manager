package main

import (
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/gSpera/ex-manager"
)

func main() {
	exs, _ := ex.NewSession("Session", "CCIT\\{.*\\}", "peppe", "vacwm1")
	a := ex.NewService(":S")
	a.AddExploit(ex.NewExploit("Exploit", "./check"))
	exs.AddService(a)

	spew.Dump(exs)
	d, err := json.MarshalIndent(&exs, "", "    ")
	fmt.Println(string(d), err)

	exs.Work()
}
