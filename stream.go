package rotating

import (
	"bufio"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
	"os"
	"sync"
)

type StreamHook struct {
	Terminator string
	writer     *bufio.Writer
	Locker     sync.Locker
	Debug      bool
}

func NewStreamHook(writer io.Writer) *StreamHook {
	return &StreamHook{Terminator: "\n", writer: bufio.NewWriter(writer), Locker: &sync.Mutex{}}
}

func (h *StreamHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func (h *StreamHook) Fire(entry *logrus.Entry) error {
	h.Lock()
	defer h.Unlock()

	var line string
	var err error

	if line, err = entry.String(); err != nil {
		if h.Debug {
			fmt.Fprintf(os.Stderr, "Unable to read entry: %v\n", err)
		}
		return err
	}
	if _, err := h.writer.WriteString(line); err != nil {
		if h.Debug {
			fmt.Fprintf(os.Stderr, "Unable to write the content: %v\n", err)
		}
		return err
	}
	if err := h.writer.Flush(); err != nil {
		if h.Debug {
			fmt.Fprintf(os.Stderr, "Unable to flush the buffer: %v\n", err)
		}
		return err
	}

	return nil
}

func (h *StreamHook) SetWriter(writer io.Writer) {
	h.writer = bufio.NewWriter(writer)
}

func (h *StreamHook) Lock() {
	if h.Locker != nil {
		h.Locker.Lock()
	}
}

func (h *StreamHook) Unlock() {
	if h.Locker != nil {
		h.Locker.Unlock()
	}
}
