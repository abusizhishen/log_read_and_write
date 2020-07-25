package log_read_and_write

import (
	"bufio"
	"os"
	"sync"
)

var lineBreak byte = '\n'
var lineBreakByte = []byte{lineBreak}
var ready = make(chan struct{})

func init() {
	close(ready)
}

type LogFile struct {
	name  string
	file  *os.File
	ready chan struct{}
	head  int64
	sync.RWMutex
}

func (l *LogFile) Read(byt []byte) (n int, err error) {
	l.RLock()
	defer l.RUnlock()
	return l.file.Read(byt)
}

func (l *LogFile) Write(byt []byte) (n int, err error) {
	l.Lock()
	defer l.Unlock()
	n, err = l.file.Write(byt)
	if err != nil {
		return 0, err
	}
	n, err = l.file.Write(lineBreakByte)
	if err == nil {
		l.head += int64(len(byt) + 1)
	}
	close(l.ready)
	l.ready = make(chan struct{})
	return
}

func New(name string) (l *LogFile, err error) {
	l = &LogFile{
		name:  name,
		file:  nil,
		ready: make(chan struct{}),
	}

	l.file, err = os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	return
}

func (l *LogFile) Close() error {
	return l.file.Close()
}

func (l *LogFile) Size() (size int64, err error) {
	l.RLock()
	defer l.RUnlock()
	info, err := l.file.Stat()
	if err != nil {
		return
	}

	return info.Size(), nil
}

type Row struct {
	Data []byte
	Err  error
}

func (l *LogFile) Seek(offset int64, whence int) (int64, error) {
	l.Lock()
	defer l.Unlock()
	return l.file.Seek(offset, whence)
}

func (l *LogFile) ReadLineByLine(rows chan<- Row) {
	l.RLock()
	defer l.RUnlock()
	defer close(rows)

	var row Row
	var reader = bufio.NewReader(l.file)
	for {
		row.Data, row.Err = reader.ReadSlice(lineBreak)
		if row.Err != nil {
			return
		}

		//返回数据包含最后一个字换行，需要去掉
		if len(row.Data) > 1 {
			row.Data = row.Data[:len(row.Data)-1]
		}

		rows <- row
	}
}

func (l *LogFile) Wait(offset int64) <-chan struct{} {
	l.RLock()
	defer l.RUnlock()

	if l.head > offset {
		return ready
	}
	return l.ready
}
