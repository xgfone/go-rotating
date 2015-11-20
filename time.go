package rotating

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

const (
	MIDNIGHT = 60 * 60 * 24
	DATE     = "2006-01-02"
	DATETIME = "2006-01-02 15-04-05"
)

var DATE_RE *regexp.Regexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(\.\w+)?$`)

func Now() int64 {
	return time.Now().Unix()
}

//////////////
// NullWriter
type NullWriter struct {
}

func NewNullWriter() NullWriter {
	return NullWriter{}
}

func (w NullWriter) Write(d []byte) (int, error) {
	return len(d), nil
}

//////////////
// TimeRotatingFileHook
type TimeRotatingFileHook struct {
	Filename    string
	Internal    int
	BackupCount int
	rotatorAt   int64
	file        *FileHook
	debug       bool
}

func NewTimeRotatingFileHook(filename string) (*TimeRotatingFileHook, error) {
	filename, _ = filepath.Abs(filename)
	file, err := NewFileHook(filename)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "Unable to open the file[%v]: %v", filename, err)
		return nil, err
	}

	h := &TimeRotatingFileHook{Filename: filename, Internal: MIDNIGHT, BackupCount: 31, file: file}
	h.ReComputeRollover(Now())
	return h, nil
}

func (h *TimeRotatingFileHook) ReComputeRollover(current_time int64) {
	t := time.Unix(current_time, 0)
	current_hour := t.Hour()
	current_minute := t.Minute()
	current_second := t.Second()
	r := h.Internal - ((current_hour*60+current_minute)*60 + current_second)
	h.rotatorAt = current_time + int64(r)
}

func (h *TimeRotatingFileHook) SetMode(mode int) (int, error) {
	return h.file.SetMode(mode)
}

func (h *TimeRotatingFileHook) SetPerm(perm os.FileMode) (os.FileMode, error) {
	return h.file.SetPerm(perm)
}

func (h *TimeRotatingFileHook) SetDebug(debug bool) *TimeRotatingFileHook {
	h.debug = debug
	h.file.SetDebug(debug)
	return h
}

func (h *TimeRotatingFileHook) SetInternal(i int) *TimeRotatingFileHook {
	h.Internal = i
	return h
}

func (h *TimeRotatingFileHook) SetBackupCount(i int) *TimeRotatingFileHook {
	h.BackupCount = i
	return h
}

func (h *TimeRotatingFileHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func (h *TimeRotatingFileHook) Fire(entry *logrus.Entry) error {
	if !h.file.Ok {
		return errors.New("Writer is not ready to write")
	}

	if h.shouldRollover(entry) {
		h.doRollover()
	}
	h.file.Fire(entry)
	return nil
}

func (h TimeRotatingFileHook) shouldRollover(entry *logrus.Entry) bool {
	if Now() >= h.rotatorAt {
		return true
	}
	return false
}

func (h *TimeRotatingFileHook) doRollover() {
	h.file.Close()
	currentTime := Now()
	dstNow := time.Unix(currentTime, 0)
	dstPath := h.Filename + "." + dstNow.Format(DATE)
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
	h.ReComputeRollover(currentTime)
}

func (h TimeRotatingFileHook) getFilesToDelete() []string {
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
