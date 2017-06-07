package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/zetamatta/experimental/writeconsole"
)

func wait(keyword string, ch chan []byte) bool {
	keywordByte := []byte(keyword)
	alltext := make([]byte, 0, 4096)
	for {
		text, ok := <-ch
		if !ok {
			return false
		}
		fmt.Printf("<" + string(text) + ">")
		alltext = append(alltext, text...)
		if bytes.Index(alltext, keywordByte) >= 0 {
			return true
		}
	}
}

func waitAndTell(args []string, keywords []string) error {
	console, err := writeconsole.NewHandle()
	if err != nil {
		return err
	}

	cmd := exec.Command(args[0], args[1:]...)
	in1, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	in2, err := cmd.StderrPipe()
	if err != nil {
		defer in1.Close()
		return err
	}

	cmd.Stdin = os.Stdin

	ch := make(chan []byte, 10)
	go func() {
		var data [256]byte
		for {
			n, err := in1.Read(data[:])
			if err != nil {
				in1.Close()
				return
			}
			ch <- data[:n]
		}
	}()
	go func() {
		var data [256]byte
		for {
			n, err := in2.Read(data[:])
			if err != nil {
				in2.Close()
				return
			}
			ch <- data[:n]
		}
	}()
	if err := cmd.Start(); err != nil {
		return err
	}
	i := 0
	for i < len(keywords) {
		if !wait(keywords[i], ch) {
			fmt.Fprintln(os.Stderr, "Found EOF")
			return io.EOF
		}
		i++
		if i >= len(keywords) {
			break
		}
		console.WriteString(keywords[i])
		console.WriteRune('\r')
		i++
	}
	cmd.Wait()
	return nil
}

func Main() error {
	if len(os.Args) < 3 {
		return errors.New("too few arguments")
	}
	keywords_bin, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		return err
	}
	keywords := strings.Split(string(keywords_bin), "\n")
	for i := 0 ; i < len(keywords) ; i ++ {
		keywords[i] = strings.TrimSpace(keywords[i])
		println("keyword='" + keywords[i] + "'")
	}
	return waitAndTell(os.Args[2:], keywords)
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
