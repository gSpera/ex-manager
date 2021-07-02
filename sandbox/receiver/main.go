package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
)

func handle(c io.ReadWriteCloser) {
	defer c.Close()
	for {
		var flag string
		_, err := fmt.Fscanln(c, &flag)
		if err != nil {
			log.Println(err)
			return
		}

		if rand.Int()%100 > 82 {
			fmt.Fprintln(c, "EXPIRED")
		}
		fmt.Fprintln(c, "DONE")
	}
}

func main() {
	l, err := net.Listen("tcp", ":8085")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handle(conn)
	}
}
