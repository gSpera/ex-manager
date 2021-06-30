package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
)

func handle(c io.ReadWriteCloser) {
	for {
		var flag string
		_, err := fmt.Fscanln(c, &flag)
		if err != nil {
			log.Println(err)
			return
		}

		if rand.Int()%100 > 82 {
			fmt.Fprintln(c, flag, "EXPIRED")
			return
		}
		fmt.Fprintln(c, flag, "DONE")
	}
}

func main() {
	l, err := net.Listen("tcp", ":8082")
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
