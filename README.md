#### Rotate the log file
A hook of [Logrus](https://github.com/Sirupsen/logrus), such as StreamHook, FileHook, TimedRotatingFileHook like TimedRotatingFileHandler of the Python logging.

#### Example
```go
package main

import (
    "fmt"
    "github.com/Sirupsen/logrus"
    "github.com/xgfone/go-rotating"
    "time"
)

var log = logrus.New()

func init() {
    log.Out = rotating.NewNullWriter()
    hook, err := rotating.NewTimedRotatingFileHook("test.log")
    if err != nil {
        fmt.Println(err)
    } else {
        hook.SetDebug(true).SetBackupCount(7)
        log.Hooks.Add(hook)
    }
}

func Loop() {
    log.WithFields(logrus.Fields{
        "test1": "test1",
        "test2": "test2",
    }).Info("This is a test.")
}

func main() {
    for true {
        time.Sleep(time.Second)
        Loop()
    }
}
```

#### Note:
Whichever user is running the go application must have read/write permissions to the log files selected, or if the files do not yet exists, then to the directory in which the files will be created.
