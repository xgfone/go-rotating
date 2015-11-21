package rotating

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"runtime"
	"time"
)

func Now() int64 {
	return time.Now().Unix()
}

func getFileno() string {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("%v:%v", file, line)
	} else {
		return "Unable to get lineno"
	}
}

func Fileno() string {
	return getFileno()
}

func FilenoToField() logrus.Fields {
	return logrus.Fields{
		"lineno": getFileno(),
	}
}
