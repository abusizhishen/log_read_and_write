package log_read_and_write

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		l, err := New("a.log")
		assert.IsType(t, &LogFile{}, l)
		assert.Nil(t, err)
		l.Close()
	})
}

func TestLogFile_Write(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		l, _ := New("a.log")
		defer l.Close()
		var s = []byte("hello")
		_, err := l.Write(s)
		assert.Nil(t, err)
	})
}

func TestReader_Read(t *testing.T) {
	t.Run("read_and_write", func(t *testing.T) {
		l, _ := New("a.log")
		defer l.Close()
		defer os.Remove(l.name)
		rand.Seed(time.Now().Unix())
		r, err := NewReader(l)
		assert.Nil(t, err)

		go func() {
			var s = "hello_%d"
			var i int
			var err error
			for {
				_, err = l.Write([]byte(fmt.Sprintf(s, i)))
				assert.Nil(t, err)
				i++
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			}
		}()

		defer r.Close()
		var offset int64
		var by []byte
		for {
			by, err = r.Read(offset)
			if err != nil {
				if err == io.EOF {
					logrus.Infof("wait appending")
					r.Wait(offset)
					continue
				}
				t.Error(err)
			}

			logrus.Infof("%s", by)
			offset += int64(len(by) + 1)
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(30)))
		}
	})
}
