package log_read_and_write

import (
	"bytes"
	"errors"
	"os"
)

type reader struct {
	write  *LogFile
	read   *os.File
	offset int64
}

var ErrBreakLine = errors.New("broken line")

const readSize = 1 << 8

func NewReader(file *LogFile) (r *reader, err error) {
	r = &reader{
		write:  file,
		read:   nil,
		offset: 0,
	}
	r.read, err = os.Open(file.name)

	return
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	return r.read.Seek(offset, whence)
}

func (r *reader) Read(offset int64) (data []byte, err error) {
	_, err = r.Seek(offset, 0)
	if err != nil {
		return
	}
	var byt [readSize]byte
	var buf = new(bytes.Buffer)
	var n, pos int
	for {
		n, err = r.read.Read(byt[:])
		if err != nil {
			return
		}

		pos = bytes.IndexByte(byt[:n], lineBreak)
		if pos != -1 {
			r.offset += int64(pos)
			buf.Write(byt[:pos])
			break
		}
		r.offset += int64(n)
		buf.Write(byt[:n])
	}

	if pos == -1 {
		err = ErrBreakLine
	}
	return buf.Bytes(), err
}

func (r *reader) Close() error {
	return r.read.Close()
}

func (r *reader) Wait(offset int64) {
	<-r.write.Wait(offset)
}
