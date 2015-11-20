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
	HOUR = 60 * 60
	DAY  = HOUR * 24
	WEEK = DAY * 7

	HOUR_FMT = "2006-01-02_15"
	DAY_FMT  = "2006-01-02"
	DATE_FMT = "2006-01-02 15-04-05"
)

var dayRE, hourRE *regexp.Regexp
var mapExtMatch map[int64]*regexp.Regexp
var mapExtFMT map[int64]string

func init() {
	dayRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(\.\w+)?$`)
	hourRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}(\.\w+)?$`)

	mapExtMatch = map[int64]*regexp.Regexp{
		HOUR: hourRE,
		DAY:  dayRE,
		WEEK: dayRE,
	}

	mapExtFMT = map[int64]string{
		HOUR: HOUR_FMT,
		DAY:  DAY_FMT,
		WEEK: DAY_FMT,
	}
}

type TimedRotatingFileHook struct {
	filename    string
	backupCount int
	interval    int64
	rotatorAt   int64
	when        int64
	extMatch    *regexp.Regexp
	suffix      string
	file        *FileHook
	debug       bool
	locker      sync.Locker
}

func NewTimedRotatingFileHook(filename string) (*TimedRotatingFileHook, error) {
	filename, _ = filepath.Abs(filename)

	file, err := NewFileHook(filename)
	if err != nil {
		return nil, err
	}
	file.SetLock(nil)

	h := &TimedRotatingFileHook{filename: filename, file: file}
	h.SetIntervalDay(1).SetBackupCount(30).SetLock(&sync.Mutex{}).ReComputeRollover()
	return h, nil
}

func (h *TimedRotatingFileHook) Lock() {
	if h.locker != nil {
		h.locker.Lock()
	}
}

func (h *TimedRotatingFileHook) Unlock() {
	if h.locker != nil {
		h.locker.Unlock()
	}
}

func (h *TimedRotatingFileHook) SetLock(locker sync.Locker) *TimedRotatingFileHook {
	h.locker = locker
	return h
}

func (h *TimedRotatingFileHook) ReComputeRollover() *TimedRotatingFileHook {
	current_time := Now()
	t := time.Unix(current_time, 0)
	current_hour := t.Hour()
	current_minute := t.Minute()
	current_second := t.Second()
	var r int64
	if h.when == HOUR {
		r = h.interval - int64(current_minute*60+current_second)
	} else {
		r = h.interval - int64((current_hour*60+current_minute)*60+current_second)
	}
	h.rotatorAt = current_time + r

	if h.debug {
		t := time.Unix(h.rotatorAt, 0).Format(DATE_FMT)
		fmt.Fprintf(os.Stderr, "[DEBUG] The next rotator is at %v\n", t)
	}

	return h
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

func (h *TimedRotatingFileHook) setInterval(per, n int64) *TimedRotatingFileHook {
	h.when = per
	h.interval = per * n
	h.suffix = mapExtFMT[per]
	h.extMatch, _ = mapExtMatch[per]
	h.ReComputeRollover()
	return h
}

func (h *TimedRotatingFileHook) SetIntervalHour(hours int) *TimedRotatingFileHook {
	return h.setInterval(HOUR, int64(hours))
}

func (h *TimedRotatingFileHook) SetIntervalDay(days int) *TimedRotatingFileHook {
	return h.setInterval(DAY, int64(days))
}

func (h *TimedRotatingFileHook) SetIntervalWeek(weeks int) *TimedRotatingFileHook {
	return h.setInterval(WEEK, int64(weeks))
}

func (h *TimedRotatingFileHook) SetBackupCount(i int) *TimedRotatingFileHook {
	h.backupCount = i
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
	dstPath := h.filename + "." + time.Unix(dstTime, 0).Format(h.suffix)
	if IsExist(dstPath) {
		os.Remove(dstPath)
	}

	if IsFile(h.filename) {
		err := os.Rename(h.filename, dstPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to rename %v to %v", h.filename, dstPath)
		}
	}

	if h.backupCount > 0 {
		files := h.getFilesToDelete()
		for _, file := range files {
			os.Remove(file)
		}
	}

	h.file.Open()
	h.ReComputeRollover()
}

func (h TimedRotatingFileHook) getFilesToDelete() []string {
	result := make([]string, 0, 30)
	dirName, baseName := filepath.Split(h.filename)
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
			if h.extMatch.MatchString(suffix) {
				result = append(result, filepath.Join(dirName, fileName))
			}
		}
	}
	sort.Strings(result)

	if len(result) < h.backupCount {
		result = []string{}
	} else {
		result = result[:len(result)-h.backupCount]
	}
	return result
}
