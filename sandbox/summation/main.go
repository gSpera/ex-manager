package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"
)

func handle(c io.ReadWriteCloser) {
	rand.Seed(time.Now().Unix())
	random1 := rand.Int() % 100
	random2 := rand.Int() % 100
	log.Println("Connection")
	log.Println(random1, random2)

	sum := random1 + random2
	flag := fmt.Sprintf("CCIT{%d_SUM_%d+%d=%d_}", time.Now().Nanosecond(), random1, random2, sum)

	var v int
	fmt.Fprintln(c, random1)
	fmt.Fprintln(c, random2)
	if _, err := fmt.Fscanf(c, "%d\n", &v); err != nil {
		fmt.Fprintln(c, "Another time??")
		return
	}
	if random1 > 70 {
		fmt.Fprintln(c, "SUM{2+2=5}")
		return
	}

	if v != sum {
		fmt.Fprintln(c, "Another time??")
		return
	}

	fmt.Fprintln(c, flag)
	c.Close()
}

func main() {
	l, err := net.Listen("tcp", ":8081")
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
