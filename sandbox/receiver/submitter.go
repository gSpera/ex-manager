package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func handle(c io.ReadWriter) {
	flags := os.Args[1:]
	for _, f := range flags {
		var line string
		_, err := fmt.Fprintln(c, f)
		if err != nil {
			panic(err)
		}
		_, err = fmt.Fscanln(c, &line)
		if err != nil {
			panic(err)
		}
		switch line {
		case "DONE":
			fmt.Println(f, "SUCCESS")
		case "EXPIRED":
			fmt.Println(f, "EXPIRED")
		}
	}
}

func main() {
	c, err := net.Dial("tcp", ":8085")
	if err != nil {
		panic(err)
	}
	handle(c)
	c.Close()
}
