## [DEPRECATED] Rotate the log file
A hook of [Logrus](https://github.com/Sirupsen/logrus), such as StreamHook, FileHook, TimedRotatingFileHook like TimedRotatingFileHandler of the Python logging.

**Notice:**

This library is **DEPRECATED**. The new one is `github.com/xgfone/go-tools/log/handler`, which implements the interface `io.WriteCloser`, so you can write an adaptor for a certain logging library. For example, the adaptor of [log15](https://github.com/inconshreveable/log15) is `TimedRotatingFileHandler` in `github.com/xgfone/go-utils/log`.

## Example
#### NullWriter
NullWriter 是一个实现了 io.Writer 接口的空写入器，它丢弃要写入的内容，并直接返回。
使用方法：
```
writer := rotating.NewNullWriter()
```
#### StreamHook
```go
package main

import (
    "github.com/Sirupsen/logrus"
    "github.com/xgfone/go-rotating"
    "os"
    "time"
)

var log = logrus.New()

func init() {
    log.Out = rotating.NewNullWriter()
    hook := rotating.NewStreamHook(os.Stderr)
    log.Hooks.Add(hook)
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

#### FileHook
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
    hook, err := rotating.NewFileHook("test.log")
    if err != nil {
        fmt.Println(err)
    } else {
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

#### TimedRotatingFileHook
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
        // Backup once every two hours. There are seven backup files in all, and others will be deleted.
        hook.SetDebug(true).SetBackupCount(7).SetIntervalHour(2) 
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

## Note:
Whichever user is running the go application must have read/write permissions to the log files selected, or if the files do not yet exists, then to the directory in which the files will be created.
