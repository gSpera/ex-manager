package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func flagsInTxt(input string, regex *regexp.Regexp) []string {
	return regex.FindAllString(input, -1)
}

func main() {
	flags := []string{}

	target := flag.String("target", "127.0.0.1", "The target")
	exploit := flag.String("exploit", "", "The exploit to run")
	submit := flag.String("submit", "", "The submitter")
	flag.Parse()

	if *exploit == "" {
		fmt.Println("No exploit given")
		return
	}

	cmd := exec.Command(*exploit, "-t", *target)
	fmt.Println(cmd)
	stdoutR, stdoutW := io.Pipe()
	stdoutBufReader := bufio.NewReader(stdoutR)
	cmd.Stderr = cmd.Stdout
	cmd.Stdout = stdoutW

	err := cmd.Start()
	if err != nil {
		log.Panicf("Cannot execute exploit: %v\n", err)
		return
	}

	go func() {
		cmd.Wait()
		log.Println("Command finished executing")
		stdoutW.Close()
	}()

	for {
		l, err := stdoutBufReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			break
		}

		log.Print("Output: ", l)
		for _, f := range flagsInTxt(l, regexp.MustCompile("CCIT\\{.*\\}")) {
			flags = append(flags, f)
			log.Println("Flag: ", f)
		}
	}

	log.Println("Command closed output")

	if *submit == "" {
		return
	}

	cmd = exec.Command(*submit, flags...)
	cmd.Stdout = log.StandardLogger().Writer()
	cmd.Stderr = log.StandardLogger().Writer()

	err = cmd.Run()
	if err != nil {
		log.Panicf("Cannot execute submit: %v\n", err)
		return
	}
}
