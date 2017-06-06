package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func wait(keyword string, ch chan []byte) bool {
	keywordByte := []byte(keyword)
	alltext := make([]byte, 0, 4096)
	for {
		text, ok := <-ch
		if !ok {
			return false
		}
		fmt.Print(string(text))
		alltext = append(alltext, text...)
		if bytes.Index(alltext, keywordByte) >= 0 {
			return true
		}
	}
}

func waitAndTell(args []string, keywords []string) error {
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
	out, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer out.Close()
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
		fmt.Fprintf(out, "%s\r\n", keywords[i])
		fmt.Println(keywords[i])
		i++
	}
	cmd.Wait()
	return nil
}

func Main() error {
	keywords := make([]string, 100)
	scnr := bufio.NewScanner(os.Stdin)
	for scnr.Scan() {
		keywords = append(keywords, scnr.Text())
	}
	return waitAndTell(os.Args[1:], keywords)
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
