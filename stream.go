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
	terminator string
	writer     *bufio.Writer
	locker     sync.Locker
	debug      bool
}

func NewStreamHook(writer io.Writer) *StreamHook {
	return &StreamHook{writer: bufio.NewWriter(writer), locker: &sync.Mutex{}}

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
		if h.debug {
			fmt.Fprintf(os.Stderr, "Unable to read entry: %v\n", err)
		}
		return err
	}
	if h.terminator != "" {
		line += h.terminator
	}
	if _, err := h.writer.WriteString(line); err != nil {
		if h.debug {
			fmt.Fprintf(os.Stderr, "Unable to write the content: %v\n", err)
		}
		return err
	}
	if err := h.writer.Flush(); err != nil {
		if h.debug {
			fmt.Fprintf(os.Stderr, "Unable to flush the buffer: %v\n", err)
		}
		return err
	}

	return nil
}

func (h *StreamHook) SetWriter(writer io.Writer) *StreamHook {
	h.writer = bufio.NewWriter(writer)
	return h
}

func (h *StreamHook) Lock() {
	if h.locker != nil {
		h.locker.Lock()
	}
}

func (h *StreamHook) Unlock() {
	if h.locker != nil {
		h.locker.Unlock()
	}
}

func (h *StreamHook) SetLock(locker sync.Locker) *StreamHook {
	h.locker = locker
	return h
}

func (h *StreamHook) SetDebug(debug bool) *StreamHook {
	h.debug = debug
	return h
}

func (h *StreamHook) SetTerminator(t string) *StreamHook {
	h.terminator = t
	return h
}
