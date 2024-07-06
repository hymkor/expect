package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var marker = []byte("\nSCRIPT\n")

func readEmbedScript(fname string) (string, error) {
	fd, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	stat, err := fd.Stat()
	if err != nil {
		return "", err
	}
	size := stat.Size()

	var startBytes [8]byte
	if _, err = fd.Seek(-int64(len(startBytes)), os.SEEK_END); err != nil {
		return "", err
	}
	n, err := io.ReadFull(fd, startBytes[:])
	if err != nil {
		return "", err
	}
	start, _ := binary.Varint(startBytes[:n])

	if start <= 0 || start >= size {
		return "", fmt.Errorf("marker address is too few or too large (0<%v<%v)", start, size)
	}
	if _, err = fd.Seek(start, os.SEEK_SET); err != nil {
		return "", err
	}
	n, err = io.ReadFull(fd, startBytes[:])
	if err != nil {
		return "", err
	}
	if !bytes.Equal(startBytes[:], marker) {
		return "", fmt.Errorf("marker does not match %#v != %#v", startBytes, marker)
	}
	script, err := io.ReadAll(fd)
	if err != nil {
		return "", nil
	}
	return string(script[:len(script)-len(startBytes)]), nil
}

func compile(newExeName, exeName, scriptName string) error {
	if len(newExeName) < 4 || !strings.EqualFold(newExeName[len(newExeName)-4:], ".exe") {
		newExeName = fmt.Sprintf("%s.exe", newExeName)
	}
	_w, err := os.Create(newExeName)
	if err != nil {
		return err
	}
	defer _w.Close()
	w := bufio.NewWriter(_w)
	defer w.Flush()

	r1, err := os.Open(exeName)
	if err != nil {
		return err
	}
	defer r1.Close()

	size, err := io.Copy(w, r1)
	if err != nil {
		return err
	}
	r1.Close()

	_, err = w.Write(marker[:])
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "--[[ Built-in-Script by %s@%s on %s]]--\n",
		os.Getenv("USERNAME"),
		os.Getenv("USERDOMAIN"),
		time.Now().Format(time.RFC1123Z))

	r2, err := os.Open(scriptName)
	if err != nil {
		return err
	}
	defer r2.Close()

	_, err = io.Copy(w, r2)
	if err != nil {
		return err
	}
	var sizeBuffer [8]byte
	binary.PutVarint(sizeBuffer[:], size)
	_, err = w.Write(sizeBuffer[:])
	return err
}
