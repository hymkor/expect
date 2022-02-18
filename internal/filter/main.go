package filter

import (
	"io"
)

type Reader struct {
	In   func() ([]byte, error)
	rest []byte
	eof  bool
}

func (this *Reader) Read(buffer []byte) (int, error) {
	copiedBytes := 0
	for len(buffer) > 0 {
		if this.eof {
			if this.rest == nil || len(this.rest) <= 0 {
				return copiedBytes, io.EOF
			}
		} else {
			bytes, err := this.In()
			if err != nil {
				if err != io.EOF {
					return copiedBytes, err
				}
				this.eof = true
			}
			this.rest = append(this.rest, bytes...)
		}
		n := copy(buffer, this.rest)
		buffer = buffer[n:]
		copiedBytes += n
		newlen := len(this.rest[n:])
		copy(this.rest, this.rest[n:])
		this.rest = this.rest[:newlen]
	}
	return copiedBytes, nil
}
