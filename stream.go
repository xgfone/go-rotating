package rotating

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
	"os"
	"sync"
)

const (
	DEFAULT_MODE = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	DEFAULT_PERM = os.ModePerm
)

//////////////
// Stream
type StreamHook struct {
	Terminator string
	writer     *bufio.Writer
	Lock       sync.Locker
	Debug      bool
}

func NewStreamHook(writer io.Writer) *StreamHook {
	return &StreamHook{Terminator: "\n", writer: bufio.NewWriter(writer), Lock: &sync.Mutex{}}
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
	h.Lock.Lock()
	defer h.Lock.Unlock()

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

//////////////
// FileHook
type FileHook struct {
	Filename string
	Mode     int
	Perm     os.FileMode
	Stream   *StreamHook
	file     *os.File
	Ok       bool
	debug    bool
}

func NewFileHook(filename string) (*FileHook, error) {
	file := &FileHook{Filename: filename, Mode: DEFAULT_MODE, Perm: DEFAULT_PERM, Stream: NewStreamHook(nil)}
	if err := file.Open(); err != nil {
		return nil, err
	}
	return file, nil
}

func (f *FileHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func (f *FileHook) Fire(entry *logrus.Entry) error {
	if f.Ok {
		return f.Stream.Fire(entry)
	} else {
		return errors.New("Stream is not ready for writing")
	}
}

func (f *FileHook) Open() error {
	var err error
	f.file, err = os.OpenFile(f.Filename, f.Mode, f.Perm)
	if err != nil {
		f.Ok = false
		if f.debug {
			fmt.Fprintf(os.Stderr, "Unable to open the file[%v]: %v\n", f.Filename, err)
		}
		return err
	} else {
		f.Ok = true
		f.Stream.SetWriter(f.file)
		return nil
	}
}

func (f *FileHook) Close() error {
	if f.Ok {
		err := f.file.Close()
		if err != nil {
			if f.debug {
				fmt.Fprintf(os.Stderr, "Unable to close the file[%v]: %v\n", f.Filename, err)
			}
			return err
		}
	}
	return nil
}

func (f *FileHook) SetMode(mode int) (int, error) {
	m := f.Mode
	f.Mode = mode

	err := f.Close()
	if err != nil {
		return m, err
	}

	err = f.Open()
	if err != nil {
		return m, err
	}

	return m, nil
}

func (f *FileHook) SetPerm(perm os.FileMode) (os.FileMode, error) {
	p := f.Perm
	f.Perm = perm

	err := f.Close()
	if err != nil {
		return p, err
	}

	err = f.Open()
	if err != nil {
		return p, err
	}
	return p, nil
}

func (f *FileHook) SetDebug(debug bool) {
	f.debug = debug
	f.Stream.Debug = debug
}
