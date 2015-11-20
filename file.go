package rotating

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"sync"
)

const (
	DEFAULT_MODE = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	DEFAULT_PERM = os.ModePerm
)

type FileHook struct {
	Filename string
	Stream   *StreamHook
	Ok       bool
	Locker   sync.Locker
	mode     int
	perm     os.FileMode
	file     *os.File
	debug    bool
}

func NewFileHook(filename string) (*FileHook, error) {
	file := &FileHook{Filename: filename, mode: DEFAULT_MODE, perm: DEFAULT_PERM,
		Locker: &sync.Mutex{}, Stream: NewStreamHook(nil)}
	if err := file.Open(); err != nil {
		return nil, err
	}
	file.Stream.Locker = nil
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
	if !f.Ok {
		return errors.New("Stream is not ready for writing")
	}

	f.Lock()
	defer f.Unlock()
	return f.Stream.Fire(entry)
}

func (f *FileHook) Lock() {
	if f.Locker != nil {
		f.Locker.Lock()
	}
}

func (f *FileHook) Unlock() {
	if f.Locker != nil {
		f.Locker.Unlock()
	}
}

func (f *FileHook) Open() error {
	var err error
	f.file, err = os.OpenFile(f.Filename, f.mode, f.perm)
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
	m := f.mode
	f.mode = mode

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
	p := f.perm
	f.perm = perm

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
