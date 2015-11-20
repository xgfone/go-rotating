package rotating

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"
)

const (
	MIDNIGHT = 60 * 60 * 24
	DATE     = "2006-01-02"
	DATETIME = "2006-01-02 15-04-05"
)

var DATE_RE *regexp.Regexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(\.\w+)?$`)

type TimedRotatingFileHook struct {
	Filename    string
	BackupCount int
	interval    int64
	rotatorAt   int64
	file        *FileHook
	debug       bool
	Locker      sync.Locker
}

func NewTimedRotatingFileHook(filename string) (*TimedRotatingFileHook, error) {
	filename, _ = filepath.Abs(filename)

	file, err := NewFileHook(filename)
	if err != nil {
		return nil, err
	}
	file.Locker = nil

	h := &TimedRotatingFileHook{Filename: filename, interval: MIDNIGHT,
		BackupCount: 31, file: file, Locker: &sync.Mutex{}}
	h.ReComputeRollover(Now())

	t := time.Unix(h.rotatorAt, 0)
	fmt.Println("rotatorAt", t.Format(DATETIME))
	return h, nil
}

func (h *TimedRotatingFileHook) Lock() {
	if h.Locker != nil {
		h.Locker.Lock()
	}
}

func (h *TimedRotatingFileHook) Unlock() {
	if h.Locker != nil {
		h.Locker.Unlock()
	}
}

func (h *TimedRotatingFileHook) ReComputeRollover(current_time int64) {
	t := time.Unix(current_time, 0)
	current_hour := t.Hour()
	current_minute := t.Minute()
	current_second := t.Second()
	r := h.interval - int64((current_hour*60+current_minute)*60+current_second)
	h.rotatorAt = current_time + r
}

func (h *TimedRotatingFileHook) SetMode(mode int) (int, error) {
	return h.file.SetMode(mode)
}

func (h *TimedRotatingFileHook) SetPerm(perm os.FileMode) (os.FileMode, error) {
	return h.file.SetPerm(perm)
}

func (h *TimedRotatingFileHook) SetDebug(debug bool) *TimedRotatingFileHook {
	h.debug = debug
	h.file.SetDebug(debug)
	return h
}

func (h *TimedRotatingFileHook) SetInternal(day int) *TimedRotatingFileHook {
	h.interval = int64(day) * MIDNIGHT
	h.ReComputeRollover(Now())
	return h
}

func (h *TimedRotatingFileHook) SetBackupCount(i int) *TimedRotatingFileHook {
	h.BackupCount = i
	return h
}

func (h *TimedRotatingFileHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func (h *TimedRotatingFileHook) Fire(entry *logrus.Entry) error {
	if !h.file.Ok {
		return errors.New("Writer is not ready to write")
	}

	h.Lock()
	defer h.Unlock()

	if h.shouldRollover(entry) {
		h.doRollover()
	}
	return h.file.Fire(entry)
}

func (h TimedRotatingFileHook) shouldRollover(entry *logrus.Entry) bool {
	if Now() >= h.rotatorAt {
		return true
	}
	return false
}

func (h *TimedRotatingFileHook) doRollover() {
	h.file.Close()

	dstTime := h.rotatorAt - h.interval
	dstPath := h.Filename + "." + time.Unix(dstTime, 0).Format(DATE)
	if IsExist(dstPath) {
		os.Remove(dstPath)
	}

	if IsFile(h.Filename) {
		err := os.Rename(h.Filename, dstPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to rename %v to %v", h.Filename, dstPath)
		}
	}

	if h.BackupCount > 0 {
		files := h.getFilesToDelete()
		for _, file := range files {
			os.Remove(file)
		}
	}

	h.file.Open()
	h.ReComputeRollover(Now())
}

func (h TimedRotatingFileHook) getFilesToDelete() []string {
	result := make([]string, 0, 30)
	dirName, baseName := filepath.Split(h.Filename)
	fileNames, err := List(dirName)
	if err != nil {
		return result
	}

	var suffix, prefix string
	_prefix := baseName + "."
	plen := len(prefix)
	for _, fileName := range fileNames {
		prefix = string(fileName[:plen])
		if _prefix == prefix {
			suffix = string(fileName[plen:])
			if DATE_RE.MatchString(suffix) {
				result = append(result, filepath.Join(dirName, fileName))
			}
		}
	}
	sort.Strings(result)

	if len(result) < h.BackupCount {
		result = []string{}
	} else {
		result = result[:len(result)-h.BackupCount]
	}
	return result
}
